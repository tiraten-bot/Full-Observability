package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/user/domain"
	"github.com/tair/full-observability/pkg/auth"
)

// UpdateUserCommand represents the command to update a user
type UpdateUserCommand struct {
	ID       uint
	Email    string
	FullName string
	Password string // Optional
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
	if cmd.ID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// Check if user exists
	user, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields if provided
	if cmd.Email != "" {
		// Check if email is already taken by another user
		if existingUser, _ := h.repo.FindByEmail(cmd.Email); existingUser != nil && existingUser.ID != cmd.ID {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = cmd.Email
	}
	
	if cmd.FullName != "" {
		user.FullName = cmd.FullName
	}

	if cmd.Password != "" {
		if len(cmd.Password) < 6 {
			return nil, fmt.Errorf("password must be at least 6 characters")
		}
		hashedPassword, err := auth.HashPassword(cmd.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	user.UpdatedAt = time.Now()

	if err := h.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
