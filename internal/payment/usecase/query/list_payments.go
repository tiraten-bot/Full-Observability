package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/payment/domain"
)

// ListPaymentsQuery represents the query to list payments
type ListPaymentsQuery struct {
	Limit  int
	Offset int
}

// ListPaymentsHandler handles list payments query
type ListPaymentsHandler struct {
	repo domain.PaymentRepository
}

// NewListPaymentsHandler creates a new list payments handler
func NewListPaymentsHandler(repo domain.PaymentRepository) *ListPaymentsHandler {
	return &ListPaymentsHandler{repo: repo}
}

// Handle executes the list payments query
func (h *ListPaymentsHandler) Handle(query ListPaymentsQuery) ([]domain.Payment, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	if query.Limit > 100 {
		query.Limit = 100
	}

	payments, err := h.repo.FindAll(query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	return payments, nil
}

