package bookings

import "time"

type Booking struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	FlightID        string    `json:"flight_id"`
	Seats           int       `json:"seats"`
	Amount          int64     `json:"amount"`
	PassengerName   string    `json:"passenger_name"`
	PassengerEmail  string    `json:"passenger_email"`
	PassengerPhone  string    `json:"passenger_phone,omitempty"`
	PaymentIntentID string    `json:"payment_intent_id,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateBookingRequest struct {
	FlightID       string `json:"flight_id"`
	Seats          int    `json:"seats"`
	PassengerName  string `json:"passenger_name"`
	PassengerEmail string `json:"passenger_email"`
	PassengerPhone string `json:"passenger_phone"`
}

type CreateBookingResponse struct {
	BookingID       string `json:"booking_id"`
	PaymentIntentID string `json:"payment_intent_id,omitempty"`
	CheckoutURL     string `json:"checkout_url,omitempty"`
	Amount          int64  `json:"amount"`
	Status          string `json:"status"`
}

type ConfirmRequest struct {
	PaymentIntentID string `json:"payment_intent_id"`
}

type RoundTripRequest struct {
	OutboundFlightID string `json:"outbound_flight_id"`
	ReturnFlightID   string `json:"return_flight_id"`
	Seats            int    `json:"seats"`
	PassengerName    string `json:"passenger_name"`
	PassengerEmail   string `json:"passenger_email"`
	PassengerPhone   string `json:"passenger_phone"`
}

type RoundTripResponse struct {
	OutboundBookingID string `json:"outbound_booking_id"`
	ReturnBookingID   string `json:"return_booking_id"`
	PaymentIntentID   string `json:"payment_intent_id,omitempty"`
	CheckoutURL       string `json:"checkout_url,omitempty"`
	TotalAmount       int64  `json:"total_amount"`
	Status            string `json:"status"`
}

type EditBookingRequest struct {
	FlightID       string `json:"flight_id,omitempty"`
	Seats          int    `json:"seats,omitempty"`
	PassengerName  string `json:"passenger_name,omitempty"`
	PassengerEmail string `json:"passenger_email,omitempty"`
	PassengerPhone string `json:"passenger_phone,omitempty"`
}

type EditBookingResponse struct {
	Booking         *Booking `json:"booking"`
	NeedsPayment    bool     `json:"needs_payment"`
	PaymentIntentID string   `json:"payment_intent_id,omitempty"`
	CheckoutURL     string   `json:"checkout_url,omitempty"`
	AmountDue       int64    `json:"amount_due,omitempty"`
	OldAmount       int64    `json:"old_amount,omitempty"`
	NewAmount       int64    `json:"new_amount,omitempty"`
}
