package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/tair/full-observability/internal/inventory/client"
	"github.com/tair/full-observability/pkg/logger"
)

// MiddlewareConfig holds configuration for middlewares
type MiddlewareConfig struct {
	EnableLogging   bool
	EnableTracing   bool
	EnableCORS      bool
	EnableRecovery  bool
	EnableTimeout   bool
	TimeoutDuration time.Duration
	CORSOptions     cors.Options
	UserClient      *client.UserServiceClient
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig(userClient *client.UserServiceClient) *MiddlewareConfig {
	return &MiddlewareConfig{
		EnableLogging:   true,
		EnableTracing:   true,
		EnableCORS:      true,
		EnableRecovery:  true,
		EnableTimeout:   true,
		TimeoutDuration: 30 * time.Second,
		CORSOptions: cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		},
		UserClient: userClient,
	}
}

// RegisterMiddlewares registers all configured middlewares to the router
func RegisterMiddlewares(router *mux.Router, config *MiddlewareConfig) {
	logger.Logger.Info().
		Bool("logging", config.EnableLogging).
		Bool("tracing", config.EnableTracing).
		Bool("cors", config.EnableCORS).
		Bool("recovery", config.EnableRecovery).
		Bool("timeout", config.EnableTimeout).
		Dur("timeout_duration", config.TimeoutDuration).
		Msg("Registering middlewares")

	// Recovery middleware (first - catches panics)
	if config.EnableRecovery {
		router.Use(RecoveryMiddleware())
	}

	// Timeout middleware (early - sets request timeout)
	if config.EnableTimeout {
		router.Use(TimeoutMiddleware(config.TimeoutDuration))
	}

	// Logging middleware (early - logs all requests)
	if config.EnableLogging {
		router.Use(LoggingMiddleware)
	}

	// Tracing middleware (after logging - traces requests)
	if config.EnableTracing {
		router.Use(func(next http.Handler) http.Handler {
			return TracingMiddleware("inventory-http-request", next)
		})
	}

	// Request ID middleware (for correlation)
	router.Use(RequestIDMiddleware())

	// Security headers middleware
	router.Use(SecurityHeadersMiddleware())

	logger.Logger.Info().Msg("All middlewares registered successfully")
}

// RecoveryMiddleware recovers from panics and returns 500 error
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Logger.Error().
						Interface("panic", err).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("Panic recovered")

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware sets a timeout for HTTP requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request timeout")
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			w.Header().Set("X-Request-ID", requestID)
			r.Header.Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			// Remove server header
			w.Header().Set("Server", "")

			next.ServeHTTP(w, r)
		})
	}
}

// SetupCORS creates and configures CORS middleware
func SetupCORS(config *MiddlewareConfig) func(http.Handler) http.Handler {
	if !config.EnableCORS {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	c := cors.New(config.CORSOptions)
	return c.Handler
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple UUID-like generation for demo purposes
	// In production, use a proper UUID library
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
