package customers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r.Get("/{customerId}", h.GetByID)
	r.Get("/{customerId}/payments", h.GetPaymentHistory)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	c, appErr := h.svc.Create(r.Context(), req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteCreated(w, r, c)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "customerId")
	c, appErr := h.svc.GetByID(r.Context(), id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, c)
}

func (h *Handler) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "customerId")
	payments, appErr := h.svc.GetPaymentHistory(r.Context(), id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, map[string]any{"payments": payments})
}
