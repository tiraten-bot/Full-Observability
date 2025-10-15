package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/inventory/domain"
)

// GetInventoryQuery represents the query to get an inventory
type GetInventoryQuery struct {
	ID uint
}

// GetInventoryHandler handles get inventory query
type GetInventoryHandler struct {
	repo domain.InventoryRepository
}

// NewGetInventoryHandler creates a new get inventory handler
func NewGetInventoryHandler(repo domain.InventoryRepository) *GetInventoryHandler {
	return &GetInventoryHandler{repo: repo}
}

// Handle executes the get inventory query
func (h *GetInventoryHandler) Handle(query GetInventoryQuery) (*domain.Inventory, error) {
	if query.ID == 0 {
		return nil, fmt.Errorf("id is required")
	}

	inventory, err := h.repo.FindByID(query.ID)
	if err != nil {
		return nil, fmt.Errorf("inventory not found: %w", err)
	}

	return inventory, nil
}
