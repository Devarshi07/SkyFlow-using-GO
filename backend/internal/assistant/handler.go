package assistant

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/skyflow/skyflow/internal/shared/errors"
	"github.com/skyflow/skyflow/internal/shared/logger"
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
	r.Post("/chat", h.Chat)
	r.Get("/conversation/{id}", h.GetConversation)
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		// Allow anonymous users to search flights but not make bookings
		userID = "anonymous"
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, errors.BadRequest("invalid JSON"), h.log)
		return
	}

	if req.Message == "" {
		response.WriteError(w, r, errors.BadRequest("message is required"), h.log)
		return
	}

	resp, err := h.svc.Chat(r.Context(), userID, &req)
	if err != nil {
		response.WriteError(w, r, errors.Internal(err), h.log)
		return
	}

	response.WriteOK(w, r, resp)
}

func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	if userID == "" {
		response.WriteError(w, r, errors.Unauthorized(""), h.log)
		return
	}

	convID := chi.URLParam(r, "id")
	if convID == "" {
		response.WriteError(w, r, errors.BadRequest("conversation id is required"), h.log)
		return
	}

	conv, err := h.svc.GetConversation(r.Context(), convID, userID)
	if err != nil {
		response.WriteError(w, r, errors.NotFound("conversation"), h.log)
		return
	}

	response.WriteOK(w, r, conv)
}

func extractUserID(r *http.Request) string {
	hdr := r.Header.Get("Authorization")
	if hdr == "" || !strings.HasPrefix(hdr, "Bearer ") {
		return ""
	}
	tokenStr := strings.TrimPrefix(hdr, "Bearer ")
	return ExtractUserIDFromToken(tokenStr)
}