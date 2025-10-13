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
)

// ReverseProxy handles proxying requests to backend services
type ReverseProxy struct {
	config *config.GatewayConfig
	client *http.Client
}

// NewReverseProxy creates a new reverse proxy
func NewReverseProxy(cfg *config.GatewayConfig) *ReverseProxy {
	return &ReverseProxy{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest forwards the request to the target service
func (p *ReverseProxy) ProxyRequest(c *fiber.Ctx, serviceName string) error {
	// Get service configuration
	serviceConfig, exists := p.config.Services[serviceName]
	if !exists {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fmt.Sprintf("Service '%s' not found", serviceName),
		})
	}

	// Build target URL
	targetURL := p.buildTargetURL(c, serviceConfig)

	// Create new request
	req, err := http.NewRequest(
		c.Method(),
		targetURL,
		bytes.NewReader(c.Body()),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create request",
		})
	}

	// Copy headers from original request
	p.copyHeaders(c, req)

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "Failed to reach backend service",
			"service": serviceName,
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	// Copy response headers
	p.copyResponseHeaders(c, resp)

	// Set status code
	c.Status(resp.StatusCode)

	// Copy response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read response",
		})
	}

	return c.Send(body)
}

// buildTargetURL constructs the full URL for the backend service
func (p *ReverseProxy) buildTargetURL(c *fiber.Ctx, service config.ServiceConfig) string {
	// Remove the service prefix from the path
	// e.g., /api/users/* -> /api/users/*
	path := string(c.Request().URI().Path())
	
	// Build query string
	queryString := string(c.Request().URI().QueryString())
	if queryString != "" {
		queryString = "?" + queryString
	}

	return service.BaseURL + path + queryString
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

