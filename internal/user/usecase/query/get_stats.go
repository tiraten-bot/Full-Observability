package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/user/domain"
)

// GetStatsQuery represents the query to get user statistics (admin only)
type GetStatsQuery struct{}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers  int64 `json:"total_users"`
	AdminCount  int64 `json:"admin_count"`
	UserCount   int64 `json:"user_count"`
	ActiveUsers int64 `json:"active_users"`
}

// GetStatsHandler handles get stats query
type GetStatsHandler struct {
	repo domain.UserRepository
}

// NewGetStatsHandler creates a new get stats handler
func NewGetStatsHandler(repo domain.UserRepository) *GetStatsHandler {
	return &GetStatsHandler{repo: repo}
}

// Handle executes the get stats query
func (h *GetStatsHandler) Handle(query GetStatsQuery) (*UserStats, error) {
	totalUsers, err := h.repo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	adminCount, err := h.repo.CountByRole(domain.RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to count admins: %w", err)
	}

	userCount, err := h.repo.CountByRole(domain.RoleUser)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	return &UserStats{
		TotalUsers:  totalUsers,
		AdminCount:  adminCount,
		UserCount:   userCount,
		ActiveUsers: totalUsers, // TODO: Add active users query
	}, nil
}

