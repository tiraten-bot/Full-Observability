package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/inventory/domain"
)

// CreateInventoryCommand represents the command to create an inventory
type CreateInventoryCommand struct {
	ProductID uint
	Quantity  int
	Location  string
}

// CreateInventoryHandler handles create inventory command
type CreateInventoryHandler struct {
	repo domain.InventoryRepository
}

// NewCreateInventoryHandler creates a new create inventory handler
func NewCreateInventoryHandler(repo domain.InventoryRepository) *CreateInventoryHandler {
	return &CreateInventoryHandler{repo: repo}
}

// Handle executes the create inventory command
func (h *CreateInventoryHandler) Handle(cmd CreateInventoryCommand) (*domain.Inventory, error) {
	if cmd.ProductID == 0 {
		return nil, fmt.Errorf("product_id is required")
	}

	if cmd.Quantity < 0 {
		return nil, fmt.Errorf("quantity cannot be negative")
	}

	if cmd.Location == "" {
		cmd.Location = "warehouse"
	}

	inventory := &domain.Inventory{
		ProductID: cmd.ProductID,
		Quantity:  cmd.Quantity,
		Location:  cmd.Location,
	}

	if err := h.repo.Create(inventory); err != nil {
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}

	return inventory, nil
}

