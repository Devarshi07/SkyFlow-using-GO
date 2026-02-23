package events

// Queue names
const (
	QueueBookingConfirmed = "booking.confirmed"
)

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
