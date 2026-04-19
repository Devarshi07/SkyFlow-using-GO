package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skyflow/skyflow/internal/auth"
	"github.com/skyflow/skyflow/internal/bookings"
	"github.com/skyflow/skyflow/internal/cities"
	"github.com/skyflow/skyflow/internal/flights"
	"github.com/skyflow/skyflow/internal/airports"
	"github.com/skyflow/skyflow/internal/shared/events"
	"github.com/skyflow/skyflow/internal/shared/logger"
)

const (
	conversationTTL = 24 * time.Hour
	convKeyPrefix   = "assistant:conversation:"
)

type Service struct {
	groq       *GroqClient
	flightSvc  *flights.Service
	bookingSvc *bookings.Service
	airportSvc *airports.Service
	citySvc    *cities.Service
	publisher  *events.Publisher
	rdb        *redis.Client
	log        *logger.Logger
}

func NewService(
	flightSvc *flights.Service,
	bookingSvc *bookings.Service,
	airportSvc *airports.Service,
	citySvc *cities.Service,
	publisher *events.Publisher,
	rdb *redis.Client,
	log *logger.Logger,
) *Service {
	return &Service{
		groq:       NewGroqClient(),
		flightSvc:  flightSvc,
		bookingSvc: bookingSvc,
		airportSvc: airportSvc,
		citySvc:    citySvc,
		publisher:  publisher,
		rdb:        rdb,
		log:        log,
	}
}

// Chat handles a user message and returns the assistant's response
func (s *Service) Chat(ctx context.Context, userID string, req *ChatRequest) (*ChatResponse, error) {
	// Get or create conversation
	var conv *Conversation
	if req.ConversationID != "" {
		conv = s.loadConversation(ctx, req.ConversationID)
	}
	if conv == nil {
		conv = &Conversation{
			ID:        uuid.New().String(),
			UserID:    userID,
			Messages:  []Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Add user message
	userMsg := Message{
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, userMsg)

	// Build messages for Groq
	groqMsgs := s.buildGroqMessages(conv)

	// Get tool definitions
	tools := GetTools(s.log)

	// Call Groq
	resp, err := s.groq.Chat(ctx, groqMsgs, tools)
	if err != nil {
		if strings.Contains(err.Error(), "rate_limit_exceeded") || strings.Contains(err.Error(), "status 429") {
			s.log.Warn("groq rate limited", "error", err)
			return &ChatResponse{
				Response:       "I'm being rate-limited by the model provider right now. Please try again in a few seconds.",
				ConversationID: conv.ID,
			}, nil
		}
		if strings.Contains(err.Error(), "tool_use_failed") {
			s.log.Warn("groq tool_use_failed, returning fallback", "error", err)
			return &ChatResponse{
				Response:       "I had trouble formulating that request. Could you rephrase with a bit more detail — e.g. origin, destination, and date?",
				ConversationID: conv.ID,
			}, nil
		}
		return nil, fmt.Errorf("Groq API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return &ChatResponse{
			Response:       "Sorry, I couldn't generate a response. Please try again.",
			ConversationID: conv.ID,
		}, nil
	}

	groqMsg := resp.Choices[0].Message

	// Handle tool calls (loop to support multi-step tool use)
	const maxToolRounds = 5
	for round := 0; round < maxToolRounds && len(groqMsg.ToolCalls) > 0; round++ {
		// Persist assistant message that requested the tool calls
		storedToolCalls := make([]ToolCall, 0, len(groqMsg.ToolCalls))
		for _, tc := range groqMsg.ToolCalls {
			storedToolCalls = append(storedToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Args:      tc.Function.Arguments,
				Timestamp: time.Now(),
			})
		}
		conv.Messages = append(conv.Messages, Message{
			Role:      "assistant",
			Content:   groqMsg.Content,
			ToolCalls: storedToolCalls,
			Timestamp: time.Now(),
		})

		// Execute each tool and append its result
		for _, tc := range groqMsg.ToolCalls {
			result, execErr := s.executeTool(ctx, userID, tc.Function.Name, tc.Function.Arguments)
			if execErr != nil {
				result = fmt.Sprintf(`{"error":%q}`, execErr.Error())
			}
			conv.Messages = append(conv.Messages, Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
				Timestamp:  time.Now(),
			})
		}

		// Re-call Groq with tool results
		groqMsgs = s.buildGroqMessages(conv)
		resp, err = s.groq.Chat(ctx, groqMsgs, tools)
		if err != nil {
			if strings.Contains(err.Error(), "rate_limit_exceeded") || strings.Contains(err.Error(), "status 429") {
				s.log.Warn("groq rate limited in follow-up", "error", err)
				return &ChatResponse{
					Response:       "I'm being rate-limited by the model provider right now. Please try again in a few seconds.",
					ConversationID: conv.ID,
				}, nil
			}
			if strings.Contains(err.Error(), "tool_use_failed") {
				s.log.Warn("groq tool_use_failed in follow-up, ending turn", "error", err)
				groqMsg.Content = "I got back some information but had trouble processing it. Could you ask again in a different way?"
				groqMsg.ToolCalls = nil
				break
			}
			return nil, fmt.Errorf("Groq API call failed: %w", err)
		}
		if len(resp.Choices) == 0 {
			break
		}
		groqMsg = resp.Choices[0].Message
	}

	// Add final assistant message
	assistantMsg := Message{
		Role:      "assistant",
		Content:   groqMsg.Content,
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, assistantMsg)
	conv.UpdatedAt = time.Now()

	// Save conversation
	s.saveConversation(ctx, conv)

	return &ChatResponse{
		Response:       groqMsg.Content,
		ConversationID: conv.ID,
	}, nil
}

