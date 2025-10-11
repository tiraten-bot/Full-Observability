package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
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

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("🚀 User service starting on port %s", port)
	log.Printf("📊 Prometheus metrics: http://localhost:%s/metrics", port)
	log.Printf("🔐 Auth endpoints: /auth/register, /auth/login", port)
	log.Printf("👤 User endpoints: /users/me")
	log.Printf("👑 Admin endpoints: /admin/*")

	if err := http.ListenAndServe(":"+port, c.Handler(router)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
