package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tair/full-observability/pkg/logger"
)

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed   CircuitState = "closed"    // Normal operation
	StateOpen     CircuitState = "open"      // Blocking requests
	StateHalfOpen CircuitState = "half-open" // Testing if service recovered
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name            string
	maxFailures     int           // Max consecutive failures before opening
	timeout         time.Duration // Time to wait before attempting recovery
	resetTimeout    time.Duration // Time to wait in half-open before closing
	state           CircuitState
	failures        int
	lastFailureTime time.Time
	lastStateChange time.Time
	successCount    int // Success count in half-open state
	mu              sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:            name,
		maxFailures:     maxFailures,
		timeout:         timeout,
		resetTimeout:    10 * time.Second,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Call executes the function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	
	// Check if we should transition to half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			logger.Logger.Info().
				Str("circuit", cb.name).
				Msg("Circuit breaker transitioning to half-open")
		}
	}

	currentState := cb.state
	cb.mu.Unlock()

	// If circuit is open, reject immediately
	if currentState == StateOpen {
		return fmt.Errorf("circuit breaker is open for %s", cb.name)
	}

	// Execute the function
	err := fn()

	// Record success or failure
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}

	return err
}

// onFailure records a failure
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		// Any failure in half-open state reopens the circuit
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		logger.Logger.Warn().
			Str("circuit", cb.name).
			Msg("Circuit breaker reopened after half-open failure")
	} else if cb.failures >= cb.maxFailures {
		// Too many failures, open the circuit
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		logger.Logger.Error().
			Str("circuit", cb.name).
			Int("failures", cb.failures).
			Int("threshold", cb.maxFailures).
			Msg("Circuit breaker opened")
	}
}

// onSuccess records a success
func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		cb.successCount++
		// After successful requests in half-open, close the circuit
		if cb.successCount >= 3 {
			cb.state = StateClosed
			cb.failures = 0
			cb.successCount = 0
			cb.lastStateChange = time.Now()
			logger.Logger.Info().
				Str("circuit", cb.name).
				Msg("Circuit breaker closed after successful recovery")
		}
	} else if cb.state == StateClosed {
		// Reset failure count on success
		cb.failures = 0
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":               cb.name,
		"state":              cb.state,
		"failures":           cb.failures,
		"max_failures":       cb.maxFailures,
		"last_failure_time":  cb.lastFailureTime,
		"last_state_change":  cb.lastStateChange,
		"time_since_change":  time.Since(cb.lastStateChange).Seconds(),
	}
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetOrCreate gets or creates a circuit breaker for a service
func (m *CircuitBreakerManager) GetOrCreate(serviceName string) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cb, exists := m.breakers[serviceName]; exists {
		return cb
	}

	cb := NewCircuitBreaker(serviceName, 5, 30*time.Second)
	m.breakers[serviceName] = cb
	
	logger.Logger.Info().
		Str("service", serviceName).
		Msg("Circuit breaker created")

	return cb
}

// GetAllStats returns stats for all circuit breakers
func (m *CircuitBreakerManager) GetAllStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	for name, cb := range m.breakers {
		stats[name] = cb.GetStats()
	}
	return stats
}

// CircuitBreakerMiddleware creates a circuit breaker middleware
func CircuitBreakerMiddleware(manager *CircuitBreakerManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Determine service from path
		serviceName := determineServiceFromPath(c.Path())
		if serviceName == "" {
			return c.Next()
		}

		// Get circuit breaker for this service
		cb := manager.GetOrCreate(serviceName)

		// Check circuit state
		if cb.GetState() == StateOpen {
			logger.Logger.Warn().
				Str("service", serviceName).
				Str("path", c.Path()).
				Msg("Circuit breaker is open - request blocked")

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   "Service temporarily unavailable",
				"service": serviceName,
				"message": "Circuit breaker is open. Service is experiencing issues.",
				"retry_after": 30,
			})
		}

		// Execute request with circuit breaker
		var responseErr error
		err := cb.Call(func() error {
			responseErr = c.Next()
			
			// Check if downstream service failed
			if c.Response().StatusCode() >= 500 {
				return fmt.Errorf("downstream service error: %d", c.Response().StatusCode())
			}
			
			return nil
		})

		// If circuit breaker rejected the call
		if err != nil && responseErr == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return responseErr
	}
}

// determineServiceFromPath extracts service name from request path
func determineServiceFromPath(path string) string {
	if len(path) < 5 {
		return ""
	}

	// Map paths to services
	switch {
	case len(path) >= 10 && path[:10] == "/api/users":
		return "user"
	case len(path) >= 13 && path[:13] == "/api/products":
		return "product"
	case len(path) >= 14 && path[:14] == "/api/inventory":
		return "inventory"
	case len(path) >= 13 && path[:13] == "/api/payments":
		return "payment"
	case len(path) >= 5 && path[:5] == "/auth":
		return "user"
	default:
		return ""
	}
}

