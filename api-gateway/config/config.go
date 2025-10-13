package config

import (
	"os"
	"time"
)

// ServiceConfig holds configuration for a backend service
type ServiceConfig struct {
	Name     string
	BaseURL  string
	Timeout  time.Duration
	HealthCheck string
}

// GatewayConfig holds the main gateway configuration
type GatewayConfig struct {
	Port     string
	Services map[string]ServiceConfig
}

// LoadConfig loads the gateway configuration
func LoadConfig() *GatewayConfig {
	return &GatewayConfig{
		Port: getEnv("GATEWAY_PORT", "8000"),
		Services: map[string]ServiceConfig{
			"user": {
				Name:        "user-service",
				BaseURL:     getEnv("USER_SERVICE_URL", "http://localhost:8080"),
				Timeout:     30 * time.Second,
				HealthCheck: "/health",
			},
			"product": {
				Name:        "product-service",
				BaseURL:     getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
				Timeout:     30 * time.Second,
				HealthCheck: "/health",
			},
			"inventory": {
				Name:        "inventory-service",
				BaseURL:     getEnv("INVENTORY_SERVICE_URL", "http://localhost:8082"),
				Timeout:     30 * time.Second,
				HealthCheck: "/health",
			},
			"payment": {
				Name:        "payment-service",
				BaseURL:     getEnv("PAYMENT_SERVICE_URL", "http://localhost:8083"),
				Timeout:     30 * time.Second,
				HealthCheck: "/health",
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

