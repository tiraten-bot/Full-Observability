package command

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tair/full-observability/internal/payment/domain"
)

// CreatePaymentCommand represents the command to create a payment
type CreatePaymentCommand struct {
	UserID        uint
	Amount        float64
	Currency      string
	PaymentMethod string
}

// CreatePaymentHandler handles create payment command
type CreatePaymentHandler struct {
	repo domain.PaymentRepository
}

// NewCreatePaymentHandler creates a new create payment handler
func NewCreatePaymentHandler(repo domain.PaymentRepository) *CreatePaymentHandler {
	return &CreatePaymentHandler{repo: repo}
}

// Handle executes the create payment command
func (h *CreatePaymentHandler) Handle(cmd CreatePaymentCommand) (*domain.Payment, error) {
	if cmd.UserID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	if cmd.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if cmd.Currency == "" {
		cmd.Currency = "USD"
	}

	// Generate unique IDs
	orderID := fmt.Sprintf("ORD-%s", uuid.New().String()[:8])
	transactionID := fmt.Sprintf("TXN-%s", uuid.New().String()[:12])

	payment := &domain.Payment{
		UserID:        cmd.UserID,
		OrderID:       orderID,
		Amount:        cmd.Amount,
		Currency:      cmd.Currency,
		Status:        domain.StatusPending,
		PaymentMethod: cmd.PaymentMethod,
		TransactionID: transactionID,
	}

	if err := h.repo.Create(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}
