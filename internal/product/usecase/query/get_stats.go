package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/product/domain"
)

// GetStatsQuery represents the query to get product statistics
type GetStatsQuery struct{}

// ProductStats represents product statistics
type ProductStats struct {
	TotalProducts      int64            `json:"total_products"`
	ActiveProducts     int64            `json:"active_products"`
	OutOfStock         int64            `json:"out_of_stock"`
	LowStock           int64            `json:"low_stock"`
	TotalStock         int64            `json:"total_stock"`
	AveragePrice       float64          `json:"average_price"`
	TotalCategories    int64            `json:"total_categories"`
	ProductsByCategory map[string]int64 `json:"products_by_category"`
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
	var outOfStock int64
	var lowStock int64
	var totalStock int64
	var totalPrice float64
	categories := make(map[string]bool)
	productsByCategory := make(map[string]int64)

	const lowStockThreshold = 10

	for _, product := range products {
		if product.IsActive {
			activeProducts++
		}
		if product.Stock == 0 {
			outOfStock++
		}
		if product.Stock > 0 && product.Stock <= lowStockThreshold {
			lowStock++
		}
		totalStock += int64(product.Stock)
		totalPrice += product.Price
		if product.Category != "" {
			categories[product.Category] = true
			productsByCategory[product.Category]++
		}
	}

	averagePrice := 0.0
	if totalProducts > 0 {
		averagePrice = totalPrice / float64(totalProducts)
	}

	stats := &ProductStats{
		TotalProducts:      totalProducts,
		ActiveProducts:     activeProducts,
		OutOfStock:         outOfStock,
		LowStock:           lowStock,
		TotalStock:         totalStock,
		AveragePrice:       averagePrice,
		TotalCategories:    int64(len(categories)),
		ProductsByCategory: productsByCategory,
	}

	return stats, nil
}

