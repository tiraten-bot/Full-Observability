package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tair/full-observability/pkg/auth"
	"github.com/tair/full-observability/pkg/logger"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
	RoleKey     contextKey = "role"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Logger.Warn().Msg("Missing authorization header")
			respondError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Logger.Warn().Msg("Invalid authorization header format")
			respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]
		claims, err := auth.ValidateToken(token)
		if err != nil {
			logger.Logger.Warn().Err(err).Msg("Invalid token")
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		logger.Logger.Debug().
			Uint("user_id", claims.UserID).
			Str("username", claims.Username).
			Str("role", claims.Role).
			Msg("User authenticated")

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AdminMiddleware checks if user has admin role
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(RoleKey).(string)
		if !ok || role != "admin" {
			logger.Logger.Warn().
				Str("role", role).
				Msg("Admin access denied")
			respondError(w, http.StatusForbidden, "Admin access required")
			return
		}

		logger.Logger.Debug().Msg("Admin access granted")
		next.ServeHTTP(w, r)
	})
}

// OptionalAuthMiddleware validates JWT token if present, but doesn't require it
func OptionalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			claims, err := auth.ValidateToken(token)
			if err == nil {
				logger.Logger.Debug().
					Uint("user_id", claims.UserID).
					Str("username", claims.Username).
					Msg("Optional auth: User identified")

				// Valid token, add to context
				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
				ctx = context.WithValue(ctx, UsernameKey, claims.Username)
				ctx = context.WithValue(ctx, RoleKey, claims.Role)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	}
}

// Helper function for error responses
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}

