package response

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
	"github.com/skyflow/skyflow/internal/shared/logger"
)

// WriteJSON writes a JSON response with status code
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteError writes an AppError as JSON and logs it. Caller should set err.RequestID via err.WithRequestID().
func WriteError(w http.ResponseWriter, r *http.Request, err *apperrors.AppError, log *logger.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	_, _ = w.Write(err.JSON())
	if err.HTTPStatus >= 500 && log != nil {
		log.Error("server error", "request_id", err.RequestID, "code", err.Code, "error", err.Error())
	}
}

// WriteOK writes a 200 JSON response
func WriteOK(w http.ResponseWriter, r *http.Request, data any) {
	WriteJSON(w, r, http.StatusOK, data)
}

// WriteCreated writes a 201 JSON response
func WriteCreated(w http.ResponseWriter, r *http.Request, data any) {
	WriteJSON(w, r, http.StatusCreated, data)
}

// WriteNoContent writes 204
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
