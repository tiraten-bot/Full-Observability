package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/user/domain"
)

// UpdateUserCommand represents the command to update a user
type UpdateUserCommand struct {
	ID       int
	Email    string
	FullName string
}

// UpdateUserHandler handles user update command
type UpdateUserHandler struct {
	repo domain.UserRepository
}

// NewUpdateUserHandler creates a new update user handler
func NewUpdateUserHandler(repo domain.UserRepository) *UpdateUserHandler {
	return &UpdateUserHandler{repo: repo}
}

// Handle executes the update user command
func (h *UpdateUserHandler) Handle(cmd UpdateUserCommand) (*domain.User, error) {
	// Validation
	if cmd.ID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if cmd.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if cmd.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	// Check if user exists
	user, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	user.Email = cmd.Email
	user.FullName = cmd.FullName
	user.UpdatedAt = time.Now()

	if err := h.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

