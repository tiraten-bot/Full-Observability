package domain

import (
	"time"

	"gorm.io/gorm"
)

// Inventory represents the inventory entity
type Inventory struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProductID uint           `json:"product_id" gorm:"not null;index"`
	Quantity  int            `json:"quantity" gorm:"not null;default:0"`
	Location  string         `json:"location" gorm:"default:'warehouse'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name
func (Inventory) TableName() string {
	return "inventories"
}

// InventoryRepository defines the contract for inventory data access
type InventoryRepository interface {
	Create(inventory *Inventory) error
	FindByID(id uint) (*Inventory, error)
	FindByProductID(productID uint) (*Inventory, error)
	FindAll(limit, offset int) ([]Inventory, error)
	Update(inventory *Inventory) error
	Delete(id uint) error
	UpdateQuantity(productID uint, quantity int) error
}
