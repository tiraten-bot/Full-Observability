package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/tair/full-observability/internal/payment"
	"github.com/tair/full-observability/internal/payment/domain"
	"github.com/tair/full-observability/internal/payment/handler"
	"github.com/tair/full-observability/pkg/database"
	"github.com/tair/full-observability/pkg/logger"
	"github.com/tair/full-observability/pkg/tracing"
)

func main() {
	// Initialize logger
	serviceName := getEnv("OTEL_SERVICE_NAME", "payment-service")
	isDevelopment := getEnv("ENVIRONMENT", "development") == "development"
	logger.Init(serviceName, isDevelopment)

	logLevel := getEnv("LOG_LEVEL", "info")
	logger.SetLevel(logLevel)

	logger.Logger.Info().
		Str("service", serviceName).
		Str("environment", getEnv("ENVIRONMENT", "development")).
		Str("log_level", logLevel).
		Msg("Starting payment service")

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

	// Load database configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "paymentdb"),
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
	if err := db.AutoMigrate(&domain.Payment{}); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to run migrations")
	}

	logger.Logger.Info().Msg("Database initialized successfully")

	// Get gRPC service addresses
	serviceAddrs := payment.ServiceAddrs{
		UserServiceAddr:      getEnv("USER_SERVICE_GRPC_ADDR", "localhost:9090"),
		ProductServiceAddr:   getEnv("PRODUCT_SERVICE_GRPC_ADDR", "localhost:9091"),
		InventoryServiceAddr: getEnv("INVENTORY_SERVICE_GRPC_ADDR", "localhost:9092"),
	}

	// Initialize handler with Wire DI (includes User, Product & Inventory Service gRPC clients)
	paymentHandler, err := payment.InitializeHandler(db, serviceAddrs)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize handler")
	}

	logger.Logger.Info().
		Str("user_service_grpc", serviceAddrs.UserServiceAddr).
		Str("product_service_grpc", serviceAddrs.ProductServiceAddr).
		Str("inventory_service_grpc", serviceAddrs.InventoryServiceAddr).
		Msg("Payment handler initialized with User, Product & Inventory Service clients")

	// Start HTTP server
	httpPort := getEnv("HTTP_PORT", "8083")
	startHTTPServer(paymentHandler, sqlDB, httpPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info().Msg("Shutting down server...")
}

func startHTTPServer(paymentHandler *handler.PaymentHandler, db *sql.DB, port string) {
	// Setup router
	router := mux.NewRouter()

	// Get middleware configuration
	middlewareConfig := paymentHandler.GetMiddlewareConfig()

	// Register all middlewares using middleware registration system
	handler.RegisterMiddlewares(router, middlewareConfig)

	// Register routes
	paymentHandler.RegisterRoutes(router)

	// Health check endpoint
	paymentHandler.RegisterHealthCheck(router, db)

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

