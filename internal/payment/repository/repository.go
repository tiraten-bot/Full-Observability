package repository

import (
	"github.com/tair/full-observability/internal/payment/domain"
	"gorm.io/gorm"
)

type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

func (r *GormPaymentRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&domain.Payment{})
}

func (r *GormPaymentRepository) Create(payment *domain.Payment) error {
	return r.db.Create(payment).Error
}

func (r *GormPaymentRepository) FindByID(id uint) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *GormPaymentRepository) FindByOrderID(orderID string) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *GormPaymentRepository) FindByUserID(userID uint, limit, offset int) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.db.Where("user_id = ?", userID).
		Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *GormPaymentRepository) FindAll(limit, offset int) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.db.Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *GormPaymentRepository) Update(payment *domain.Payment) error {
	return r.db.Save(payment).Error
}

func (r *GormPaymentRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&domain.Payment{}).
		Where("id = ?", id).
		Update("status", status).Error
}