// GetConversation retrieves a conversation by ID
func (s *Service) GetConversation(ctx context.Context, convID, userID string) (*Conversation, error) {
	conv := s.loadConversation(ctx, convID)
	if conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}
	if conv.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	return conv, nil
}

// executeTool runs a tool and returns its result as a string
func (s *Service) executeTool(ctx context.Context, userID, toolName, argsJSON string) (string, error) {
	needsBookings := map[string]bool{
		ToolGetBooking:            true,
		ToolGetMyBookings:         true,
		ToolCreateBooking:         true,
		ToolCancelBooking:         true,
		ToolSendConfirmationEmail: true,
	}
	if needsBookings[toolName] && s.bookingSvc == nil {
		return toolErr("Booking service is not configured on this server."), nil
	}

	switch toolName {
	case ToolSearchFlights:
		return s.searchFlights(ctx, argsJSON)
	case ToolGetBooking:
		return s.getBooking(ctx, userID, argsJSON)
	case ToolGetMyBookings:
		return s.getMyBookings(ctx, userID)
	case ToolCreateBooking:
		return s.createBooking(ctx, userID, argsJSON)
	case ToolCancelBooking:
		return s.cancelBooking(ctx, userID, argsJSON)
	case ToolSendConfirmationEmail:
		return s.sendConfirmationEmail(ctx, userID, argsJSON)
	case ToolGetAirports:
		return s.getAirports(ctx)
	case ToolGetCities:
		return s.getCities(ctx)
	default:
		return toolErr(fmt.Sprintf("Unknown tool: %s", toolName)), nil
	}
}

func (s *Service) searchFlights(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
		Date        string `json:"date"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	originID := s.resolveLocation(ctx, args.Origin)
	destID := s.resolveLocation(ctx, args.Destination)
	if originID == "" {
		return toolErr(fmt.Sprintf("Could not recognize origin %q. Ask the user for a known city or airport (e.g. 'Mumbai', 'BOM', 'JFK').", args.Origin)), nil
	}
	if destID == "" {
		return toolErr(fmt.Sprintf("Could not recognize destination %q. Ask the user for a known city or airport.", args.Destination)), nil
	}
	if args.Date == "" {
		return toolErr("Missing date. Ask the user for a departure date in YYYY-MM-DD."), nil
	}

	byID, _ := s.buildAirportIndex(ctx)
	results := s.flightSvc.Search(ctx, originID, destID, args.Date)

	type flightOut struct {
		FlightID       string  `json:"flight_id"`
		FlightNumber   string  `json:"flight_number"`
		Origin         string  `json:"origin"`
		Destination    string  `json:"destination"`
		DepartureTime  string  `json:"departure_time"`
		ArrivalTime    string  `json:"arrival_time"`
		DurationMin    int     `json:"duration_minutes"`
		PriceUSD       float64 `json:"price_usd"`
		SeatsAvailable int     `json:"seats_available"`
	}
	out := struct {
		Origin      string      `json:"origin"`
		Destination string      `json:"destination"`
		Date        string      `json:"date"`
		Count       int         `json:"count"`
		Flights     []flightOut `json:"flights"`
	}{
		Origin:      describeAirport(byID, originID),
		Destination: describeAirport(byID, destID),
		Date:        args.Date,
	}

	// Cap to 10 flights to keep the context small for the model.
	for i, f := range results {
		if i >= 10 {
			break
		}
		out.Flights = append(out.Flights, flightOut{
			FlightID:       f.ID,
			FlightNumber:   f.FlightNumber,
			Origin:         describeAirport(byID, f.OriginID),
			Destination:    describeAirport(byID, f.DestinationID),
			DepartureTime:  f.DepartureTime.Format(time.RFC3339),
			ArrivalTime:    f.ArrivalTime.Format(time.RFC3339),
			DurationMin:    int(f.ArrivalTime.Sub(f.DepartureTime).Minutes()),
			PriceUSD:       float64(f.Price) / 100.0,
			SeatsAvailable: f.SeatsAvailable,
		})
	}
	out.Count = len(out.Flights)
	return toJSON(out), nil
}

func describeAirport(byID map[string]airportInfo, id string) string {
	if info, ok := byID[id]; ok {
		if info.CityName != "" {
			return fmt.Sprintf("%s (%s)", info.CityName, info.Code)
		}
		return info.Code
	}
	return id
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func toolErr(msg string) string {
	return toJSON(map[string]string{"error": msg})
}

func (s *Service) getBooking(ctx context.Context, userID, argsJSON string) (string, error) {
	if !isAuthed(userID) {
		return toolErr("User is not signed in. Ask them to sign in to view bookings."), nil
	}
	var args struct {
		BookingID string `json:"booking_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	b, appErr := s.bookingSvc.GetByID(ctx, args.BookingID)
	if appErr != nil {
		return toolErr("Booking not found."), nil
	}
	if b.UserID != userID {
		return toolErr("You are not authorized to view this booking."), nil
	}
	return toJSON(s.enrichBooking(ctx, b)), nil
}

