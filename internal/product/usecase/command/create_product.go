package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/product/domain"
)

// CreateProductCommand represents the command to create a new product
type CreateProductCommand struct {
	Name        string
	Description string
	Price       float64
	Stock       int
	Category    string
	SKU         string
	IsActive    bool
}

// CreateProductHandler handles product creation command
type CreateProductHandler struct {
	repo domain.ProductRepository
}

// NewCreateProductHandler creates a new create product handler
func NewCreateProductHandler(repo domain.ProductRepository) *CreateProductHandler {
	return &CreateProductHandler{repo: repo}
}

// Handle executes the create product command
func (h *CreateProductHandler) Handle(cmd CreateProductCommand) (*domain.Product, error) {
	// Validation
	if cmd.Name == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if cmd.Price < 0 {
		return nil, fmt.Errorf("price cannot be negative")
	}
	if cmd.Stock < 0 {
		return nil, fmt.Errorf("stock cannot be negative")
	}
	if cmd.SKU == "" {
		return nil, fmt.Errorf("SKU is required")
	}

	// Check if SKU already exists
	if existingProduct, _ := h.repo.FindBySKU(cmd.SKU); existingProduct != nil {
		return nil, fmt.Errorf("SKU already exists")
	}

	product := &domain.Product{
		Name:        cmd.Name,
		Description: cmd.Description,
		Price:       cmd.Price,
		Stock:       cmd.Stock,
		Category:    cmd.Category,
		SKU:         cmd.SKU,
		IsActive:    cmd.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}
