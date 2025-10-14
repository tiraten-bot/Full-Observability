package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware adds OpenTelemetry tracing to requests
func TracingMiddleware() fiber.Handler {
	tracer := otel.Tracer("api-gateway")

	return func(c *fiber.Ctx) error {
		// Start a new span
		ctx, span := tracer.Start(
			c.UserContext(),
			c.Method()+" "+c.Path(),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.target", c.Path()),
				attribute.String("http.scheme", c.Protocol()),
				attribute.String("http.host", c.Hostname()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.client_ip", c.IP()),
				attribute.String("net.peer.ip", c.IP()),
			),
		)
		defer span.End()

		// Store span in context
		c.SetUserContext(ctx)

		// Inject trace context into headers for backend services
		carrier := propagation.HeaderCarrier{}
		otel.GetTextMapPropagator().Inject(ctx, carrier)
		
		for key, values := range carrier {
			for _, value := range values {
				c.Request().Header.Set(key, value)
			}
		}

		// Add trace ID to response header
		if span.SpanContext().HasTraceID() {
			c.Set("X-Trace-Id", span.SpanContext().TraceID().String())
		}

		// Process request
		err := c.Next()

		// Record response status
		statusCode := c.Response().StatusCode()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int("http.response.size", len(c.Response().Body())),
		)

		// Set span status based on HTTP status code
		if statusCode >= 500 {
			span.SetStatus(codes.Error, "Server Error")
		} else if statusCode >= 400 {
			span.SetStatus(codes.Error, "Client Error")
		} else {
			span.SetStatus(codes.Ok, "Success")
		}

		// Record error if exists
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}