func (s *Service) getMyBookings(ctx context.Context, userID string) (string, error) {
	if !isAuthed(userID) {
		return toolErr("User is not signed in. Ask them to sign in to view bookings."), nil
	}
	list, appErr := s.bookingSvc.ListByUser(ctx, userID)
	if appErr != nil {
		return toolErr("Unable to fetch bookings."), nil
	}
	out := struct {
		Count    int            `json:"count"`
		Bookings []bookingOut   `json:"bookings"`
	}{}
	for _, b := range list {
		out.Bookings = append(out.Bookings, s.enrichBooking(ctx, b))
	}
	out.Count = len(out.Bookings)
	return toJSON(out), nil
}

func (s *Service) createBooking(ctx context.Context, userID, argsJSON string) (string, error) {
	if !isAuthed(userID) {
		return toolErr("User is not signed in. Ask them to sign in before booking."), nil
	}
	var args struct {
		FlightID       string `json:"flight_id"`
		Seats          int    `json:"seats"`
		PassengerName  string `json:"passenger_name"`
		PassengerEmail string `json:"passenger_email"`
		PassengerPhone string `json:"passenger_phone"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if args.FlightID == "" {
		return toolErr("Missing flight_id. Run search_flights first and ask the user which flight to book."), nil
	}
	if args.PassengerName == "" || args.PassengerEmail == "" {
		return toolErr("Missing passenger_name or passenger_email. Ask the user for these details."), nil
	}
	if args.Seats <= 0 {
		args.Seats = 1
	}

	resp, appErr := s.bookingSvc.Create(ctx, userID, bookings.CreateBookingRequest{
		FlightID:       args.FlightID,
		Seats:          args.Seats,
		PassengerName:  args.PassengerName,
		PassengerEmail: args.PassengerEmail,
		PassengerPhone: args.PassengerPhone,
	})
	if appErr != nil {
		return toolErr("Booking failed: " + appErr.Message), nil
	}

	return toJSON(map[string]any{
		"booking_id":        resp.BookingID,
		"status":            resp.Status,
		"amount_usd":        float64(resp.Amount) / 100.0,
		"payment_intent_id": resp.PaymentIntentID,
		"checkout_url":      resp.CheckoutURL,
		"message":           "Booking created. If checkout_url is present, tell the user to open it to complete payment.",
	}), nil
}

func (s *Service) cancelBooking(ctx context.Context, userID, argsJSON string) (string, error) {
	if !isAuthed(userID) {
		return toolErr("User is not signed in. Ask them to sign in before cancelling bookings."), nil
	}
	var args struct {
		BookingID string `json:"booking_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	b, appErr := s.bookingSvc.GetByID(ctx, args.BookingID)
	if appErr != nil {
		return toolErr("Booking not found."), nil
	}
	if b.UserID != userID {
		return toolErr("You are not authorized to cancel this booking."), nil
	}

	result, appErr := s.bookingSvc.CancelBooking(ctx, args.BookingID)
	if appErr != nil {
		return toolErr("Cancel failed: " + appErr.Message), nil
	}
	return toJSON(map[string]any{
		"booking_id": result.ID,
		"status":     result.Status,
		"message":    "Booking cancelled.",
	}), nil
}

func (s *Service) sendConfirmationEmail(ctx context.Context, userID, argsJSON string) (string, error) {
	if !isAuthed(userID) {
		return toolErr("User is not signed in."), nil
	}
	var args struct {
		BookingID string `json:"booking_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	b, appErr := s.bookingSvc.GetByID(ctx, args.BookingID)
	if appErr != nil {
		return toolErr("Booking not found."), nil
	}
	if b.UserID != userID {
		return toolErr("You are not authorized to send confirmation for this booking."), nil
	}

	if s.publisher == nil {
		return toolErr("Email service is not configured."), nil
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
		Status:         "confirmed",
	})
	return toJSON(map[string]any{
		"sent_to": b.PassengerEmail,
		"message": "Confirmation email queued.",
	}), nil
}

func (s *Service) getAirports(ctx context.Context) (string, error) {
	list := s.airportSvc.List(ctx)
	type item struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	out := make([]item, 0, len(list))
	for _, a := range list {
		out = append(out, item{Code: a.Code, Name: a.Name})
	}
	return toJSON(map[string]any{"count": len(out), "airports": out}), nil
}

func (s *Service) getCities(ctx context.Context) (string, error) {
	list := s.citySvc.List(ctx)
	type item struct {
		Code    string `json:"code"`
		Name    string `json:"name"`
		Country string `json:"country"`
	}
	out := make([]item, 0, len(list))
	for _, c := range list {
		out = append(out, item{Code: c.Code, Name: c.Name, Country: c.Country})
	}
	return toJSON(map[string]any{"count": len(out), "cities": out}), nil
}

type bookingOut struct {
	BookingID      string  `json:"booking_id"`
	Status         string  `json:"status"`
	Seats          int     `json:"seats"`
	AmountUSD      float64 `json:"amount_usd"`
	PassengerName  string  `json:"passenger_name"`
	PassengerEmail string  `json:"passenger_email"`
	FlightNumber   string  `json:"flight_number,omitempty"`
	Origin         string  `json:"origin,omitempty"`
	Destination    string  `json:"destination,omitempty"`
	DepartureTime  string  `json:"departure_time,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

func (s *Service) enrichBooking(ctx context.Context, b *bookings.Booking) bookingOut {
	out := bookingOut{
		BookingID:      b.ID,
		Status:         b.Status,
		Seats:          b.Seats,
		AmountUSD:      float64(b.Amount) / 100.0,
		PassengerName:  b.PassengerName,
		PassengerEmail: b.PassengerEmail,
		CreatedAt:      b.CreatedAt.Format(time.RFC3339),
	}
	if f, err := s.flightSvc.GetByID(ctx, b.FlightID); err == nil {
		byID, _ := s.buildAirportIndex(ctx)
		out.FlightNumber = f.FlightNumber
		out.Origin = describeAirport(byID, f.OriginID)
		out.Destination = describeAirport(byID, f.DestinationID)
		out.DepartureTime = f.DepartureTime.Format(time.RFC3339)
	}
	return out
}

func isAuthed(userID string) bool {
	return userID != "" && userID != "anonymous"
}

// Common user-facing aliases that aren't in the seed data. Case-insensitive.
var locationAliases = map[string]string{
	"nyc":       "JFK",
	"new york":  "JFK",
	"la":        "LAX",
	"bombay":    "BOM",
	"bengaluru": "BLR",
	"vegas":     "LAS",
	"sf":        "SFO",
}

type airportInfo struct {
	ID       string
	Code     string
	Name     string
	CityID   string
	CityCode string
	CityName string
}

// buildAirportIndex returns a map keyed by airport UUID and a resolver map keyed
// by lowercase code/name (both airport and city) → airport UUID.
func (s *Service) buildAirportIndex(ctx context.Context) (byID map[string]airportInfo, byAlias map[string]string) {
	byID = map[string]airportInfo{}
	byAlias = map[string]string{}

	cityByID := map[string]struct {
		Code, Name string
	}{}
	for _, c := range s.citySvc.List(ctx) {
		cityByID[c.ID] = struct {
			Code, Name string
		}{c.Code, c.Name}
	}

	for _, a := range s.airportSvc.List(ctx) {
		info := airportInfo{ID: a.ID, Code: a.Code, Name: a.Name, CityID: a.CityID}
		if city, ok := cityByID[a.CityID]; ok {
			info.CityCode = city.Code
			info.CityName = city.Name
		}
		byID[a.ID] = info
		if a.Code != "" {
			byAlias[strings.ToLower(a.Code)] = a.ID
		}
		if a.Name != "" {
			byAlias[strings.ToLower(a.Name)] = a.ID
		}
		if info.CityCode != "" {
			byAlias[strings.ToLower(info.CityCode)] = a.ID
		}
		if info.CityName != "" {
			byAlias[strings.ToLower(info.CityName)] = a.ID
		}
	}
	return byID, byAlias
}

// resolveLocation always returns an airport UUID if it can map the input, or
// the empty string otherwise. Flight searches require an airport UUID to match
// flights.origin_id / flights.destination_id.
func (s *Service) resolveLocation(ctx context.Context, input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	// Already looks like a UUID — trust it.
	if len(input) == 36 && strings.Count(input, "-") == 4 {
		return input
	}
	_, byAlias := s.buildAirportIndex(ctx)
	key := strings.ToLower(input)
	if id, ok := byAlias[key]; ok {
		return id
	}
	if alias, ok := locationAliases[key]; ok {
		if id, ok := byAlias[strings.ToLower(alias)]; ok {
			return id
		}
	}
	return ""
}

// buildGroqMessages converts a conversation to Groq message format
func (s *Service) buildGroqMessages(conv *Conversation) []groqMsg {
	var msgs []groqMsg

	// Add system prompt
	systemPrompt := os.Getenv("ASSISTANT_SYSTEM_PROMPT")
	if systemPrompt == "" {
		systemPrompt = SystemPrompt
	}
	msgs = append(msgs, groqMsg{Role: "system", Content: systemPrompt})

	// Per-turn context: current date + auth status so the model can resolve
	// relative dates ("tomorrow") and gate auth-required actions.
	authLine := "The user is NOT signed in. They can search flights, but booking/viewing/cancelling bookings requires sign-in."
	if conv.UserID != "" && conv.UserID != "anonymous" {
		authLine = fmt.Sprintf("The user is signed in (user_id=%s). They can search, book, view, and cancel bookings.", conv.UserID)
	}
	ctxMsg := fmt.Sprintf("Today's date is %s (local time). %s", time.Now().Format("2006-01-02"), authLine)
	msgs = append(msgs, groqMsg{Role: "system", Content: ctxMsg})

	// Add conversation history, skipping orphan/malformed tool turns
	// (e.g. older conversations that predate tool_call_id support).
	// Also window the history to the last ~16 messages to keep token cost low.
	const historyWindow = 16
	start := 0
	if len(conv.Messages) > historyWindow {
		start = len(conv.Messages) - historyWindow
		// Don't start on a stray tool/assistant-with-tool_calls pair — walk back
		// to the nearest user message to keep the trail coherent.
		for start > 0 && conv.Messages[start].Role != "user" {
			start--
		}
	}
	for _, m := range conv.Messages[start:] {
		if m.Role == "tool" && m.ToolCallID == "" {
			continue
		}
		content := m.Content
		if m.Role == "tool" && len(content) > 2000 {
			content = content[:2000] + `..."truncated":true}`
		}
		gm := groqMsg{Role: m.Role, Content: content}
		if m.Role == "tool" {
			gm.ToolCallID = m.ToolCallID
		}
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			gm.ToolCalls = make([]groqToolCall, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				if tc.ID == "" {
					continue
				}
				gm.ToolCalls = append(gm.ToolCalls, groqToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: groqToolCallFunc{
						Name:      tc.Name,
						Arguments: tc.Args,
					},
				})
			}
			if len(gm.ToolCalls) == 0 {
				// No valid tool calls and possibly empty content — skip to avoid confusing the model
				if gm.Content == "" {
					continue
				}
			}
		}
		msgs = append(msgs, gm)
	}

	return msgs
}

// saveConversation stores a conversation in Redis
func (s *Service) saveConversation(ctx context.Context, conv *Conversation) {
	if s.rdb == nil {
		return
	}
	data, err := json.Marshal(conv)
	if err != nil {
		s.log.Warn("failed to marshal conversation", "error", err)
		return
	}
	key := convKeyPrefix + conv.ID
	if err := s.rdb.Set(ctx, key, data, conversationTTL).Err(); err != nil {
		s.log.Warn("failed to save conversation", "error", err)
	}
}

// loadConversation retrieves a conversation from Redis
func (s *Service) loadConversation(ctx context.Context, convID string) *Conversation {
	if s.rdb == nil {
		return nil
	}
	key := convKeyPrefix + convID
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil
	}
	return &conv
}

// ExtractUserIDFromToken extracts the user ID from a JWT token string
func ExtractUserIDFromToken(tokenStr string) string {
	c, err := auth.ParseToken(tokenStr)
	if err != nil {
		return ""
	}
	return c.UserID
}