package repository

import (
	"github.com/tair/full-observability/internal/inventory/domain"
	"gorm.io/gorm"
)

type GormInventoryRepository struct {
	db *gorm.DB
}

func NewGormInventoryRepository(db *gorm.DB) *GormInventoryRepository {
	return &GormInventoryRepository{db: db}
}

func (r *GormInventoryRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&domain.Inventory{})
}

func (r *GormInventoryRepository) Create(inventory *domain.Inventory) error {
	return r.db.Create(inventory).Error
}

func (r *GormInventoryRepository) FindByID(id uint) (*domain.Inventory, error) {
	var inventory domain.Inventory
	err := r.db.First(&inventory, id).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *GormInventoryRepository) FindByProductID(productID uint) (*domain.Inventory, error) {
	var inventory domain.Inventory
	err := r.db.Where("product_id = ?", productID).First(&inventory).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *GormInventoryRepository) FindAll(limit, offset int) ([]domain.Inventory, error) {
	var inventories []domain.Inventory
	err := r.db.Limit(limit).Offset(offset).Find(&inventories).Error
	return inventories, err
}

func (r *GormInventoryRepository) Update(inventory *domain.Inventory) error {
	return r.db.Save(inventory).Error
}

func (r *GormInventoryRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Inventory{}, id).Error
}

func (r *GormInventoryRepository) UpdateQuantity(productID uint, quantity int) error {
	return r.db.Model(&domain.Inventory{}).
		Where("product_id = ?", productID).
		Update("quantity", quantity).Error
}

