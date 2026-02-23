package flights

import "time"

type Flight struct {
	ID             string    `json:"id"`
	FlightNumber   string    `json:"flight_number"`
	OriginID       string    `json:"origin_id"`
	DestinationID  string    `json:"destination_id"`
	DepartureTime  time.Time `json:"departure_time"`
	ArrivalTime    time.Time `json:"arrival_time"`
	Price          int64     `json:"price"`
	SeatsTotal     int       `json:"seats_total"`
	SeatsAvailable int       `json:"seats_available"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateFlightRequest struct {
	FlightNumber  string    `json:"flight_number"`
	OriginID      string    `json:"origin_id"`
	DestinationID string    `json:"destination_id"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	Price         int64     `json:"price"`
	SeatsTotal    int       `json:"seats_total"`
}

type UpdateFlightRequest struct {
	FlightNumber   *string    `json:"flight_number,omitempty"`
	OriginID       *string    `json:"origin_id,omitempty"`
	DestinationID  *string    `json:"destination_id,omitempty"`
	DepartureTime  *time.Time `json:"departure_time,omitempty"`
	ArrivalTime    *time.Time `json:"arrival_time,omitempty"`
	Price          *int64     `json:"price,omitempty"`
	SeatsTotal     *int       `json:"seats_total,omitempty"`
	SeatsAvailable *int       `json:"seats_available,omitempty"`
}

type SearchParams struct {
	OriginID      string
	DestinationID string
	Date          string
}

type ConnectingFlight struct {
	Leg1              *Flight `json:"leg1"`
	Leg2              *Flight `json:"leg2"`
	TotalPrice        int64   `json:"total_price"`
	Discount          int     `json:"discount"`
	TotalDurationHrs  float64 `json:"total_duration_hours"`
	LayoverMinutes    int     `json:"layover_minutes"`
	LayoverAirportID  string  `json:"layover_airport_id"`
}

type SearchResult struct {
	Flights    []*Flight           `json:"flights"`
	Connecting []*ConnectingFlight `json:"connecting"`
}
