package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/product/domain"
)

// ListProductsQuery represents the query to list all products
type ListProductsQuery struct {
	Limit    int
	Offset   int
	Category string // Optional: filter by category
}

// ListProductsHandler handles list products query
type ListProductsHandler struct {
	repo domain.ProductRepository
}

// NewListProductsHandler creates a new list products handler
func NewListProductsHandler(repo domain.ProductRepository) *ListProductsHandler {
	return &ListProductsHandler{repo: repo}
}

// Handle executes the list products query
func (h *ListProductsHandler) Handle(query ListProductsQuery) ([]domain.Product, error) {
	var products []domain.Product
	var err error

	// Set defaults
	if query.Limit <= 0 {
		query.Limit = 50
	}

	// Filter by category if specified
	if query.Category != "" {
		products, err = h.repo.FindByCategory(query.Category, query.Limit, query.Offset)
	} else {
		products, err = h.repo.FindAll(query.Limit, query.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}
