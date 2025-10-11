package domain

import (
	"time"

	"gorm.io/gorm"
)

// Role types
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User represents the user entity (domain model)
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // Never expose password in JSON
	FullName  string         `json:"full_name" gorm:"not null"`
	Role      string         `json:"role" gorm:"not null;default:'user'"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UserRepository defines the contract for user data access
type UserRepository interface {
	Create(user *User) error
	FindByID(id uint) (*User, error)
	FindByUsername(username string) (*User, error)
	FindByEmail(email string) (*User, error)
	FindAll(limit, offset int) ([]User, error)
	FindByRole(role string, limit, offset int) ([]User, error)
	Update(user *User) error
	Delete(id uint) error
	Count() (int64, error)
	CountByRole(role string) (int64, error)
}
