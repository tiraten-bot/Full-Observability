package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/user/domain"
)

// ToggleActiveCommand represents the command to activate/deactivate user (admin only)
type ToggleActiveCommand struct {
	UserID   uint
	IsActive bool
}

// ToggleActiveHandler handles user activation toggle command
type ToggleActiveHandler struct {
	repo domain.UserRepository
}

// NewToggleActiveHandler creates a new toggle active handler
func NewToggleActiveHandler(repo domain.UserRepository) *ToggleActiveHandler {
	return &ToggleActiveHandler{repo: repo}
}

// Handle executes the toggle active command
func (h *ToggleActiveHandler) Handle(cmd ToggleActiveCommand) (*domain.User, error) {
	// Validation
	if cmd.UserID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// Find user
	user, err := h.repo.FindByID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update active status
	user.IsActive = cmd.IsActive
	user.UpdatedAt = time.Now()

	if err := h.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	return user, nil
}
