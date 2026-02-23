package bookings

import (
	"context"
	"fmt"
	"os"

	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
	"github.com/skyflow/skyflow/internal/shared/events"
	"github.com/skyflow/skyflow/internal/flights"
	"github.com/skyflow/skyflow/internal/payments"
)

type Service struct {
	store      Store
	flightSvc  *flights.Service
	paymentSvc *payments.Service
	publisher  *events.Publisher
}

func NewService(store Store, flightSvc *flights.Service, paymentSvc *payments.Service) *Service {
	return &Service{store: store, flightSvc: flightSvc, paymentSvc: paymentSvc}
}

// SetPublisher attaches the event publisher (optional, set after construction)
func (s *Service) SetPublisher(p *events.Publisher) {
	s.publisher = p
}

func (s *Service) Create(ctx context.Context, userID string, req CreateBookingRequest) (*CreateBookingResponse, *apperrors.AppError) {
	if req.FlightID == "" || req.PassengerName == "" || req.PassengerEmail == "" {
		return nil, apperrors.BadRequest("flight_id, passenger_name, passenger_email required")
	}
	if req.Seats <= 0 {
		req.Seats = 1
	}

	f, appErr := s.flightSvc.GetByID(ctx, req.FlightID)
	if appErr != nil {
		return nil, appErr
	}
	if f.SeatsAvailable < req.Seats {
		return nil, apperrors.BadRequest("not enough seats available")
	}

	amount := f.Price * int64(req.Seats)

	// ── Reserve seats immediately ──────────────────────────
	newAvail := f.SeatsAvailable - req.Seats
	if _, updateErr := s.flightSvc.Update(ctx, req.FlightID, flights.UpdateFlightRequest{
		SeatsAvailable: &newAvail,
	}); updateErr != nil {
		return nil, updateErr
	}

	// ── Create payment ─────────────────────────────────────
	frontendOrigin := os.Getenv("FRONTEND_URL")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	var paymentIntentID string
	var checkoutURL string

	if s.paymentSvc.IsLive() {
		// Stripe Checkout Session — redirect user to Stripe's hosted page
		// We create a placeholder booking first to get the ID for URLs
		b := &Booking{
			UserID:         userID,
			FlightID:       req.FlightID,
			Seats:          req.Seats,
			Amount:         amount,
			PassengerName:  req.PassengerName,
			PassengerEmail: req.PassengerEmail,
			PassengerPhone: req.PassengerPhone,
			Status:         "pending",
		}
		created, err := s.store.Create(ctx, b)
		if err != nil {
			// Restore seats on failure
			s.restoreSeats(ctx, req.FlightID, req.Seats)
			return nil, apperrors.Internal(err)
		}

		successURL := frontendOrigin + "/bookings/" + created.ID + "?payment=success&session_id={CHECKOUT_SESSION_ID}"
		cancelURL := frontendOrigin + "/flights?payment=cancelled&booking_id=" + created.ID

		sessResp, payErr := s.paymentSvc.CreateCheckoutSession(
			amount, "usd", created.ID, f.FlightNumber, successURL, cancelURL,
		)
		if payErr != nil {
			s.restoreSeats(ctx, req.FlightID, req.Seats)
			return nil, payErr
		}

		// Update booking with payment intent ID
		_ = s.store.UpdatePaymentIntent(ctx, created.ID, sessResp.PaymentIntentID)

		return &CreateBookingResponse{
			BookingID:       created.ID,
			PaymentIntentID: sessResp.PaymentIntentID,
			CheckoutURL:     sessResp.CheckoutURL,
			Amount:          amount,
			Status:          "pending",
		}, nil
	}

	// ── Mock / demo mode ───────────────────────────────────
	intentResp, payErr := s.paymentSvc.CreateIntent(payments.CreateIntentRequest{
		Amount:   amount,
		Currency: "usd",
	})
	if payErr != nil {
		s.restoreSeats(ctx, req.FlightID, req.Seats)
		return nil, payErr
	}
	paymentIntentID = intentResp.PaymentIntentID

	b := &Booking{
		UserID:          userID,
		FlightID:        req.FlightID,
		Seats:           req.Seats,
		Amount:          amount,
		PassengerName:   req.PassengerName,
		PassengerEmail:  req.PassengerEmail,
		PassengerPhone:  req.PassengerPhone,
		PaymentIntentID: paymentIntentID,
		Status:          "pending",
	}

	created, err := s.store.Create(ctx, b)
	if err != nil {
		s.restoreSeats(ctx, req.FlightID, req.Seats)
		return nil, apperrors.Internal(err)
	}

	return &CreateBookingResponse{
		BookingID:       created.ID,
		PaymentIntentID: paymentIntentID,
		CheckoutURL:     checkoutURL,
		Amount:          amount,
		Status:          "pending",
	}, nil
}

