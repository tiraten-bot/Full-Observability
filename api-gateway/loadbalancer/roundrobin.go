package loadbalancer

import (
	"sync"

	"github.com/tair/full-observability/pkg/logger"
)

// RoundRobin implements round-robin load balancing
type RoundRobin struct {
	servers []string
	current int
	mu      sync.Mutex
}

// NewRoundRobin creates a new round-robin load balancer
func NewRoundRobin(servers []string) *RoundRobin {
	if len(servers) == 0 {
		servers = []string{"http://localhost:8080"} // Default fallback
	}

	logger.Logger.Info().
		Int("server_count", len(servers)).
		Strs("servers", servers).
		Msg("Round-robin load balancer initialized")

	return &RoundRobin{
		servers: servers,
		current: 0,
	}
}

// Next returns the next server in round-robin order
func (rr *RoundRobin) Next() string {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.servers) == 0 {
		return ""
	}

	server := rr.servers[rr.current]
	rr.current = (rr.current + 1) % len(rr.servers)

	return server
}

// GetServers returns all available servers
func (rr *RoundRobin) GetServers() []string {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	return append([]string{}, rr.servers...)
}

// AddServer adds a new server to the pool
func (rr *RoundRobin) AddServer(server string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	rr.servers = append(rr.servers, server)
	logger.Logger.Info().
		Str("server", server).
		Int("total_servers", len(rr.servers)).
		Msg("Server added to load balancer")
}

// RemoveServer removes a server from the pool
func (rr *RoundRobin) RemoveServer(server string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	for i, s := range rr.servers {
		if s == server {
			rr.servers = append(rr.servers[:i], rr.servers[i+1:]...)
			logger.Logger.Info().
				Str("server", server).
				Int("total_servers", len(rr.servers)).
				Msg("Server removed from load balancer")
			break
		}
	}

	// Reset current index if needed
	if rr.current >= len(rr.servers) && len(rr.servers) > 0 {
		rr.current = 0
	}
}

// GetStats returns load balancer statistics
func (rr *RoundRobin) GetStats() map[string]interface{} {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	return map[string]interface{}{
		"algorithm":     "round-robin",
		"server_count":  len(rr.servers),
		"servers":       rr.servers,
		"current_index": rr.current,
	}
}

