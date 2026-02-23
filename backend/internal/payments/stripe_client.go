package payments

import (
	"os"

	"github.com/google/uuid"
	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/refund"
)

// StripeClient wraps Stripe API calls; uses mock when key is not set
type StripeClient struct {
	live   bool
	mockDB map[string]*mockIntent
}

type mockIntent struct {
	ID       string
	Secret   string
	Amount   int64
	Currency string
	Status   string
}

func NewStripeClient() *StripeClient {
	key := os.Getenv("STRIPE_SECRET_KEY")
	sc := &StripeClient{mockDB: make(map[string]*mockIntent)}
	if key != "" {
		stripe.Key = key
		sc.live = true
	}
	return sc
}

func (c *StripeClient) IsLive() bool {
	return c.live
}

// GetCheckoutSession retrieves a Checkout Session and returns payment status
func (c *StripeClient) GetCheckoutSession(sessionID string) (*CheckoutSessionStatus, *apperrors.AppError) {
	if !c.live {
		return &CheckoutSessionStatus{Status: "complete", PaymentStatus: "paid"}, nil
	}
	params := &stripe.CheckoutSessionParams{}
	sess, err := session.Get(sessionID, params)
	if err != nil {
		return nil, apperrors.NotFound("checkout session")
	}
	piID := ""
	if sess.PaymentIntent != nil {
		piID = sess.PaymentIntent.ID
	}
	return &CheckoutSessionStatus{
		Status:          string(sess.Status),
		PaymentStatus:   string(sess.PaymentStatus),
		PaymentIntentID: piID,
	}, nil
}

// CreateCheckoutSession creates a Stripe Checkout Session for hosted payment.
// Returns the session URL for redirect and the session ID for later confirmation.
func (c *StripeClient) CreateCheckoutSession(amount int64, currency, bookingID, flightNumber, successURL, cancelURL string) (*CheckoutSessionResponse, *apperrors.AppError) {
	if currency == "" {
		currency = "usd"
	}
	if amount <= 0 {
		return nil, apperrors.BadRequest("amount must be positive")
	}

	if c.live {
		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency:   stripe.String(currency),
						UnitAmount: stripe.Int64(amount),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String("Flight " + flightNumber),
							Description: stripe.String("SkyFlow Booking: " + bookingID),
						},
					},
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL:         stripe.String(successURL),
			CancelURL:          stripe.String(cancelURL),
			PaymentIntentData:  &stripe.CheckoutSessionPaymentIntentDataParams{
				Metadata: map[string]string{
					"booking_id": bookingID,
				},
			},
		}
		params.AddMetadata("booking_id", bookingID)

		sess, err := session.New(params)
		if err != nil {
			return nil, apperrors.NewWithDetails(apperrors.CodePaymentFailed, "Failed to create checkout session", 402, map[string]string{"cause": err.Error()}, err)
		}
		piID := ""
		if sess.PaymentIntent != nil {
			piID = sess.PaymentIntent.ID
		}
		return &CheckoutSessionResponse{
			SessionID:       sess.ID,
			CheckoutURL:     sess.URL,
			PaymentIntentID: piID,
			Amount:          amount,
			Currency:        currency,
		}, nil
	}

	// Demo/mock mode — generate fake session
	sessionID := "cs_mock_" + uuid.New().String()[:8]
	piID := "pi_mock_" + uuid.New().String()[:8]
	c.mockDB[piID] = &mockIntent{ID: piID, Amount: amount, Currency: currency, Status: "requires_payment_method"}
	return &CheckoutSessionResponse{
		SessionID:       sessionID,
		CheckoutURL:     "", // empty = frontend uses built-in checkout
		PaymentIntentID: piID,
		Amount:          amount,
		Currency:        currency,
	}, nil
}

