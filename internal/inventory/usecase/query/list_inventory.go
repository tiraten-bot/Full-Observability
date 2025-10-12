package query

import (
	"fmt"

	"github.com/tair/full-observability/internal/inventory/domain"
)

// ListInventoryQuery represents the query to list inventories
type ListInventoryQuery struct {
	Limit  int
	Offset int
}

// ListInventoryHandler handles list inventory query
type ListInventoryHandler struct {
	repo domain.InventoryRepository
}

// NewListInventoryHandler creates a new list inventory handler
func NewListInventoryHandler(repo domain.InventoryRepository) *ListInventoryHandler {
	return &ListInventoryHandler{repo: repo}
}

// Handle executes the list inventory query
func (h *ListInventoryHandler) Handle(query ListInventoryQuery) ([]domain.Inventory, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	if query.Limit > 100 {
		query.Limit = 100
	}

	inventories, err := h.repo.FindAll(query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventories: %w", err)
	}

	return inventories, nil
}

