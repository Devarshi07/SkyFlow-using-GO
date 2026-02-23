package cities

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
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cities := h.svc.List(r.Context())
	response.WriteOK(w, r, map[string]any{"cities": cities})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, appErr := h.svc.GetByID(r.Context(), id)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, c)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCityRequest
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

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateCityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	c, appErr := h.svc.Update(r.Context(), id, req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, c)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if appErr := h.svc.Delete(r.Context(), id); appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteNoContent(w)
}
