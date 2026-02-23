package bookings

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/skyflow/skyflow/internal/auth"
	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
	"github.com/skyflow/skyflow/internal/shared/logger"
	"github.com/skyflow/skyflow/internal/shared/middleware"
	"github.com/skyflow/skyflow/internal/shared/response"
)

type Handler struct {
	svc *Service
	log *logger.Logger
}

func NewHandler(svc *Service, log *logger.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Routes(r chi.Router) {
	r.Post("/", h.Create)
	r.Post("/confirm", h.Confirm)
	r.Get("/my", h.ListMy)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Edit)
	r.Post("/{id}/cancel", h.Cancel)
	r.Post("/{id}/confirm-payment", h.ConfirmByBookingID)
	r.Post("/{id}/confirm-edit", h.ConfirmEdit)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		response.WriteError(w, r, apperrors.Unauthorized(""), h.log)
		return
	}
	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	res, appErr := h.svc.Create(r.Context(), userID, req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteCreated(w, r, res)
}

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	b, appErr := h.svc.Confirm(r.Context(), req.PaymentIntentID)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, b)
}

// ConfirmByBookingID is called after Stripe Checkout redirects back.
// It verifies the Stripe Session status and confirms the booking.
func (h *Handler) ConfirmByBookingID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		SessionID string `json:"session_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	// If we have a Stripe session ID, use session-based confirmation
	if req.SessionID != "" {
		b, appErr := h.svc.ConfirmBySession(r.Context(), id, req.SessionID)
		if appErr != nil {
			response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
			return
		}
		response.WriteOK(w, r, b)
		return
	}

	// Fallback: look up by payment_intent_id
	booking, appErr := h.svc.GetByID(r.Context(), id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	if booking.Status == "confirmed" {
		response.WriteOK(w, r, booking)
		return
	}
	if booking.PaymentIntentID == "" {
		response.WriteError(w, r, apperrors.BadRequest("no payment intent for this booking"), h.log)
		return
	}
	b, confirmErr := h.svc.Confirm(r.Context(), booking.PaymentIntentID)
	if confirmErr != nil {
		response.WriteError(w, r, confirmErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, b)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		response.WriteError(w, r, apperrors.Unauthorized(""), h.log)
		return
	}
	id := chi.URLParam(r, "id")
	var req EditBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	res, appErr := h.svc.EditBooking(r.Context(), userID, id, req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) ConfirmEdit(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		response.WriteError(w, r, apperrors.Unauthorized(""), h.log)
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		PaymentIntentID string `json:"payment_intent_id"`
		SessionID       string `json:"session_id"`
		NewFlightID     string `json:"new_flight_id"`
		NewSeats        int    `json:"new_seats"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	b, appErr := h.svc.ConfirmEdit(r.Context(), userID, id, req.PaymentIntentID, req.SessionID, req.NewFlightID, req.NewSeats)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, b)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	b, appErr := h.svc.CancelBooking(r.Context(), id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, b)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	b, appErr := h.svc.GetByID(r.Context(), id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, b)
}

func (h *Handler) ListMy(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		response.WriteError(w, r, apperrors.Unauthorized(""), h.log)
		return
	}
	list, appErr := h.svc.ListByUser(r.Context(), userID)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, map[string]any{"bookings": list})
}

func extractUserID(r *http.Request) string {
	hdr := r.Header.Get("Authorization")
	if hdr == "" || !strings.HasPrefix(hdr, "Bearer ") {
		return ""
	}
	tokenStr := strings.TrimPrefix(hdr, "Bearer ")
	c, err := auth.ParseToken(tokenStr)
	if err != nil {
		return ""
	}
	return c.UserID
}
