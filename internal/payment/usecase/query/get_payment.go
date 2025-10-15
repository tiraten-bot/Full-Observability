package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/payment/domain"
)

// GetPaymentQuery represents the query to get a payment
type GetPaymentQuery struct {
	ID uint
}

// GetPaymentHandler handles get payment query
type GetPaymentHandler struct {
	repo domain.PaymentRepository
}

// NewGetPaymentHandler creates a new get payment handler
func NewGetPaymentHandler(repo domain.PaymentRepository) *GetPaymentHandler {
	return &GetPaymentHandler{repo: repo}
}

// Handle executes the get payment query
func (h *GetPaymentHandler) Handle(query GetPaymentQuery) (*domain.Payment, error) {
	if query.ID == 0 {
		return nil, fmt.Errorf("id is required")
	}

	payment, err := h.repo.FindByID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	return payment, nil
}
