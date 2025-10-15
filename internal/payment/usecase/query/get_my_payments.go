package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/payment/domain"
)

// GetMyPaymentsQuery represents the query to get user's own payments
type GetMyPaymentsQuery struct {
	UserID uint
	Limit  int
	Offset int
}

// GetMyPaymentsHandler handles get my payments query
type GetMyPaymentsHandler struct {
	repo domain.PaymentRepository
}

// NewGetMyPaymentsHandler creates a new get my payments handler
func NewGetMyPaymentsHandler(repo domain.PaymentRepository) *GetMyPaymentsHandler {
	return &GetMyPaymentsHandler{repo: repo}
}

// Handle executes the get my payments query
func (h *GetMyPaymentsHandler) Handle(query GetMyPaymentsQuery) ([]domain.Payment, error) {
	if query.UserID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	if query.Limit > 100 {
		query.Limit = 100
	}

	payments, err := h.repo.FindByUserID(query.UserID, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user payments: %w", err)
	}

	return payments, nil
}
