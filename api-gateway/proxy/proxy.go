package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tair/full-observability/api-gateway/config"
	"github.com/tair/full-observability/api-gateway/loadbalancer"
	"github.com/tair/full-observability/pkg/logger"
)

// ReverseProxy handles proxying requests to backend services
type ReverseProxy struct {
	config        *config.GatewayConfig
	client        *http.Client
	loadBalancers map[string]*loadbalancer.RoundRobin
	retryConfig   RetryConfig
}

// RetryConfig holds retry configuration for proxy
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// NewReverseProxy creates a new reverse proxy
func NewReverseProxy(cfg *config.GatewayConfig) *ReverseProxy {
	// Initialize load balancers for each service
	loadBalancers := make(map[string]*loadbalancer.RoundRobin)

	for name, svc := range cfg.Services {
		loadBalancers[name] = loadbalancer.NewRoundRobin(svc.Instances)
	}

	return &ReverseProxy{
		config:        cfg,
		loadBalancers: loadBalancers,
		retryConfig: RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     2 * time.Second,
			Multiplier:   2.0,
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest forwards the request to the target service with retry logic
func (p *ReverseProxy) ProxyRequest(c *fiber.Ctx, serviceName string) error {
	// Check if method is idempotent (only retry safe methods)
	isIdempotent := isIdempotentMethod(c.Method())
	maxAttempts := 1
	if isIdempotent {
		maxAttempts = p.retryConfig.MaxAttempts
	}

	// Save request body for retries
	bodyBytes := c.Body()

	var lastErr error
	delay := p.retryConfig.InitialDelay

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Get next server from load balancer
		lb, lbExists := p.loadBalancers[serviceName]
		if !lbExists {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fmt.Sprintf("Load balancer for '%s' not found", serviceName),
			})
		}

		serverURL := lb.Next()
		if serverURL == "" {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": fmt.Sprintf("No available instances for '%s'", serviceName),
			})
		}

		if attempt > 0 {
			logger.Logger.Info().
				Str("service", serviceName).
				Str("target_url", serverURL).
				Int("attempt", attempt+1).
				Msg("Retrying request with different instance")
		} else {
			logger.Logger.Debug().
				Str("service", serviceName).
				Str("target_url", serverURL).
				Str("path", c.Path()).
				Msg("Load balancer selected instance")
		}

		// Build target URL
		targetURL := p.buildTargetURLWithServer(c, serverURL)

		// Create new request
		req, err := http.NewRequest(
			c.Method(),
			targetURL,
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			lastErr = err
			continue
		}

		// Copy headers from original request
		p.copyHeaders(c, req)

		// Execute request
		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = err

			// Retry if not last attempt
			if attempt < maxAttempts-1 {
				logger.Logger.Warn().
					Err(err).
					Str("service", serviceName).
					Int("attempt", attempt+1).
					Dur("delay", delay).
					Msg("Request failed, retrying...")

				time.Sleep(delay)
				delay = time.Duration(float64(delay) * p.retryConfig.Multiplier)
				if delay > p.retryConfig.MaxDelay {
					delay = p.retryConfig.MaxDelay
				}
				continue
			}

			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":   "Failed to reach backend service",
				"service": serviceName,
				"details": err.Error(),
			})
		}
		defer resp.Body.Close()

		// Check if status code is retryable (502, 503, 504)
		if resp.StatusCode >= 502 && resp.StatusCode <= 504 && attempt < maxAttempts-1 {
			logger.Logger.Warn().
				Int("status", resp.StatusCode).
				Str("service", serviceName).
				Int("attempt", attempt+1).
				Dur("delay", delay).
				Msg("Retryable status code, retrying...")

			time.Sleep(delay)
			delay = time.Duration(float64(delay) * p.retryConfig.Multiplier)
			if delay > p.retryConfig.MaxDelay {
				delay = p.retryConfig.MaxDelay
			}
			continue
		}

		// Success - copy response
		p.copyResponseHeaders(c, resp)
		c.Status(resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read response",
			})
		}

		if attempt > 0 {
			logger.Logger.Info().
				Str("service", serviceName).
				Int("attempt", attempt+1).
				Int("status", resp.StatusCode).
				Msg("Request succeeded after retry")
		}

		return c.Send(body)
	}

	// All retries failed
	errorDetails := "Unknown error"
	if lastErr != nil {
		errorDetails = lastErr.Error()
	}

	return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
		"error":   "All retry attempts failed",
		"service": serviceName,
		"details": errorDetails,
	})
}

// isIdempotentMethod checks if HTTP method is safe to retry
func isIdempotentMethod(method string) bool {
	idempotentMethods := map[string]bool{
		"GET":     true,
		"HEAD":    true,
		"PUT":     true,
		"DELETE":  true,
		"OPTIONS": true,
	}
	return idempotentMethods[method]
}

// buildTargetURL constructs the full URL for the backend service
func (p *ReverseProxy) buildTargetURL(c *fiber.Ctx, service config.ServiceConfig) string {
	return p.buildTargetURLWithServer(c, service.BaseURL)
}

// buildTargetURLWithServer constructs the full URL with a specific server
func (p *ReverseProxy) buildTargetURLWithServer(c *fiber.Ctx, serverURL string) string {
	path := string(c.Request().URI().Path())

	// Build query string
	queryString := string(c.Request().URI().QueryString())
	if queryString != "" {
		queryString = "?" + queryString
	}

	return serverURL + path + queryString
}

// GetLoadBalancers returns all load balancers (for stats)
func (p *ReverseProxy) GetLoadBalancers() map[string]*loadbalancer.RoundRobin {
	return p.loadBalancers
}

// copyHeaders copies relevant headers from Fiber context to http.Request
func (p *ReverseProxy) copyHeaders(c *fiber.Ctx, req *http.Request) {
	// Copy all headers except Host
	c.Request().Header.VisitAll(func(key, value []byte) {
		keyStr := string(key)
		// Skip certain headers
		if strings.ToLower(keyStr) == "host" {
			return
		}
		req.Header.Set(keyStr, string(value))
	})

	// Set Content-Type if exists
	if contentType := c.Get("Content-Type"); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Set Authorization if exists
	if auth := c.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	// Set X-Forwarded headers
	req.Header.Set("X-Forwarded-For", c.IP())
	req.Header.Set("X-Forwarded-Proto", c.Protocol())
	req.Header.Set("X-Forwarded-Host", c.Hostname())
}

// copyResponseHeaders copies headers from http.Response to Fiber context
func (p *ReverseProxy) copyResponseHeaders(c *fiber.Ctx, resp *http.Response) {
	for key, values := range resp.Header {
		// Skip certain headers
		if strings.ToLower(key) == "content-length" {
			continue
		}
		for _, value := range values {
			c.Set(key, value)
		}
	}
}
