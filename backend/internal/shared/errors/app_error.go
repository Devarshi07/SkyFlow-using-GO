package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Code represents a stable machine-readable error code for clients
type Code string

const (
	CodeInvalidRequest   Code = "INVALID_REQUEST"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeForbidden        Code = "FORBIDDEN"
	CodeNotFound         Code = "NOT_FOUND"
	CodeConflict         Code = "CONFLICT"
	CodeInternal         Code = "INTERNAL_ERROR"
	CodeValidation       Code = "VALIDATION_ERROR"
	CodeBadRequest       Code = "BAD_REQUEST"
	CodePaymentFailed    Code = "PAYMENT_FAILED"
	CodePaymentCanceled  Code = "PAYMENT_CANCELED"
	CodeCustomerNotFound Code = "CUSTOMER_NOT_FOUND"
)

// AppError is the standard application error used across all services
type AppError struct {
	Code       Code     `json:"code"`
	Message    string   `json:"message"`
	Details    any      `json:"details,omitempty"`
	HTTPStatus int      `json:"-"`
	RequestID  string   `json:"request_id,omitempty"`
	Err        error    `json:"-"` // original error, not exposed to clients
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// WithRequestID attaches request ID for tracing
func (e *AppError) WithRequestID(id string) *AppError {
	e.RequestID = id
	return e
}

// JSON marshals the error for API response (without internal Err)
func (e *AppError) JSON() []byte {
	out := struct {
		Code      Code   `json:"code"`
		Message   string `json:"message"`
		Details   any    `json:"details,omitempty"`
		RequestID string `json:"request_id,omitempty"`
	}{
		Code:      e.Code,
		Message:   e.Message,
		Details:   e.Details,
		RequestID: e.RequestID,
	}
	b, _ := json.Marshal(out)
	return b
}

// New creates a new AppError
func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewWithDetails creates AppError with optional details and wrapped error
func NewWithDetails(code Code, message string, httpStatus int, details any, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Predefined factory functions for common errors
var (
	ErrInvalidRequest = New(CodeInvalidRequest, "Invalid request", http.StatusBadRequest)
	ErrUnauthorized   = New(CodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden      = New(CodeForbidden, "Forbidden", http.StatusForbidden)
	ErrNotFound       = New(CodeNotFound, "Resource not found", http.StatusNotFound)
	ErrConflict       = New(CodeConflict, "Resource conflict", http.StatusConflict)
	ErrInternal       = New(CodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrBadRequest     = New(CodeBadRequest, "Bad request", http.StatusBadRequest)
)

// BadRequest returns 400 error with optional message override
func BadRequest(msg string) *AppError {
	if msg == "" {
		return ErrBadRequest
	}
	return New(CodeBadRequest, msg, http.StatusBadRequest)
}

// Unauthorized returns 401
func Unauthorized(msg string) *AppError {
	if msg == "" {
		return ErrUnauthorized
	}
	return New(CodeUnauthorized, msg, http.StatusUnauthorized)
}

// NotFound returns 404 with optional resource name
func NotFound(resource string) *AppError {
	msg := "Resource not found"
	if resource != "" {
		msg = resource + " not found"
	}
	return New(CodeNotFound, msg, http.StatusNotFound)
}

// Internal wraps an error for 500 response
func Internal(err error) *AppError {
	return NewWithDetails(CodeInternal, "Internal server error", http.StatusInternalServerError, nil, err)
}

// Validation returns 422 with validation details
func Validation(details any) *AppError {
	return NewWithDetails(CodeValidation, "Validation failed", http.StatusUnprocessableEntity, details, nil)
}
