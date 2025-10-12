package domain

import (
	"time"

	"gorm.io/gorm"
)

// Product represents the product entity
type Product struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null"`
	Stock       int            `json:"stock" gorm:"not null;default:0"`
	Category    string         `json:"category"`
	SKU         string         `json:"sku" gorm:"uniqueIndex"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name
func (Product) TableName() string {
	return "products"
}

// IsAvailable checks if product is in stock
func (p *Product) IsAvailable() bool {
	return p.Stock > 0 && p.IsActive
}

// ProductRepository defines the contract for product data access
type ProductRepository interface {
	Create(product *Product) error
	FindByID(id uint) (*Product, error)
	FindBySKU(sku string) (*Product, error)
	FindAll(limit, offset int) ([]Product, error)
	FindByCategory(category string, limit, offset int) ([]Product, error)
	Update(product *Product) error
	Delete(id uint) error
	Count() (int64, error)
	UpdateStock(id uint, stock int) error
	
	// User favorites
	AddFavorite(userID, productID uint) error
	RemoveFavorite(userID, productID uint) error
	GetUserFavorites(userID uint) ([]uint, error)
	IsFavorite(userID, productID uint) (bool, error)
}

