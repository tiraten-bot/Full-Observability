package domain

import (
	"time"

	"gorm.io/gorm"
)

// Payment represents the payment entity
type Payment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	OrderID       string         `json:"order_id" gorm:"not null;uniqueIndex"`
	Amount        float64        `json:"amount" gorm:"not null"`
	Currency      string         `json:"currency" gorm:"default:'USD'"`
	Status        string         `json:"status" gorm:"default:'pending'"` // pending, completed, failed, refunded
	PaymentMethod string         `json:"payment_method"`                  // credit_card, debit_card, paypal, etc.
	TransactionID string         `json:"transaction_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name
func (Payment) TableName() string {
	return "payments"
}

// Payment statuses
const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusRefunded  = "refunded"
)

// PaymentRepository defines the contract for payment data access
type PaymentRepository interface {
	Create(payment *Payment) error
	FindByID(id uint) (*Payment, error)
	FindByOrderID(orderID string) (*Payment, error)
	FindByUserID(userID uint, limit, offset int) ([]Payment, error)
	FindAll(limit, offset int) ([]Payment, error)
	Update(payment *Payment) error
	UpdateStatus(id uint, status string) error
}
