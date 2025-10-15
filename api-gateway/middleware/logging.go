package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"

	"github.com/tair/full-observability/pkg/logger"
)

// StructuredLoggingMiddleware provides structured logging for requests
func StructuredLoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Get trace ID if available
		traceID := "no-trace"
		if span := trace.SpanFromContext(c.UserContext()); span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}

		// Get request ID
		requestID := c.Get("X-Request-Id")

		// Log request start
		logger.Info(c.UserContext()).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Str("trace_id", traceID).
			Str("request_id", requestID).
			Msg("Gateway request started")

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Log request completion
		logEvent := logger.WithContext(c.UserContext()).Info()
		if statusCode >= 500 {
			logEvent = logger.WithContext(c.UserContext()).Error()
		} else if statusCode >= 400 {
			logEvent = logger.WithContext(c.UserContext()).Warn()
		}

		logEvent.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", statusCode).
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Int("response_size", len(c.Response().Body())).
			Str("trace_id", traceID).
			Str("request_id", requestID).
			Msg("Gateway request completed")

		// Log error if exists
		if err != nil {
			logger.Error(c.UserContext()).
				Err(err).
				Str("method", c.Method()).
				Str("path", c.Path()).
				Str("trace_id", traceID).
				Msg("Gateway request error")
		}

		return err
	}
}
