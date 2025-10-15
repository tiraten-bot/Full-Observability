package handler

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TracingMiddleware wraps HTTP handlers with OpenTelemetry tracing
func TracingMiddleware(operationName string, next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, operationName)
}

// TracingMiddlewareFunc wraps HTTP handler functions with OpenTelemetry tracing
func TracingMiddlewareFunc(operationName string, next http.HandlerFunc) http.Handler {
	return otelhttp.NewHandler(next, operationName)
}
