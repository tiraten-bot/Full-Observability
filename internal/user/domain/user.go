package domain

import "time"

// User represents the user entity (domain model)
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository defines the contract for user data access
type UserRepository interface {
	Create(user *User) error
	FindByID(id int) (*User, error)
	FindAll() ([]User, error)
	Update(user *User) error
	Delete(id int) error
	Count() (int, error)
}

