package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/tair/full-observability/api-gateway/config"
	"github.com/tair/full-observability/api-gateway/middleware"
	"github.com/tair/full-observability/api-gateway/routes"
	"github.com/tair/full-observability/pkg/logger"
	"github.com/tair/full-observability/pkg/tracing"
)

func main() {
	// Initialize logger
	serviceName := getEnv("OTEL_SERVICE_NAME", "api-gateway")
	isDevelopment := getEnv("ENVIRONMENT", "development") == "development"
	logger.Init(serviceName, isDevelopment)

	logLevel := getEnv("LOG_LEVEL", "info")
	logger.SetLevel(logLevel)

	logger.Logger.Info().
		Str("service", serviceName).
		Str("environment", getEnv("ENVIRONMENT", "development")).
		Msg("Starting API Gateway")

	// Initialize tracer
	tp, err := tracing.InitTracer(serviceName)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to initialize tracer")
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracing.Shutdown(ctx, tp); err != nil {
				logger.Logger.Error().Err(err).Msg("Failed to shutdown tracer")
			}
		}()
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:           "API Gateway",
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		EnablePrintRoutes: true,
		ErrorHandler:      customErrorHandler,
	})

	// Global middleware
	setupMiddleware(app)

	// Setup routes
	routes.SetupRoutes(app, cfg)

	// Start server
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("🚀 API Gateway starting on %s", addr)
		log.Printf("📊 Routing to services:")
		for name, svc := range cfg.Services {
			log.Printf("   - %s: %s", name, svc.BaseURL)
		}

		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down API Gateway...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ API Gateway stopped")
}

// setupMiddleware configures global middleware
func setupMiddleware(app *fiber.App) {
	// Recover from panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// Request ID (must be first)
	app.Use(requestid.New())

	// OpenTelemetry Tracing (second - after request ID)
	app.Use(middleware.TracingMiddleware())

	// Structured Logging (third - after tracing for trace ID)
	app.Use(middleware.StructuredLoggingMiddleware())

	// Basic Fiber Logger (optional - for quick debugging)
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
	}))

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-Id, traceparent, tracestate",
		AllowCredentials: true,
		ExposeHeaders:    "X-Request-Id, X-Trace-Id",
	}))

	// Compression
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":      err.Error(),
		"statusCode": code,
		"path":       c.Path(),
		"method":     c.Method(),
		"requestId":  c.Get("X-Request-Id"),
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

