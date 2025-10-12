package command

import (
	"fmt"
	"time"

	"github.com/tair/full-observability/internal/product/domain"
)

// UpdateProductCommand represents the command to update a product
type UpdateProductCommand struct {
	ID          uint
	Name        string
	Description string
	Price       float64
	Stock       int
	Category    string
	SKU         string
	IsActive    bool
}

// UpdateProductHandler handles product update command
type UpdateProductHandler struct {
	repo domain.ProductRepository
}

// NewUpdateProductHandler creates a new update product handler
func NewUpdateProductHandler(repo domain.ProductRepository) *UpdateProductHandler {
	return &UpdateProductHandler{repo: repo}
}

// Handle executes the update product command
func (h *UpdateProductHandler) Handle(cmd UpdateProductCommand) (*domain.Product, error) {
	// Validation
	if cmd.ID == 0 {
		return nil, fmt.Errorf("invalid product id")
	}

	// Check if product exists
	product, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Update fields if provided
	if cmd.Name != "" {
		product.Name = cmd.Name
	}

	if cmd.Description != "" {
		product.Description = cmd.Description
	}

	if cmd.Price >= 0 {
		product.Price = cmd.Price
	}

	if cmd.Stock >= 0 {
		product.Stock = cmd.Stock
	}

	if cmd.Category != "" {
		product.Category = cmd.Category
	}

	if cmd.SKU != "" {
		// Check if SKU is already taken by another product
		if existingProduct, _ := h.repo.FindBySKU(cmd.SKU); existingProduct != nil && existingProduct.ID != cmd.ID {
			return nil, fmt.Errorf("SKU already exists")
		}
		product.SKU = cmd.SKU
	}

	product.IsActive = cmd.IsActive
	product.UpdatedAt = time.Now()

	if err := h.repo.Update(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

