package config

import (
	"os"
	"strings"
	"time"
)

// ServiceConfig holds configuration for a backend service
type ServiceConfig struct {
	Name        string
	BaseURL     string   // Primary URL (for backward compatibility)
	Instances   []string // Multiple instances for load balancing
	Timeout     time.Duration
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
			"user":      loadServiceConfig("user-service", "USER_SERVICE_URL", "http://localhost:8080"),
			"product":   loadServiceConfig("product-service", "PRODUCT_SERVICE_URL", "http://localhost:8081"),
			"inventory": loadServiceConfig("inventory-service", "INVENTORY_SERVICE_URL", "http://localhost:8082"),
			"payment":   loadServiceConfig("payment-service", "PAYMENT_SERVICE_URL", "http://localhost:8083"),
		},
	}
}

// loadServiceConfig loads configuration for a single service (supports multiple instances)
func loadServiceConfig(name, envKey, defaultURL string) ServiceConfig {
	urlsStr := getEnv(envKey, defaultURL)

	// Split by comma for multiple instances
	urls := strings.Split(urlsStr, ",")
	for i, url := range urls {
		urls[i] = strings.TrimSpace(url)
	}

	return ServiceConfig{
		Name:        name,
		BaseURL:     urls[0], // Primary URL
		Instances:   urls,    // All instances
		Timeout:     30 * time.Second,
		HealthCheck: "/health",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