func (s *Service) Confirm(ctx context.Context, paymentIntentID string) (*Booking, *apperrors.AppError) {
	b, err := s.store.GetByPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return nil, apperrors.NotFound("booking")
	}

	if b.Status == "confirmed" {
		// Already confirmed (e.g. Stripe webhook or duplicate call)
		return b, nil
	}

	if s.paymentSvc.IsLive() {
		// Verify payment status with Stripe
		pd, payErr := s.paymentSvc.Get(paymentIntentID)
		if payErr != nil {
			return nil, payErr
		}
		if pd.Status != "succeeded" {
			return nil, apperrors.BadRequest("payment not yet completed (status: " + pd.Status + ")")
		}
	} else {
		// Mock mode — just mark as succeeded
		_, payErr := s.paymentSvc.Confirm(paymentIntentID)
		if payErr != nil {
			return nil, payErr
		}
	}

	if err := s.store.UpdateStatus(ctx, b.ID, "confirmed"); err != nil {
		return nil, apperrors.Internal(err)
	}

	// Seats were already reserved at booking creation time — no change needed
	b.Status = "confirmed"

	// Publish event to RabbitMQ for email notification
	s.publishConfirmation(ctx, b)

	return b, nil
}

// ConfirmBySession verifies payment via Stripe Checkout Session and confirms the booking
func (s *Service) ConfirmBySession(ctx context.Context, bookingID, sessionID string) (*Booking, *apperrors.AppError) {
	b, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, apperrors.NotFound("booking")
	}
	if b.Status == "confirmed" {
		return b, nil
	}

	// Verify with Stripe that payment is complete
	sessStatus, payErr := s.paymentSvc.GetCheckoutSession(sessionID)
	if payErr != nil {
		return nil, payErr
	}
	if sessStatus.PaymentStatus != "paid" {
		return nil, apperrors.BadRequest("payment not yet completed (status: " + sessStatus.PaymentStatus + ")")
	}

	// Save the payment intent ID if we got one
	if sessStatus.PaymentIntentID != "" && b.PaymentIntentID == "" {
		_ = s.store.UpdatePaymentIntent(ctx, b.ID, sessStatus.PaymentIntentID)
	}

	if err := s.store.UpdateStatus(ctx, b.ID, "confirmed"); err != nil {
		return nil, apperrors.Internal(err)
	}
	b.Status = "confirmed"

	// Publish event to RabbitMQ for email notification
	s.publishConfirmation(ctx, b)

	return b, nil
}

// CancelBooking releases reserved seats and marks booking as cancelled
func (s *Service) CancelBooking(ctx context.Context, bookingID string) (*Booking, *apperrors.AppError) {
	b, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, apperrors.NotFound("booking")
	}

	if b.Status == "cancelled" {
		return b, nil
	}

	// Restore seats
	s.restoreSeats(ctx, b.FlightID, b.Seats)

	// Cancel payment if live
	if b.PaymentIntentID != "" && s.paymentSvc.IsLive() {
		s.paymentSvc.Cancel(b.PaymentIntentID)
	}

	if err := s.store.UpdateStatus(ctx, b.ID, "cancelled"); err != nil {
		return nil, apperrors.Internal(err)
	}
	b.Status = "cancelled"
	return b, nil
}

