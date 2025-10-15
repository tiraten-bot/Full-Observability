package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tair/full-observability/api-gateway/config"
	"github.com/tair/full-observability/pkg/logger"
)

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"` // healthy, unhealthy, unknown
	URL       string        `json:"url"`
	Latency   time.Duration `json:"latency_ms"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// GatewayHealth represents the overall gateway health
type GatewayHealth struct {
	Gateway  string                   `json:"gateway"`
	Status   string                   `json:"status"` // healthy, degraded, unhealthy
	Services map[string]ServiceHealth `json:"services"`
	Uptime   time.Duration            `json:"uptime_seconds"`
}

// HealthChecker checks health of downstream services
type HealthChecker struct {
	config    *config.GatewayConfig
	client    *http.Client
	startTime time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg *config.GatewayConfig) *HealthChecker {
	return &HealthChecker{
		config: cfg,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		startTime: time.Now(),
	}
}

// CheckService checks health of a single service
func (h *HealthChecker) CheckService(ctx context.Context, name string, svc config.ServiceConfig) ServiceHealth {
	start := time.Now()
	healthURL := svc.BaseURL + svc.HealthCheck

	result := ServiceHealth{
		Name:      name,
		URL:       svc.BaseURL,
		Timestamp: time.Now(),
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		result.Status = "unhealthy"
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Latency = time.Since(start)
		return result
	}

	// Execute request
	resp, err := h.client.Do(req)
	if err != nil {
		result.Status = "unhealthy"
		result.Error = fmt.Sprintf("Failed to reach service: %v", err)
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.Latency = time.Since(start)

	// Check status code
	if resp.StatusCode == http.StatusOK {
		result.Status = "healthy"
	} else {
		result.Status = "unhealthy"
		result.Error = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
	}

	return result
}

// CheckAllServices checks health of all downstream services
func (h *HealthChecker) CheckAllServices(ctx context.Context) GatewayHealth {
	services := make(map[string]ServiceHealth)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Check all services concurrently
	for name, svc := range h.config.Services {
		wg.Add(1)
		go func(n string, s config.ServiceConfig) {
			defer wg.Done()
			health := h.CheckService(ctx, n, s)

			mu.Lock()
			services[n] = health
			mu.Unlock()

			// Log service health
			if health.Status == "healthy" {
				logger.Logger.Debug().
					Str("service", n).
					Str("status", health.Status).
					Dur("latency", health.Latency).
					Msg("Service health check")
			} else {
				logger.Logger.Warn().
					Str("service", n).
					Str("status", health.Status).
					Str("error", health.Error).
					Msg("Service health check failed")
			}
		}(name, svc)
	}

	wg.Wait()

	// Determine overall status
	overallStatus := h.determineOverallStatus(services)

	return GatewayHealth{
		Gateway:  "api-gateway",
		Status:   overallStatus,
		Services: services,
		Uptime:   time.Since(h.startTime),
	}
}

// determineOverallStatus determines the overall health status
func (h *HealthChecker) determineOverallStatus(services map[string]ServiceHealth) string {
	healthyCount := 0
	totalCount := len(services)

	for _, svc := range services {
		if svc.Status == "healthy" {
			healthyCount++
		}
	}

	if healthyCount == totalCount {
		return "healthy"
	} else if healthyCount > 0 {
		return "degraded"
	}
	return "unhealthy"
}

// QuickCheck performs a quick health check (just gateway itself)
func (h *HealthChecker) QuickCheck() map[string]interface{} {
	return map[string]interface{}{
		"status":    "healthy",
		"gateway":   "api-gateway",
		"uptime":    time.Since(h.startTime).Seconds(),
		"timestamp": time.Now(),
	}
}
