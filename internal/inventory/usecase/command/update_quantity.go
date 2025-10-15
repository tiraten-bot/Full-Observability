package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/inventory/domain"
)

// UpdateQuantityCommand represents the command to update inventory quantity
type UpdateQuantityCommand struct {
	ProductID uint
	Quantity  int
}

// UpdateQuantityHandler handles update quantity command
type UpdateQuantityHandler struct {
	repo domain.InventoryRepository
}

// NewUpdateQuantityHandler creates a new update quantity handler
func NewUpdateQuantityHandler(repo domain.InventoryRepository) *UpdateQuantityHandler {
	return &UpdateQuantityHandler{repo: repo}
}

// Handle executes the update quantity command
func (h *UpdateQuantityHandler) Handle(cmd UpdateQuantityCommand) error {
	if cmd.ProductID == 0 {
		return fmt.Errorf("product_id is required")
	}

	if cmd.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	if err := h.repo.UpdateQuantity(cmd.ProductID, cmd.Quantity); err != nil {
		return fmt.Errorf("failed to update quantity: %w", err)
	}

	return nil
}