func (s *Service) EditBooking(ctx context.Context, userID, bookingID string, req EditBookingRequest) (*EditBookingResponse, *apperrors.AppError) {
	b, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, apperrors.NotFound("booking")
	}
	if b.UserID != userID {
		return nil, apperrors.Unauthorized("not your booking")
	}
	if b.Status == "cancelled" {
		return nil, apperrors.BadRequest("cannot edit a cancelled booking")
	}

	oldAmount := b.Amount

	// If flight is changing, handle seat swap
	if req.FlightID != "" && req.FlightID != b.FlightID {
		newFlight, appErr := s.flightSvc.GetByID(ctx, req.FlightID)
		if appErr != nil {
			return nil, appErr
		}
		seats := req.Seats
		if seats <= 0 {
			seats = b.Seats
		}
		if newFlight.SeatsAvailable < seats {
			return nil, apperrors.BadRequest("not enough seats on the new flight")
		}

		newAmount := newFlight.Price * int64(seats)
		amountDue := newAmount - oldAmount

		// If the new flight costs more, require additional payment
		if amountDue > 0 {
			// Update passenger details on the booking (but don't change flight yet)
			if req.PassengerName != "" {
				b.PassengerName = req.PassengerName
			}
			if req.PassengerEmail != "" {
				b.PassengerEmail = req.PassengerEmail
			}
			if req.PassengerPhone != "" {
				b.PassengerPhone = req.PassengerPhone
			}
			_ = s.store.UpdateBooking(ctx, b)

			resp, payErr := s.createEditPayment(amountDue, b.ID, req.FlightID, seats)
			if payErr != nil {
				return nil, payErr
			}
			resp.Booking = b
			resp.OldAmount = oldAmount
			resp.NewAmount = newAmount
			return resp, nil
		}

		// Same price or cheaper — just swap flights directly
		s.restoreSeats(ctx, b.FlightID, b.Seats)
		newAvail := newFlight.SeatsAvailable - seats
		s.flightSvc.Update(ctx, req.FlightID, flights.UpdateFlightRequest{SeatsAvailable: &newAvail})

		b.FlightID = req.FlightID
		b.Seats = seats
		b.Amount = newAmount
	} else if req.Seats > 0 && req.Seats != b.Seats {
		// Just changing seat count on same flight
		f, appErr := s.flightSvc.GetByID(ctx, b.FlightID)
		if appErr != nil {
			return nil, appErr
		}
		diff := req.Seats - b.Seats
		if diff > 0 && f.SeatsAvailable < diff {
			return nil, apperrors.BadRequest("not enough seats available")
		}

		newAmount := f.Price * int64(req.Seats)
		amountDue := newAmount - oldAmount

		// If more seats cost more, require payment
		if amountDue > 0 {
			if req.PassengerName != "" {
				b.PassengerName = req.PassengerName
			}
			if req.PassengerEmail != "" {
				b.PassengerEmail = req.PassengerEmail
			}
			if req.PassengerPhone != "" {
				b.PassengerPhone = req.PassengerPhone
			}
			_ = s.store.UpdateBooking(ctx, b)

			resp, payErr := s.createEditPayment(amountDue, b.ID, b.FlightID, req.Seats)
			if payErr != nil {
				return nil, payErr
			}
			resp.Booking = b
			resp.OldAmount = oldAmount
			resp.NewAmount = newAmount
			return resp, nil
		}

		newAvail := f.SeatsAvailable - diff
		s.flightSvc.Update(ctx, b.FlightID, flights.UpdateFlightRequest{SeatsAvailable: &newAvail})
		b.Seats = req.Seats
		b.Amount = newAmount
	}

	// Update passenger details
	if req.PassengerName != "" {
		b.PassengerName = req.PassengerName
	}
	if req.PassengerEmail != "" {
		b.PassengerEmail = req.PassengerEmail
	}
	if req.PassengerPhone != "" {
		b.PassengerPhone = req.PassengerPhone
	}

	if err := s.store.UpdateBooking(ctx, b); err != nil {
		return nil, apperrors.Internal(err)
	}

	// Send updated booking email
	s.publishUpdate(ctx, b)

	return &EditBookingResponse{
		Booking:      b,
		NeedsPayment: false,
	}, nil
}

