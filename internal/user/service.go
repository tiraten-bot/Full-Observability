package user

import (
	"fmt"
)

// Service handles business logic for users
type Service struct {
	repo *Repository
}

// NewService creates a new user service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateUser creates a new user
func (s *Service) CreateUser(req CreateUserRequest) (*User, error) {
	// Validate request
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	return s.repo.Create(req)
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	return s.repo.GetByID(id)
}

// GetAllUsers retrieves all users
func (s *Service) GetAllUsers() ([]User, error) {
	return s.repo.GetAll()
}

// UpdateUser updates a user's information
func (s *Service) UpdateUser(id int, req UpdateUserRequest) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	return s.repo.Update(id, req)
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id")
	}
	return s.repo.Delete(id)
}

