package middleware

import (
	"context"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func setRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID returns the request ID from context, or empty string
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
