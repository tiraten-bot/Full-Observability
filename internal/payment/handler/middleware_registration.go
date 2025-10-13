package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tair/full-observability/internal/payment/client"
)

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	EnableLogging bool
	EnableTracing bool
	UserClient    *client.UserServiceClient
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig(userClient *client.UserServiceClient) MiddlewareConfig {
	return MiddlewareConfig{
		EnableLogging: true,
		EnableTracing: true,
		UserClient:    userClient,
	}
}

// RegisterMiddlewares registers all middlewares to the router
func RegisterMiddlewares(router *mux.Router, config MiddlewareConfig) {
	// Logging middleware (first in chain)
	if config.EnableLogging {
		router.Use(LoggingMiddleware)
	}

	// Tracing middleware (second in chain)
	if config.EnableTracing {
		router.Use(func(next http.Handler) http.Handler {
			return TracingMiddleware("http-request", next)
		})
	}
}

// GetAuthMiddleware returns the auth middleware with user client
func (config MiddlewareConfig) GetAuthMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(config.UserClient)
}

// GetAdminMiddleware returns the admin middleware with user client
func (config MiddlewareConfig) GetAdminMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return AdminMiddleware(config.UserClient)
}

