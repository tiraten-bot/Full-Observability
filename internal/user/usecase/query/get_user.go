package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/user/domain"
)

// GetUserQuery represents the query to get a user by ID
type GetUserQuery struct {
	ID uint
}

// GetUserHandler handles get user query
type GetUserHandler struct {
	repo domain.UserRepository
}

// NewGetUserHandler creates a new get user handler
func NewGetUserHandler(repo domain.UserRepository) *GetUserHandler {
	return &GetUserHandler{repo: repo}
}

// Handle executes the get user query
func (h *GetUserHandler) Handle(query GetUserQuery) (*domain.User, error) {
	if query.ID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	user, err := h.repo.FindByID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
