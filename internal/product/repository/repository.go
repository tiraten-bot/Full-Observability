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
	return r.db.AutoMigrate(&domain.Product{}, &domain.UserFavorite{})
}

// User Favorite methods
func (r *GormProductRepository) AddFavorite(userID, productID uint) error {
	favorite := &domain.UserFavorite{
		UserID:    userID,
		ProductID: productID,
	}
	return r.db.Create(favorite).Error
}

func (r *GormProductRepository) RemoveFavorite(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&domain.UserFavorite{}).Error
}

func (r *GormProductRepository) GetUserFavorites(userID uint) ([]uint, error) {
	var favorites []domain.UserFavorite
	err := r.db.Where("user_id = ?", userID).Find(&favorites).Error
	if err != nil {
		return nil, err
	}

	productIDs := make([]uint, len(favorites))
	for i, fav := range favorites {
		productIDs[i] = fav.ProductID
	}
	return productIDs, nil
}

func (r *GormProductRepository) IsFavorite(userID, productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&domain.UserFavorite{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
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

