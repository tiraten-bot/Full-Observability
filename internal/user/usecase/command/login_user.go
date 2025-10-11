package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/user/domain"
	"github.com/tair/full-observability/pkg/auth"
)

// LoginUserCommand represents the command to login a user
type LoginUserCommand struct {
	Username string
	Password string
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

// LoginUserHandler handles user login command
type LoginUserHandler struct {
	repo domain.UserRepository
}

// NewLoginUserHandler creates a new login user handler
func NewLoginUserHandler(repo domain.UserRepository) *LoginUserHandler {
	return &LoginUserHandler{repo: repo}
}

// Handle executes the login user command
func (h *LoginUserHandler) Handle(cmd LoginUserCommand) (*LoginResponse, error) {
	// Validation
	if cmd.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if cmd.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Find user by username
	user, err := h.repo.FindByUsername(cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Verify password
	if !auth.CheckPassword(user.Password, cmd.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

