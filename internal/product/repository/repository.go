package repository

import (
	"github.com/tair/full-observability/internal/product/domain"
	"gorm.io/gorm"
)

type GormProductRepository struct {
	db *gorm.DB
}

func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

func (r *GormProductRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&domain.Product{})
}

func (r *GormProductRepository) Create(product *domain.Product) error {
	return r.db.Create(product).Error
}

func (r *GormProductRepository) FindByID(id uint) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *GormProductRepository) FindBySKU(sku string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *GormProductRepository) FindAll(limit, offset int) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

func (r *GormProductRepository) FindByCategory(category string, limit, offset int) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Where("category = ?", category).Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

func (r *GormProductRepository) Update(product *domain.Product) error {
	return r.db.Save(product).Error
}

func (r *GormProductRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Product{}, id).Error
}

func (r *GormProductRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&domain.Product{}).Count(&count).Error
	return count, err
}

func (r *GormProductRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&domain.Product{}).Where("id = ?", id).Update("stock", stock).Error
}

