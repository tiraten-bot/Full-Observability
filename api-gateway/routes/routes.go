package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tair/full-observability/api-gateway/config"
	"github.com/tair/full-observability/api-gateway/proxy"
)

// RouteDefinition defines a route mapping
type RouteDefinition struct {
	Prefix      string
	ServiceName string
	Description string
}

// Routes holds all route definitions
var Routes = []RouteDefinition{
	{
		Prefix:      "/api/users",
		ServiceName: "user",
		Description: "User service endpoints (auth, users, roles)",
	},
	{
		Prefix:      "/auth",
		ServiceName: "user",
		Description: "Authentication endpoints",
	},
	{
		Prefix:      "/api/products",
		ServiceName: "product",
		Description: "Product service endpoints",
	},
	{
		Prefix:      "/api/inventory",
		ServiceName: "inventory",
		Description: "Inventory service endpoints",
	},
	{
		Prefix:      "/api/payments",
		ServiceName: "payment",
		Description: "Payment service endpoints",
	},
}

// SetupRoutes configures all routes in the gateway
func SetupRoutes(app *fiber.App, cfg *config.GatewayConfig) {
	// Create reverse proxy
	reverseProxy := proxy.NewReverseProxy(cfg)

	// Gateway health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// API routes overview
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "API Gateway",
			"version": "1.0.0",
			"routes":  Routes,
		})
	})

	// Service health checks
	app.Get("/health/services", func(c *fiber.Ctx) error {
		services := make(map[string]interface{})
		for name, svc := range cfg.Services {
			services[name] = fiber.Map{
				"url":    svc.BaseURL,
				"status": "unknown", // TODO: Implement actual health check
			}
		}
		return c.JSON(fiber.Map{
			"gateway": "healthy",
			"services": services,
		})
	})

	// Register all service routes
	for _, route := range Routes {
		registerServiceRoutes(app, route, reverseProxy)
	}
}

// registerServiceRoutes registers all HTTP methods for a service prefix
func registerServiceRoutes(app *fiber.App, route RouteDefinition, proxy *proxy.ReverseProxy) {
	// Create a route group for this service
	group := app.Group(route.Prefix)

	// Handle all HTTP methods with wildcard path
	group.All("/*", func(c *fiber.Ctx) error {
		return proxy.ProxyRequest(c, route.ServiceName)
	})

	// Also handle the exact prefix path (without /*)
	app.All(route.Prefix, func(c *fiber.Ctx) error {
		return proxy.ProxyRequest(c, route.ServiceName)
	})
}