// createEditPayment creates a Stripe Checkout session or mock intent for edit payment
func (s *Service) createEditPayment(amountDue int64, bookingID, newFlightID string, newSeats int) (*EditBookingResponse, *apperrors.AppError) {
	frontendOrigin := os.Getenv("FRONTEND_URL")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	if s.paymentSvc.IsLive() {
		successURL := fmt.Sprintf("%s/bookings/%s?edit_paid=true&session_id={CHECKOUT_SESSION_ID}&new_flight_id=%s&new_seats=%d",
			frontendOrigin, bookingID, newFlightID, newSeats)
		cancelURL := fmt.Sprintf("%s/bookings/%s?edit_cancelled=true", frontendOrigin, bookingID)

		sessResp, payErr := s.paymentSvc.CreateCheckoutSession(
			amountDue, "usd", bookingID, "Booking upgrade", successURL, cancelURL,
		)
		if payErr != nil {
			return nil, payErr
		}

		return &EditBookingResponse{
			NeedsPayment:    true,
			PaymentIntentID: sessResp.PaymentIntentID,
			CheckoutURL:     sessResp.CheckoutURL,
			AmountDue:       amountDue,
		}, nil
	}

	// Demo mode
	intentResp, payErr := s.paymentSvc.CreateIntent(payments.CreateIntentRequest{
		Amount:   amountDue,
		Currency: "usd",
	})
	if payErr != nil {
		return nil, payErr
	}

	return &EditBookingResponse{
		NeedsPayment:    true,
		PaymentIntentID: intentResp.PaymentIntentID,
		AmountDue:       amountDue,
	}, nil
}

// ConfirmEdit finalizes a booking edit after additional payment
func (s *Service) ConfirmEdit(ctx context.Context, userID, bookingID string, paymentIntentID string, sessionID string, newFlightID string, newSeats int) (*Booking, *apperrors.AppError) {
	b, err := s.store.GetByID(ctx, bookingID)
	if err != nil {
		return nil, apperrors.NotFound("booking")
	}
	if b.UserID != userID {
		return nil, apperrors.Unauthorized("not your booking")
	}

	// Confirm the payment
	if sessionID != "" && s.paymentSvc.IsLive() {
		// Stripe Checkout session flow
		sessStatus, payErr := s.paymentSvc.GetCheckoutSession(sessionID)
		if payErr != nil {
			return nil, payErr
		}
		if sessStatus.PaymentStatus != "paid" {
			return nil, apperrors.BadRequest("payment not yet completed (status: " + sessStatus.PaymentStatus + ")")
		}
	} else if s.paymentSvc.IsLive() {
		pd, payErr := s.paymentSvc.Get(paymentIntentID)
		if payErr != nil {
			return nil, payErr
		}
		if pd.Status != "succeeded" {
			return nil, apperrors.BadRequest("payment not yet completed")
		}
	} else {
		_, payErr := s.paymentSvc.Confirm(paymentIntentID)
		if payErr != nil {
			return nil, payErr
		}
	}

	// Now apply the flight/seat change
	seats := newSeats
	if seats <= 0 {
		seats = b.Seats
	}

	if newFlightID != "" && newFlightID != b.FlightID {
		newFlight, appErr := s.flightSvc.GetByID(ctx, newFlightID)
		if appErr != nil {
			return nil, appErr
		}
		if newFlight.SeatsAvailable < seats {
			return nil, apperrors.BadRequest("not enough seats on the new flight")
		}

		s.restoreSeats(ctx, b.FlightID, b.Seats)
		newAvail := newFlight.SeatsAvailable - seats
		s.flightSvc.Update(ctx, newFlightID, flights.UpdateFlightRequest{SeatsAvailable: &newAvail})

		b.FlightID = newFlightID
		b.Seats = seats
		b.Amount = newFlight.Price * int64(seats)
	} else if seats != b.Seats {
		f, appErr := s.flightSvc.GetByID(ctx, b.FlightID)
		if appErr != nil {
			return nil, appErr
		}
		diff := seats - b.Seats
		newAvail := f.SeatsAvailable - diff
		s.flightSvc.Update(ctx, b.FlightID, flights.UpdateFlightRequest{SeatsAvailable: &newAvail})
		b.Seats = seats
		b.Amount = f.Price * int64(seats)
	}

	if err := s.store.UpdateBooking(ctx, b); err != nil {
		return nil, apperrors.Internal(err)
	}

	// Send updated booking email
	s.publishUpdate(ctx, b)

	return b, nil
}

