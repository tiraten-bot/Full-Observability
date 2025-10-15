package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/product/domain"
)

// GetProductQuery represents the query to get a product by ID
type GetProductQuery struct {
	ID uint
}

// GetProductHandler handles get product query
type GetProductHandler struct {
	repo domain.ProductRepository
}

// NewGetProductHandler creates a new get product handler
func NewGetProductHandler(repo domain.ProductRepository) *GetProductHandler {
	return &GetProductHandler{repo: repo}
}

// Handle executes the get product query
func (h *GetProductHandler) Handle(query GetProductQuery) (*domain.Product, error) {
	if query.ID == 0 {
		return nil, fmt.Errorf("invalid product id")
	}

	product, err := h.repo.FindByID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	return product, nil
}
