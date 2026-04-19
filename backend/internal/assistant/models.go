package assistant

// ChatRequest is the incoming request for a chat message
type ChatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// ChatResponse is the response from a chat message
type ChatResponse struct {
	Response       string `json:"response"`
	ConversationID string `json:"conversation_id"`
	ToolCalls      []struct {
		Tool   string `json:"tool"`
		Params any    `json:"params"`
		Result string `json:"result"`
	} `json:"tool_calls,omitempty"`
}

// Assistant message for frontend
type AssistantMessage struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	IsLoading bool   `json:"is_loading,omitempty"`
}

// Tool execution results for search
type SearchFlightsResult struct {
	Flights []struct {
		ID             string `json:"id"`
		FlightNumber   string `json:"flight_number"`
		OriginID       string `json:"origin_id"`
		DestinationID  string `json:"destination_id"`
		DepartureTime  string `json:"departure_time"`
		ArrivalTime    string `json:"arrival_time"`
		Price          int64  `json:"price"`
		SeatsAvailable int    `json:"seats_available"`
	} `json:"flights"`
}

type GetBookingResult struct {
	ID              string `json:"id"`
	FlightID        string `json:"flight_id"`
	Seats           int    `json:"seats"`
	Amount          int64  `json:"amount"`
	PassengerName   string `json:"passenger_name"`
	PassengerEmail  string `json:"passenger_email"`
	PassengerPhone  string `json:"passenger_phone"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
}

type GetAirportsResult struct {
	Airports []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Code   string `json:"code"`
		CityID string `json:"city_id"`
	} `json:"airports"`
}

type GetCitiesResult struct {
	Cities []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Country string `json:"country"`
		Code    string `json:"code"`
	} `json:"cities"`
}

type CreateBookingResult struct {
	BookingID       string `json:"booking_id"`
	PaymentIntentID string `json:"payment_intent_id,omitempty"`
	CheckoutURL     string `json:"checkout_url,omitempty"`
	Amount          int64  `json:"amount"`
	Status          string `json:"status"`
	Message         string `json:"message,omitempty"`
}

type MyBookingsResult struct {
	Bookings []GetBookingResult `json:"bookings"`
}

type CancelBookingResult struct {
	BookingID string `json:"booking_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type SendEmailResult struct {
	Message string `json:"message"`
}