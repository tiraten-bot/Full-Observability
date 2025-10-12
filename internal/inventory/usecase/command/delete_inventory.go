package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/inventory/domain"
)

// DeleteInventoryCommand represents the command to delete an inventory
type DeleteInventoryCommand struct {
	ID uint
}

// DeleteInventoryHandler handles delete inventory command
type DeleteInventoryHandler struct {
	repo domain.InventoryRepository
}

// NewDeleteInventoryHandler creates a new delete inventory handler
func NewDeleteInventoryHandler(repo domain.InventoryRepository) *DeleteInventoryHandler {
	return &DeleteInventoryHandler{repo: repo}
}

// Handle executes the delete inventory command
func (h *DeleteInventoryHandler) Handle(cmd DeleteInventoryCommand) error {
	if cmd.ID == 0 {
		return fmt.Errorf("id is required")
	}

	if err := h.repo.Delete(cmd.ID); err != nil {
		return fmt.Errorf("failed to delete inventory: %w", err)
	}

	return nil
}

