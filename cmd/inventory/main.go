package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	pb "github.com/tair/full-observability/api/proto/inventory"
	"github.com/tair/full-observability/internal/inventory"
	grpcDelivery "github.com/tair/full-observability/internal/inventory/delivery/grpc"
	httpDelivery "github.com/tair/full-observability/internal/inventory/delivery/http"
	"github.com/tair/full-observability/internal/inventory/domain"
	"github.com/tair/full-observability/kafka"
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

	// Initialize gRPC server with Wire DI
	grpcServer, err := inventory.InitializeGRPCServer(db)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize gRPC server")
	}

	logger.Logger.Info().Msg("gRPC server initialized")

	// Initialize Kafka consumer
	kafkaBrokersStr := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaBrokers := strings.Split(kafkaBrokersStr, ",")
	kafkaConsumer, err := kafka.NewConsumer(kafkaBrokers, "inventory-service-group", []string{kafka.TopicProductPurchased})
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize Kafka consumer")
	}
	defer kafkaConsumer.Close()

	// Register event handler for product purchased events
	repo := grpcServer.GetRepository()
	kafkaConsumer.RegisterHandler(kafka.EventTypeProductPurchased, func(ctx context.Context, event kafka.ProductPurchasedEvent) error {
		logger.Logger.Info().
			Uint("product_id", event.ProductID).
			Int32("quantity", event.Quantity).
			Uint("payment_id", event.PaymentID).
			Msg("Processing product purchased event")

		// Get current inventory
		inv, err := repo.FindByProductID(event.ProductID)
		if err != nil {
			logger.Logger.Error().
				Err(err).
				Uint("product_id", event.ProductID).
				Msg("Failed to find inventory")
			return err
		}

		// Decrease stock
		newQuantity := inv.Quantity - int(event.Quantity)
		if newQuantity < 0 {
			newQuantity = 0
		}

		if err := repo.UpdateQuantity(event.ProductID, newQuantity); err != nil {
			logger.Logger.Error().
				Err(err).
				Uint("product_id", event.ProductID).
				Int("new_quantity", newQuantity).
				Msg("Failed to update inventory quantity")
			return err
		}

		logger.Logger.Info().
			Uint("product_id", event.ProductID).
			Int("old_quantity", inv.Quantity).
			Int("new_quantity", newQuantity).
			Int32("purchased", event.Quantity).
			Msg("Inventory updated successfully")

		return nil
	})

	// Start Kafka consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := kafkaConsumer.Start(ctx); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to start Kafka consumer")
	}

	logger.Logger.Info().
		Strs("kafka_brokers", kafkaBrokers).
		Str("topic", kafka.TopicProductPurchased).
		Msg("Kafka consumer started")

	// Start gRPC server in goroutine
	grpcPort := getEnv("GRPC_PORT", "9092")
	go startGRPCServer(grpcServer, grpcPort)

	// Start HTTP server in goroutine
	httpPort := getEnv("HTTP_PORT", "8082")
	go startHTTPServer(handler, sqlDB, httpPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info().Msg("Shutting down servers...")
	cancel() // Stop Kafka consumer
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

	logger.Logger.Info().
		Str("port", port).
		Str("metrics_endpoint", "/metrics").
		Msg("HTTP server started")

	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to start HTTP server")
	}
}

func startGRPCServer(inventoryServer *grpcDelivery.InventoryGRPCServer, port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to listen for gRPC")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterInventoryServiceServer(grpcServer, inventoryServer)

	logger.Logger.Info().
		Str("port", port).
		Msg("gRPC server started")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to serve gRPC")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

