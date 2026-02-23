package payments

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
	r.Post("/intent", h.CreateIntent)
	r.Get("/methods", h.GetMethods)
	r.Get("/{paymentIntentId}", h.Get)
	r.Post("/{paymentIntentId}/confirm", h.Confirm)
	r.Post("/{paymentIntentId}/refund", h.Refund)
	r.Post("/{paymentIntentId}/cancel", h.Cancel)
}

func (h *Handler) CreateIntent(w http.ResponseWriter, r *http.Request) {
	var req CreateIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	res, appErr := h.svc.CreateIntent(req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteCreated(w, r, res)
}

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "paymentIntentId")
	res, appErr := h.svc.Confirm(id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "paymentIntentId")
	res, appErr := h.svc.Get(id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "paymentIntentId")
	var req RefundRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	res, appErr := h.svc.Refund(id, req.Amount)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "paymentIntentId")
	res, appErr := h.svc.Cancel(id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) GetMethods(w http.ResponseWriter, r *http.Request) {
	methods := h.svc.GetMethods()
	response.WriteOK(w, r, map[string]any{"methods": methods})
}
