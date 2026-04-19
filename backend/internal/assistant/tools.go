package assistant

import (
	"encoding/json"

	"github.com/skyflow/skyflow/internal/shared/logger"
)

const (
	ToolSearchFlights          = "search_flights"
	ToolGetBooking             = "get_booking"
	ToolGetMyBookings          = "get_my_bookings"
	ToolCreateBooking          = "create_booking"
	ToolCancelBooking          = "cancel_booking"
	ToolSendConfirmationEmail = "send_confirmation_email"
	ToolGetAirports            = "get_airports"
	ToolGetCities              = "get_cities"
)

// SystemPrompt is intentionally terse to keep per-turn tokens small — small
// models hit rate limits fast. The service appends date + auth status each turn.
const SystemPrompt = `You are SkyFlow Assistant, a flight booking helper.

Rules:
- Keep replies short and friendly, in plain English. Never show raw JSON.
- Use tools via the function-calling API. Do NOT write <function=...> in text.
- Never invent flight IDs, booking IDs, or prices — only use values from tool results.
- If something is missing (origin/destination/date), ask one short question.

Booking flow:
1. search_flights(origin, destination, date) — origin/destination can be a city name or airport code (e.g. "Mumbai" or "BOM"). Date is YYYY-MM-DD.
2. List results with flight number, time, price ($USD), seats. Ask which to book.
3. Collect passenger_name + passenger_email (phone optional). Confirm back to user.
4. Only after the user says yes, call create_booking.
5. Relay booking_id and amount; mention the checkout_url if present.

Other: get_my_bookings, get_booking, cancel_booking, send_confirmation_email.
If not signed in, tell the user to sign in before booking or viewing bookings; search works without sign-in.`

// GetTools returns the tool definitions for Groq. Descriptions are kept short
// to reduce per-request token cost.
func GetTools(log *logger.Logger) []groqTool {
	return []groqTool{
		{Type: "function", Function: groqToolFunc(
			ToolSearchFlights,
			"Search flights between two places on a date.",
			`{"type":"object","properties":{"origin":{"type":"string","description":"City or airport (name or code)"},"destination":{"type":"string","description":"City or airport (name or code)"},"date":{"type":"string","description":"YYYY-MM-DD"}},"required":["origin","destination","date"]}`,
		)},
		{Type: "function", Function: groqToolFunc(
			ToolGetMyBookings,
			"List the signed-in user's bookings.",
			`{"type":"object","properties":{}}`,
		)},
		{Type: "function", Function: groqToolFunc(
			ToolGetBooking,
			"Get one booking by id.",
			`{"type":"object","properties":{"booking_id":{"type":"string"}},"required":["booking_id"]}`,
		)},
		{Type: "function", Function: groqToolFunc(
			ToolCreateBooking,
			"Book a flight. Only call after user confirms details.",
			`{"type":"object","properties":{"flight_id":{"type":"string"},"seats":{"type":"integer"},"passenger_name":{"type":"string"},"passenger_email":{"type":"string"},"passenger_phone":{"type":"string"}},"required":["flight_id","passenger_name","passenger_email"]}`,
		)},
		{Type: "function", Function: groqToolFunc(
			ToolCancelBooking,
			"Cancel a booking by id.",
			`{"type":"object","properties":{"booking_id":{"type":"string"}},"required":["booking_id"]}`,
		)},
		{Type: "function", Function: groqToolFunc(
			ToolSendConfirmationEmail,
			"Resend booking confirmation email.",
			`{"type":"object","properties":{"booking_id":{"type":"string"}},"required":["booking_id"]}`,
		)},
	}
}

func groqToolFunc(name, description, parameters string) struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
} {
	var params any
	_ = json.Unmarshal([]byte(parameters), &params)
	return struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  any    `json:"parameters"`
	}{
		Name:        name,
		Description: description,
		Parameters:  params,
	}
}