package main

import (
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/tair/full-observability/internal/inventory"
	httpDelivery "github.com/tair/full-observability/internal/inventory/delivery/http"
	"github.com/tair/full-observability/internal/inventory/domain"
	"github.com/tair/full-observability/pkg/database"
	"github.com/tair/full-observability/pkg/logger"
)

func main() {
	// Initialize logger
	serviceName := getEnv("OTEL_SERVICE_NAME", "inventory-service")
	isDevelopment := getEnv("ENVIRONMENT", "development") == "development"
	logger.Init(serviceName, isDevelopment)

	logLevel := getEnv("LOG_LEVEL", "info")
	logger.SetLevel(logLevel)

	logger.Logger.Info().
		Str("service", serviceName).
		Str("environment", getEnv("ENVIRONMENT", "development")).
		Str("log_level", logLevel).
		Msg("Starting inventory service")

	// Load database configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "inventorydb"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to database
	db, err := database.NewGormConnection(dbConfig)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to get database instance")
	}
	defer sqlDB.Close()

	// Run migrations
	if err := db.AutoMigrate(&domain.Inventory{}); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to run migrations")
	}

	logger.Logger.Info().Msg("Database initialized successfully")

	// Get User Service gRPC address
	userServiceAddr := getEnv("USER_SERVICE_GRPC_ADDR", "localhost:9090")

	// Initialize handler with Wire DI (includes User Service gRPC client)
	handler, err := inventory.InitializeHTTPHandler(db, userServiceAddr)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize handler")
	}

	logger.Logger.Info().
		Str("user_service_grpc", userServiceAddr).
		Msg("Inventory handler initialized with User Service client")

	// Start HTTP server
	httpPort := getEnv("HTTP_PORT", "8082")
	startHTTPServer(handler, sqlDB, httpPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info().Msg("Shutting down server...")
}

func startHTTPServer(handler *httpDelivery.InventoryHandler, db *sql.DB, port string) {
	// Setup router
	router := mux.NewRouter()

	// Register routes
	handler.RegisterRoutes(router)

	// Health check endpoint
	handler.RegisterHealthCheck(router, db)

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	logger.Logger.Info().
		Str("port", port).
		Str("metrics_endpoint", "/metrics").
		Msg("HTTP server started")

	if err := http.ListenAndServe(":"+port, c.Handler(router)); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to start HTTP server")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

