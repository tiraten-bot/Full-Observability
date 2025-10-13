package handler

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/tair/full-observability/pkg/logger"
)

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		ww := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Get trace ID
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		traceID := "no-trace"
		if span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}

		// Log request start
		logger.Info(ctx).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Str("trace_id", traceID).
			Msg("HTTP request started")

		// Call next handler
		next.ServeHTTP(ww, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion
		logEvent := logger.WithContext(ctx).Info()
		if ww.statusCode >= 400 {
			logEvent = logger.WithContext(ctx).Error()
		}

		logEvent.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Str("trace_id", traceID).
			Msg("HTTP request completed")
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