// publishUpdate publishes a booking updated event for email notification
func (s *Service) publishUpdate(ctx context.Context, b *Booking) {
	if s.publisher == nil {
		return
	}
	flightNumber := ""
	departureTime := ""
	arrivalTime := ""
	if f, err := s.flightSvc.GetByID(ctx, b.FlightID); err == nil {
		flightNumber = f.FlightNumber
		departureTime = f.DepartureTime.Format("Monday, January 2, 2006 at 3:04 PM")
		arrivalTime = f.ArrivalTime.Format("Monday, January 2, 2006 at 3:04 PM")
	}
	s.publisher.PublishBookingConfirmed(ctx, events.BookingConfirmedEvent{
		BookingID:      b.ID,
		UserID:         b.UserID,
		FlightID:       b.FlightID,
		FlightNumber:   flightNumber,
		DepartureTime:  departureTime,
		ArrivalTime:    arrivalTime,
		PassengerName:  b.PassengerName,
		PassengerEmail: b.PassengerEmail,
		Seats:          b.Seats,
		AmountCents:    b.Amount,
		Status:         "updated",
	})
}

func (s *Service) GetByID(ctx context.Context, id string) (*Booking, *apperrors.AppError) {
	b, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("booking")
	}
	return b, nil
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]*Booking, *apperrors.AppError) {
	list, err := s.store.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if list == nil {
		list = []*Booking{}
	}
	return list, nil
}

// publishConfirmation publishes a booking confirmed event to RabbitMQ
func (s *Service) publishConfirmation(ctx context.Context, b *Booking) {
	if s.publisher == nil {
		return
	}
	// Get flight details for the email
	flightNumber := ""
	departureTime := ""
	arrivalTime := ""
	if f, err := s.flightSvc.GetByID(ctx, b.FlightID); err == nil {
		flightNumber = f.FlightNumber
		departureTime = f.DepartureTime.Format("Monday, January 2, 2006 at 3:04 PM")
		arrivalTime = f.ArrivalTime.Format("Monday, January 2, 2006 at 3:04 PM")
	}
	s.publisher.PublishBookingConfirmed(ctx, events.BookingConfirmedEvent{
		BookingID:      b.ID,
		UserID:         b.UserID,
		FlightID:       b.FlightID,
		FlightNumber:   flightNumber,
		DepartureTime:  departureTime,
		ArrivalTime:    arrivalTime,
		PassengerName:  b.PassengerName,
		PassengerEmail: b.PassengerEmail,
		Seats:          b.Seats,
		AmountCents:    b.Amount,
		Status:         "confirmed",
	})
}

// restoreSeats adds back reserved seats to the flight
func (s *Service) restoreSeats(ctx context.Context, flightID string, seats int) {
	f, err := s.flightSvc.GetByID(ctx, flightID)
	if err != nil {
		return
	}
	restored := f.SeatsAvailable + seats
	s.flightSvc.Update(ctx, flightID, flights.UpdateFlightRequest{SeatsAvailable: &restored})
}
