package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/user/domain"
)

// ChangeRoleCommand represents the command to change user role (admin only)
type ChangeRoleCommand struct {
	UserID uint
	Role   string
}

// ChangeRoleHandler handles user role change command
type ChangeRoleHandler struct {
	repo domain.UserRepository
}

// NewChangeRoleHandler creates a new change role handler
func NewChangeRoleHandler(repo domain.UserRepository) *ChangeRoleHandler {
	return &ChangeRoleHandler{repo: repo}
}

// Handle executes the change role command
func (h *ChangeRoleHandler) Handle(cmd ChangeRoleCommand) (*domain.User, error) {
	// Validation
	if cmd.UserID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if cmd.Role != domain.RoleUser && cmd.Role != domain.RoleAdmin {
		return nil, fmt.Errorf("invalid role")
	}

	// Find user
	user, err := h.repo.FindByID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update role
	user.Role = cmd.Role
	user.UpdatedAt = time.Now()

	if err := h.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	return user, nil
}

