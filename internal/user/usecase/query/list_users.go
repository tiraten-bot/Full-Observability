package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/user/domain"
)

// ListUsersQuery represents the query to list all users
type ListUsersQuery struct {
	Limit  int
	Offset int
	Role   string // Optional: filter by role
}

// ListUsersHandler handles list users query
type ListUsersHandler struct {
	repo domain.UserRepository
}

// NewListUsersHandler creates a new list users handler
func NewListUsersHandler(repo domain.UserRepository) *ListUsersHandler {
	return &ListUsersHandler{repo: repo}
}

// Handle executes the list users query
func (h *ListUsersHandler) Handle(query ListUsersQuery) ([]domain.User, error) {
	var users []domain.User
	var err error

	// Set defaults
	if query.Limit <= 0 {
		query.Limit = 50
	}

	// Filter by role if specified
	if query.Role != "" {
		users, err = h.repo.FindByRole(query.Role, query.Limit, query.Offset)
	} else {
		users, err = h.repo.FindAll(query.Limit, query.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
