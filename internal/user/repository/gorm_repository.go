package repository

import (
	"fmt"

	"github.com/tair/full-observability/internal/user/domain"
	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository interface using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GORM user repository
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// Create inserts a new user into the database
func (r *GormUserRepository) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByID retrieves a user by ID
func (r *GormUserRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// FindByUsername retrieves a user by username
func (r *GormUserRepository) FindByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// FindByEmail retrieves a user by email
func (r *GormUserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// FindAll retrieves all users with pagination
func (r *GormUserRepository) FindAll(limit, offset int) ([]domain.User, error) {
	var users []domain.User
	query := r.db.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	return users, nil
}

// FindByRole retrieves users by role with pagination
func (r *GormUserRepository) FindByRole(role string, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	query := r.db.Where("role = ?", role).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to find users by role: %w", err)
	}
	return users, nil
}

// Update updates a user's information
func (r *GormUserRepository) Update(user *domain.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete soft deletes a user from the database
func (r *GormUserRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// Count returns the total number of users
func (r *GormUserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// CountByRole returns the number of users by role
func (r *GormUserRepository) CountByRole(role string) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("role = ?", role).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}
	return count, nil
}

// AutoMigrate runs database migrations
func (r *GormUserRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&domain.User{})
}
