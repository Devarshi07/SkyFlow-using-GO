package payments

import (
	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
)

var defaultMethods = []PaymentMethod{
	{ID: "card", Name: "Card", Type: "card"},
	{ID: "apple_pay", Name: "Apple Pay", Type: "card"},
	{ID: "google_pay", Name: "Google Pay", Type: "card"},
}

type Service struct {
	stripe *StripeClient
}

func NewService(stripe *StripeClient) *Service {
	return &Service{stripe: stripe}
}

func (s *Service) IsLive() bool {
	return s.stripe.IsLive()
}

func (s *Service) GetCheckoutSession(sessionID string) (*CheckoutSessionStatus, *apperrors.AppError) {
	return s.stripe.GetCheckoutSession(sessionID)
}

func (s *Service) CreateCheckoutSession(amount int64, currency, bookingID, flightNumber, successURL, cancelURL string) (*CheckoutSessionResponse, *apperrors.AppError) {
	return s.stripe.CreateCheckoutSession(amount, currency, bookingID, flightNumber, successURL, cancelURL)
}

func (s *Service) CreateIntent(req CreateIntentRequest) (*IntentResponse, *apperrors.AppError) {
	return s.stripe.CreateIntent(req.Amount, req.Currency, req.CustomerID)
}

func (s *Service) Confirm(paymentIntentID string) (*PaymentDetails, *apperrors.AppError) {
	return s.stripe.Confirm(paymentIntentID)
}

func (s *Service) Get(paymentIntentID string) (*PaymentDetails, *apperrors.AppError) {
	return s.stripe.Get(paymentIntentID)
}

func (s *Service) Refund(paymentIntentID string, amount *int64) (*PaymentDetails, *apperrors.AppError) {
	return s.stripe.Refund(paymentIntentID, amount)
}

func (s *Service) Cancel(paymentIntentID string) (*PaymentDetails, *apperrors.AppError) {
	return s.stripe.Cancel(paymentIntentID)
}

func (s *Service) GetMethods() []PaymentMethod {
	return defaultMethods
}
