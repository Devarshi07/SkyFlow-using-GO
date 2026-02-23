package middleware

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
	"github.com/skyflow/skyflow/internal/shared/logger"
	"github.com/skyflow/skyflow/internal/shared/response"
)

const RequestIDHeader = "X-Request-ID"

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(RequestIDHeader)
		if rid == "" {
			rid = uuid.New().String()
		}
		ctx := setRequestID(r.Context(), rid)
		w.Header().Set(RequestIDHeader, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recovery recovers from panics and returns 500
func Recovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					rid := GetRequestID(r.Context())
					err := fmt.Errorf("panic: %v", rec)
					log.Error("panic recovered", "request_id", rid, "panic", rec, "error", err.Error())
					response.WriteError(w, r, apperrors.Internal(err).WithRequestID(rid), log)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logging logs each request and response
func Logging(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := GetRequestID(r.Context())
			log.Info("request started",
				"request_id", rid,
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			ww := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(ww, r)
			log.Info("request completed",
				"request_id", rid,
				"status", ww.status,
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
