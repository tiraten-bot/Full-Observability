package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/user/domain"
)

// CreateUserCommand represents the command to create a user
type CreateUserCommand struct {
	Username string
	Email    string
	FullName string
}

// CreateUserHandler handles user creation command
type CreateUserHandler struct {
	repo domain.UserRepository
}

// NewCreateUserHandler creates a new create user handler
func NewCreateUserHandler(repo domain.UserRepository) *CreateUserHandler {
	return &CreateUserHandler{repo: repo}
}

// Handle executes the create user command
func (h *CreateUserHandler) Handle(cmd CreateUserCommand) (*domain.User, error) {
	// Validation
	if cmd.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if cmd.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if cmd.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	user := &domain.User{
		Username:  cmd.Username,
		Email:     cmd.Email,
		FullName:  cmd.FullName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

