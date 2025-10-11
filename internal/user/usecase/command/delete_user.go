package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/user/domain"
)

// DeleteUserCommand represents the command to delete a user
type DeleteUserCommand struct {
	ID uint
}

// DeleteUserHandler handles user deletion command
type DeleteUserHandler struct {
	repo domain.UserRepository
}

// NewDeleteUserHandler creates a new delete user handler
func NewDeleteUserHandler(repo domain.UserRepository) *DeleteUserHandler {
	return &DeleteUserHandler{repo: repo}
}

// Handle executes the delete user command
func (h *DeleteUserHandler) Handle(cmd DeleteUserCommand) error {
	// Validation
	if cmd.ID == 0 {
		return fmt.Errorf("invalid user id")
	}

	if err := h.repo.Delete(cmd.ID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

