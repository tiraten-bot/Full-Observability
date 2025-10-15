package domain

import "time"

// UserFavorite represents a user's favorite product
type UserFavorite struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name
func (UserFavorite) TableName() string {
	return "user_favorites"
}

// UserFavoriteRepository defines the contract for user favorites data access
type UserFavoriteRepository interface {
	AddFavorite(userID, productID uint) error
	RemoveFavorite(userID, productID uint) error
	GetUserFavorites(userID uint) ([]uint, error)
	IsFavorite(userID, productID uint) (bool, error)
}
