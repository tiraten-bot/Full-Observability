package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/product/domain"
)

// UpdateStockCommand represents the command to update product stock
type UpdateStockCommand struct {
	ProductID uint
	Stock     int
}

// UpdateStockHandler handles stock update command
type UpdateStockHandler struct {
	repo domain.ProductRepository
}

// NewUpdateStockHandler creates a new update stock handler
func NewUpdateStockHandler(repo domain.ProductRepository) *UpdateStockHandler {
	return &UpdateStockHandler{repo: repo}
}

// Handle executes the update stock command
func (h *UpdateStockHandler) Handle(cmd UpdateStockCommand) error {
	if cmd.ProductID == 0 {
		return fmt.Errorf("invalid product id")
	}

	if cmd.Stock < 0 {
		return fmt.Errorf("stock cannot be negative")
	}

	// Check if product exists
	if _, err := h.repo.FindByID(cmd.ProductID); err != nil {
		return fmt.Errorf("product not found")
	}

	if err := h.repo.UpdateStock(cmd.ProductID, cmd.Stock); err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

