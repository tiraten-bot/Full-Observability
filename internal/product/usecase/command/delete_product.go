package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/product/domain"
)

// DeleteProductCommand represents the command to delete a product
type DeleteProductCommand struct {
	ID uint
}

// DeleteProductHandler handles product deletion command
type DeleteProductHandler struct {
	repo domain.ProductRepository
}

// NewDeleteProductHandler creates a new delete product handler
func NewDeleteProductHandler(repo domain.ProductRepository) *DeleteProductHandler {
	return &DeleteProductHandler{repo: repo}
}

// Handle executes the delete product command
func (h *DeleteProductHandler) Handle(cmd DeleteProductCommand) error {
	if cmd.ID == 0 {
		return fmt.Errorf("invalid product id")
	}

	// Check if product exists
	if _, err := h.repo.FindByID(cmd.ID); err != nil {
		return fmt.Errorf("product not found")
	}

	if err := h.repo.Delete(cmd.ID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

