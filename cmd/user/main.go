package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/tair/full-observability/api/proto/user"
	grpcDelivery "github.com/tair/full-observability/internal/user/delivery/grpc"
	httpDelivery "github.com/tair/full-observability/internal/user/delivery/http"
	"github.com/tair/full-observability/internal/user/repository"
	"github.com/tair/full-observability/pkg/database"
)

func main() {
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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get underlying *sql.DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Initialize repository
	repo := repository.NewGormUserRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start HTTP server in a goroutine
	httpPort := getEnv("HTTP_PORT", "8080")
	go startHTTPServer(repo, httpPort)

	// Start gRPC server in a goroutine
	grpcPort := getEnv("GRPC_PORT", "9090")
	go startGRPCServer(repo, grpcPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
}

func startHTTPServer(repo *repository.GormUserRepository, port string) {
	// Initialize HTTP handler
	handler := httpDelivery.NewUserHandler(repo)

	// Setup router
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	log.Printf("🌐 HTTP server starting on port %s", port)
	log.Printf("📊 Prometheus metrics: http://localhost:%s/metrics", port)
	log.Printf("🔐 Auth endpoints: /auth/register, /auth/login")
	log.Printf("👤 User endpoints: /users/me")
	log.Printf("👑 Admin endpoints: /admin/*")

	if err := http.ListenAndServe(":"+port, c.Handler(router)); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startGRPCServer(repo *repository.GormUserRepository, port string) {
	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
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
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("🚀 gRPC server starting on port %s", port)
	log.Printf("📡 gRPC reflection enabled")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
