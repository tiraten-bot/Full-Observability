package command

import (
	"fmt"

	"github.com/tair/full-observability/internal/payment/domain"
)

// UpdateStatusCommand represents the command to update payment status
type UpdateStatusCommand struct {
	PaymentID uint
	Status    string
}

// UpdateStatusHandler handles update status command
type UpdateStatusHandler struct {
	repo domain.PaymentRepository
}

// NewUpdateStatusHandler creates a new update status handler
func NewUpdateStatusHandler(repo domain.PaymentRepository) *UpdateStatusHandler {
	return &UpdateStatusHandler{repo: repo}
}

// Handle executes the update status command
func (h *UpdateStatusHandler) Handle(cmd UpdateStatusCommand) error {
	if cmd.PaymentID == 0 {
		return fmt.Errorf("payment_id is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		domain.StatusPending:   true,
		domain.StatusCompleted: true,
		domain.StatusFailed:    true,
		domain.StatusRefunded:  true,
	}

	if !validStatuses[cmd.Status] {
		return fmt.Errorf("invalid status: %s", cmd.Status)
	}

	if err := h.repo.UpdateStatus(cmd.PaymentID, cmd.Status); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}
