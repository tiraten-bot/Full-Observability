package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/tair/full-observability/api/proto/user"
	_ "github.com/tair/full-observability/cmd/user/docs"
	grpcDelivery "github.com/tair/full-observability/internal/user/delivery/grpc"
	httpDelivery "github.com/tair/full-observability/internal/user/delivery/http"
	"github.com/tair/full-observability/internal/user/repository"
	"github.com/tair/full-observability/pkg/database"
	"github.com/tair/full-observability/pkg/logger"
	"github.com/tair/full-observability/pkg/tracing"
)

func main() {
	// Initialize logger
	serviceName := getEnv("OTEL_SERVICE_NAME", "user-service")
	isDevelopment := getEnv("ENVIRONMENT", "development") == "development"
	logger.Init(serviceName, isDevelopment)
	
	// Set log level from environment
	logLevel := getEnv("LOG_LEVEL", "info")
	logger.SetLevel(logLevel)

	logger.Logger.Info().
		Str("service", serviceName).
		Str("environment", getEnv("ENVIRONMENT", "development")).
		Str("log_level", logLevel).
		Msg("Starting user service")

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

	// Load configuration from environment variables
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "userdb"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to database with GORM
	db, err := database.NewGormConnection(dbConfig)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Get underlying *sql.DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to get database instance")
	}
	defer sqlDB.Close()

	// Initialize repository
	repo := repository.NewGormUserRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to run migrations")
	}

	logger.Logger.Info().Msg("Database initialized successfully")

	// Start HTTP server in a goroutine
	httpPort := getEnv("HTTP_PORT", "8080")
	go startHTTPServer(repo, sqlDB, httpPort)

	// Start gRPC server in a goroutine
	grpcPort := getEnv("GRPC_PORT", "9090")
	go startGRPCServer(repo, grpcPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info().Msg("Shutting down servers...")
}

func startHTTPServer(repo *repository.GormUserRepository, db *sql.DB, port string) {
	// Initialize HTTP handler
	handler := httpDelivery.NewUserHandler(repo)

	// Setup router
	router := mux.NewRouter()

	// Add middlewares (order matters!)
	router.Use(httpDelivery.LoggingMiddleware)  // Logging first
	router.Use(func(next http.Handler) http.Handler {
		return httpDelivery.TracingMiddleware("http-request", next)
	})

	handler.RegisterRoutes(router)
	
	// Health check endpoint
	handler.RegisterHealthCheck(router, db)

	// Swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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

func startGRPCServer(repo *repository.GormUserRepository, port string) {
	// Create gRPC server with interceptors (including tracing)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcDelivery.TracingInterceptor,  // Add tracing first
			grpcDelivery.LoggingInterceptor,
			grpcDelivery.AuthInterceptor,
		),
	)

	// Register user service
	userServer := grpcDelivery.NewUserServer(repo)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	// Register reflection service (for grpcurl and grpc tools)
	reflection.Register(grpcServer)

	// Listen on TCP port
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Logger.Fatal().
			Str("port", port).
			Err(err).
			Msg("Failed to listen on gRPC port")
	}

	logger.Logger.Info().
		Str("port", port).
		Bool("reflection_enabled", true).
		Bool("tracing_enabled", true).
		Msg("gRPC server started")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to start gRPC server")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
