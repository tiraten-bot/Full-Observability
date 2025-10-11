package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/user/domain"
)

// ListUsersQuery represents the query to list all users
type ListUsersQuery struct{}

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
	users, err := h.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

