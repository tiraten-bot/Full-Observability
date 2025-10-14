package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tair/full-observability/api-gateway/config"
	"github.com/tair/full-observability/api-gateway/health"
	"github.com/tair/full-observability/api-gateway/middleware"
	"github.com/tair/full-observability/api-gateway/proxy"
)

// RouteDefinition defines a route mapping
type RouteDefinition struct {
	Prefix       string
	ServiceName  string
	Description  string
	RequireAuth  bool // Requires authentication
	RequireAdmin bool // Requires admin role
}

// Routes holds all route definitions
var Routes = []RouteDefinition{
	// Public routes (no auth required)
	{
		Prefix:       "/auth",
		ServiceName:  "user",
		Description:  "Authentication endpoints (login, register)",
		RequireAuth:  false,
		RequireAdmin: false,
	},
	{
		Prefix:       "/health",
		ServiceName:  "user",
		Description:  "Health check endpoints",
		RequireAuth:  false,
		RequireAdmin: false,
	},

	// User service routes (auth required for most)
	{
		Prefix:       "/api/users",
		ServiceName:  "user",
		Description:  "User management (mixed: some need admin)",
		RequireAuth:  true,
		RequireAdmin: false,
	},

	// Product service routes
	{
		Prefix:       "/api/products",
		ServiceName:  "product",
		Description:  "Product management (mixed: some need admin)",
		RequireAuth:  true,
		RequireAdmin: false,
	},

	// Inventory service routes
	{
		Prefix:       "/api/inventory",
		ServiceName:  "inventory",
		Description:  "Inventory management (mixed: some need admin)",
		RequireAuth:  true,
		RequireAdmin: false,
	},

	// Payment service routes
	{
		Prefix:       "/api/payments",
		ServiceName:  "payment",
		Description:  "Payment management (mixed: some need admin)",
		RequireAuth:  true,
		RequireAdmin: false,
	},
}

// SetupRoutes configures all routes in the gateway
func SetupRoutes(app *fiber.App, cfg *config.GatewayConfig) {
	// Create reverse proxy
	reverseProxy := proxy.NewReverseProxy(cfg)

	// Create health checker
	healthChecker := health.NewHealthChecker(cfg)

	// Gateway quick health check (no downstream checks)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(healthChecker.QuickCheck())
	})

	// Liveness probe (for Kubernetes)
	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "alive",
		})
	})

	// Readiness probe (checks downstream services)
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
		defer cancel()

		healthStatus := healthChecker.CheckAllServices(ctx)
		
		statusCode := fiber.StatusOK
		if healthStatus.Status == "unhealthy" {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(healthStatus)
	})

	// Detailed service health checks
	app.Get("/health/services", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()

		healthStatus := healthChecker.CheckAllServices(ctx)
		return c.JSON(healthStatus)
	})

	// API routes overview
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "API Gateway",
			"version": "1.0.0",
			"routes":  Routes,
		})
	})

	// Register all service routes
	for _, route := range Routes {
		registerServiceRoutes(app, route, reverseProxy)
	}
}

// registerServiceRoutes registers all HTTP methods for a service prefix
func registerServiceRoutes(app *fiber.App, route RouteDefinition, proxyHandler *proxy.ReverseProxy) {
	// Create handler function
	handler := func(c *fiber.Ctx) error {
		return proxyHandler.ProxyRequest(c, route.ServiceName)
	}

	// Apply middleware based on route requirements
	var middlewares []fiber.Handler
	
	if route.RequireAdmin {
		// Admin routes need both auth and admin check
		middlewares = append(middlewares, middleware.AuthMiddleware(), middleware.AdminMiddleware())
	} else if route.RequireAuth {
		// Auth required routes
		middlewares = append(middlewares, middleware.AuthMiddleware())
	}
	// Public routes have no middleware

	// Create a route group for this service
	group := app.Group(route.Prefix, middlewares...)

	// Handle all HTTP methods with wildcard path
	group.All("/*", handler)

	// Also handle the exact prefix path (without /*)
	if len(middlewares) > 0 {
		app.All(route.Prefix, append(middlewares, handler)...)
	} else {
		app.All(route.Prefix, handler)
	}
}

