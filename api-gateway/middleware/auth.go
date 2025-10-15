package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tair/full-observability/pkg/auth"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token := parts[1]

		// Validate token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		// Add user info to headers for backend services
		c.Request().Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		c.Request().Header.Set("X-Username", claims.Username)
		c.Request().Header.Set("X-User-Role", claims.Role)

		return c.Next()
	}
}

// AdminMiddleware checks if user has admin role
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role == nil || role.(string) != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}
		return c.Next()
	}
}

// OptionalAuthMiddleware validates token if present but doesn't require it
func OptionalAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			if claims, err := auth.ValidateToken(token); err == nil {
				c.Locals("user_id", claims.UserID)
				c.Locals("username", claims.Username)
				c.Locals("role", claims.Role)
			}
		}

		return c.Next()
	}
}
