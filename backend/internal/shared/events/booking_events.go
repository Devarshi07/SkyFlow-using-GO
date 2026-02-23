package events

// Queue names
const (
	QueueBookingConfirmed = "booking.confirmed"
	QueuePasswordReset    = "password.reset"
	QueueWelcome          = "user.welcome"
)

// PasswordResetEvent is published when a user requests a password reset
type PasswordResetEvent struct {
	Email     string `json:"email"`
	ResetLink string `json:"reset_link"`
}

// WelcomeEvent is published when a new user registers
type WelcomeEvent struct {
	Email string `json:"email"`
}

// BookingConfirmedEvent is published to RabbitMQ when a booking is confirmed
type BookingConfirmedEvent struct {
	BookingID      string `json:"booking_id"`
	UserID         string `json:"user_id"`
	FlightID       string `json:"flight_id"`
	FlightNumber   string `json:"flight_number"`
	DepartureTime  string `json:"departure_time"`
	ArrivalTime    string `json:"arrival_time"`
	PassengerName  string `json:"passenger_name"`
	PassengerEmail string `json:"passenger_email"`
	Seats          int    `json:"seats"`
	AmountCents    int64  `json:"amount_cents"`
	Status         string `json:"status"`
}
