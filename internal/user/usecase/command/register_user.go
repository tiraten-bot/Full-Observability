package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/user/domain"
	"github.com/tair/full-observability/pkg/auth"
)

// RegisterUserCommand represents the command to register a new user
type RegisterUserCommand struct {
	Username string
	Email    string
	Password string
	FullName string
	Role     string // Optional, defaults to "user"
}

// RegisterUserHandler handles user registration command
type RegisterUserHandler struct {
	repo domain.UserRepository
}

// NewRegisterUserHandler creates a new register user handler
func NewRegisterUserHandler(repo domain.UserRepository) *RegisterUserHandler {
	return &RegisterUserHandler{repo: repo}
}

// Handle executes the register user command
func (h *RegisterUserHandler) Handle(cmd RegisterUserCommand) (*domain.User, error) {
	// Validation
	if cmd.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if cmd.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if cmd.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if len(cmd.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	if cmd.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	// Check if user already exists
	if existingUser, _ := h.repo.FindByUsername(cmd.Username); existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}
	if existingUser, _ := h.repo.FindByEmail(cmd.Email); existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role if not provided
	role := cmd.Role
	if role == "" {
		role = domain.RoleUser
	}
	// Validate role
	if role != domain.RoleUser && role != domain.RoleAdmin {
		return nil, fmt.Errorf("invalid role")
	}

	user := &domain.User{
		Username:  cmd.Username,
		Email:     cmd.Email,
		Password:  hashedPassword,
		FullName:  cmd.FullName,
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
