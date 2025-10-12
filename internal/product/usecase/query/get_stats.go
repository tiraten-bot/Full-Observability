package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/product/domain"
)

// GetStatsQuery represents the query to get product statistics
type GetStatsQuery struct{}

// ProductStats represents product statistics
type ProductStats struct {
	TotalProducts   int64   `json:"total_products"`
	ActiveProducts  int64   `json:"active_products"`
	TotalStock      int64   `json:"total_stock"`
	AveragePrice    float64 `json:"average_price"`
	TotalCategories int64   `json:"total_categories"`
}

// GetStatsHandler handles get stats query
type GetStatsHandler struct {
	repo domain.ProductRepository
}

// NewGetStatsHandler creates a new get stats handler
func NewGetStatsHandler(repo domain.ProductRepository) *GetStatsHandler {
	return &GetStatsHandler{repo: repo}
}

// Handle executes the get stats query
func (h *GetStatsHandler) Handle(query GetStatsQuery) (*ProductStats, error) {
	totalProducts, err := h.repo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to get product count: %w", err)
	}

	// Get all products to calculate stats
	products, err := h.repo.FindAll(10000, 0) // Get all products
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	var activeProducts int64
	var totalStock int64
	var totalPrice float64
	categories := make(map[string]bool)

	for _, product := range products {
		if product.IsActive {
			activeProducts++
		}
		totalStock += int64(product.Stock)
		totalPrice += product.Price
		if product.Category != "" {
			categories[product.Category] = true
		}
	}

	averagePrice := 0.0
	if totalProducts > 0 {
		averagePrice = totalPrice / float64(totalProducts)
	}

	stats := &ProductStats{
		TotalProducts:   totalProducts,
		ActiveProducts:  activeProducts,
		TotalStock:      totalStock,
		AveragePrice:    averagePrice,
		TotalCategories: int64(len(categories)),
	}

	return stats, nil
}

