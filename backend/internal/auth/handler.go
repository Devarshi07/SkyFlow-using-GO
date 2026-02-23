package auth

import (
	"encoding/json"
	"net/http"
	"strings"

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
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.Post("/google", h.GoogleAuth)
	r.Get("/me", h.Me)
	r.Put("/profile", h.UpdateProfile)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	res, appErr := h.svc.Login(r.Context(), req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	res, appErr := h.svc.Register(r.Context(), req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteCreated(w, r, res)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	res, appErr := h.svc.Refresh(r.Context(), req.RefreshToken)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if appErr := h.svc.Logout(r.Context(), req.RefreshToken); appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteNoContent(w)
}

func (h *Handler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	var req GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	res, appErr := h.svc.GoogleLogin(r.Context(), req.Code, req.RedirectURI)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, res)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := h.extractUserID(r)
	if userID == "" {
		response.WriteError(w, r, apperrors.Unauthorized(""), h.log)
		return
	}
	u, appErr := h.svc.GetProfile(r.Context(), userID)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, u)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := h.extractUserID(r)
	if userID == "" {
		response.WriteError(w, r, apperrors.Unauthorized(""), h.log)
		return
	}
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, apperrors.BadRequest("invalid JSON"), h.log)
		return
	}
	u, appErr := h.svc.UpdateProfile(r.Context(), userID, req)
	if appErr != nil {
		response.WriteError(w, r, appErr.WithRequestID(middleware.GetRequestID(r.Context())), h.log)
		return
	}
	response.WriteOK(w, r, u)
}

func (h *Handler) extractUserID(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	c, err := ParseToken(tokenStr)
	if err != nil {
		return ""
	}
	return c.UserID
}