func (c *StripeClient) CreateIntent(amount int64, currency, customerID string) (*IntentResponse, *apperrors.AppError) {
	if currency == "" {
		currency = "usd"
	}
	if amount <= 0 {
		return nil, apperrors.BadRequest("amount must be positive")
	}

	if c.live {
		params := &stripe.PaymentIntentParams{
			Amount:   stripe.Int64(amount),
			Currency: stripe.String(currency),
		}
		if customerID != "" {
			params.Customer = stripe.String(customerID)
		}
		pi, err := paymentintent.New(params)
		if err != nil {
			return nil, apperrors.NewWithDetails(apperrors.CodePaymentFailed, err.Error(), 402, nil, err)
		}
		return &IntentResponse{
			ClientSecret:    pi.ClientSecret,
			PaymentIntentID: pi.ID,
			Amount:          pi.Amount,
			Currency:        string(pi.Currency),
		}, nil
	}

	// Demo/mock mode
	id := "pi_mock_" + uuid.New().String()[:8]
	secret := "pi_mock_" + uuid.New().String()[:8] + "_secret_" + uuid.New().String()[:16]
	c.mockDB[id] = &mockIntent{ID: id, Secret: secret, Amount: amount, Currency: currency, Status: "requires_payment_method"}
	return &IntentResponse{
		ClientSecret:    secret,
		PaymentIntentID: id,
		Amount:          amount,
		Currency:        currency,
	}, nil
}

func (c *StripeClient) Confirm(paymentIntentID string) (*PaymentDetails, *apperrors.AppError) {
	if c.live {
		pi, err := paymentintent.Confirm(paymentIntentID, nil)
		if err != nil {
			return nil, apperrors.NewWithDetails(apperrors.CodePaymentFailed, err.Error(), 402, nil, err)
		}
		return &PaymentDetails{ID: pi.ID, Amount: pi.Amount, Currency: string(pi.Currency), Status: string(pi.Status)}, nil
	}
	m, ok := c.mockDB[paymentIntentID]
	if !ok {
		return nil, apperrors.NotFound("payment intent")
	}
	m.Status = "succeeded"
	return &PaymentDetails{ID: m.ID, Amount: m.Amount, Currency: m.Currency, Status: m.Status}, nil
}

func (c *StripeClient) Get(paymentIntentID string) (*PaymentDetails, *apperrors.AppError) {
	if c.live {
		pi, err := paymentintent.Get(paymentIntentID, nil)
		if err != nil {
			return nil, apperrors.NotFound("payment intent")
		}
		return &PaymentDetails{ID: pi.ID, Amount: pi.Amount, Currency: string(pi.Currency), Status: string(pi.Status)}, nil
	}
	m, ok := c.mockDB[paymentIntentID]
	if !ok {
		return nil, apperrors.NotFound("payment intent")
	}
	return &PaymentDetails{ID: m.ID, Amount: m.Amount, Currency: m.Currency, Status: m.Status}, nil
}

func (c *StripeClient) Refund(paymentIntentID string, amount *int64) (*PaymentDetails, *apperrors.AppError) {
	if c.live {
		params := &stripe.RefundParams{PaymentIntent: stripe.String(paymentIntentID)}
		if amount != nil && *amount > 0 {
			params.Amount = stripe.Int64(*amount)
		}
		ref, err := refund.New(params)
		if err != nil {
			return nil, apperrors.NewWithDetails(apperrors.CodePaymentFailed, err.Error(), 402, nil, err)
		}
		return &PaymentDetails{ID: ref.ID, Status: string(ref.Status)}, nil
	}
	m, ok := c.mockDB[paymentIntentID]
	if !ok {
		return nil, apperrors.NotFound("payment intent")
	}
	m.Status = "refunded"
	return &PaymentDetails{ID: m.ID, Amount: m.Amount, Currency: m.Currency, Status: "refunded"}, nil
}

func (c *StripeClient) Cancel(paymentIntentID string) (*PaymentDetails, *apperrors.AppError) {
	if c.live {
		pi, err := paymentintent.Cancel(paymentIntentID, nil)
		if err != nil {
			return nil, apperrors.NewWithDetails(apperrors.CodePaymentCanceled, err.Error(), 400, nil, err)
		}
		return &PaymentDetails{ID: pi.ID, Status: string(pi.Status)}, nil
	}
	m, ok := c.mockDB[paymentIntentID]
	if !ok {
		return nil, apperrors.NotFound("payment intent")
	}
	m.Status = "canceled"
	return &PaymentDetails{ID: m.ID, Amount: m.Amount, Currency: m.Currency, Status: "canceled"}, nil
}
