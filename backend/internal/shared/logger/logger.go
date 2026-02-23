package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog for structured logging
type Logger struct {
	z *zerolog.Logger
}

// New creates a new Logger with optional output destination
func New(out io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}
	zl := zerolog.New(out).
		With().
		Timestamp().
		Caller().
		Logger()
	return &Logger{z: &zl}
}

// WithRequestID returns a child logger with request_id
func (l *Logger) WithRequestID(requestID string) *Logger {
	child := l.z.With().Str("request_id", requestID).Logger()
	return &Logger{z: &child}
}

// WithService returns a child logger with service name
func (l *Logger) WithService(service string) *Logger {
	child := l.z.With().Str("service", service).Logger()
	return &Logger{z: &child}
}

// Debug logs at debug level
func (l *Logger) Debug(msg string, fields ...any) {
	l.z.Debug().Fields(fieldsToMap(fields)).Msg(msg)
}

// Info logs at info level
func (l *Logger) Info(msg string, fields ...any) {
	l.z.Info().Fields(fieldsToMap(fields)).Msg(msg)
}

// Warn logs at warn level
func (l *Logger) Warn(msg string, fields ...any) {
	l.z.Warn().Fields(fieldsToMap(fields)).Msg(msg)
}

// Error logs at error level
func (l *Logger) Error(msg string, fields ...any) {
	l.z.Error().Fields(fieldsToMap(fields)).Msg(msg)
}

// Fatal logs at fatal level and exits
func (l *Logger) Fatal(msg string, fields ...any) {
	l.z.Fatal().Fields(fieldsToMap(fields)).Msg(msg)
}

// Err adds error to the event
func (l *Logger) Err(err error) *zerolog.Event {
	return l.z.Error().Err(err)
}

func fieldsToMap(fields []any) map[string]interface{} {
	m := make(map[string]interface{}, len(fields)/2)
	for i := 0; i+1 < len(fields); i += 2 {
		if k, ok := fields[i].(string); ok {
			m[k] = fields[i+1]
		}
	}
	return m
}

// Default returns a logger for development
func Default() *Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return New(output)
}
