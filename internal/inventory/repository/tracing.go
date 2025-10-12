package repository

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/tair/full-observability/internal/inventory/domain"
)

var tracer = otel.Tracer("inventory-repository")

// GormInventoryRepositoryWithTracing wraps GormInventoryRepository with tracing
type GormInventoryRepositoryWithTracing struct {
	*GormInventoryRepository
}

// NewGormInventoryRepositoryWithTracing creates a new repository with tracing
func NewGormInventoryRepositoryWithTracing(db *gorm.DB) *GormInventoryRepositoryWithTracing {
	return &GormInventoryRepositoryWithTracing{
		GormInventoryRepository: NewGormInventoryRepository(db),
	}
}

// Create with tracing
func (r *GormInventoryRepositoryWithTracing) CreateWithContext(ctx context.Context, inventory *domain.Inventory) error {
	_, span := tracer.Start(ctx, "repository.Create",
		trace.WithAttributes(
			attribute.Int("inventory.product_id", int(inventory.ProductID)),
			attribute.Int("inventory.quantity", inventory.Quantity),
			attribute.String("inventory.location", inventory.Location),
		),
	)
	defer span.End()

	err := r.GormInventoryRepository.Create(inventory)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Int("inventory.id", int(inventory.ID)))
	return nil
}

// FindByID with tracing
func (r *GormInventoryRepositoryWithTracing) FindByIDWithContext(ctx context.Context, id uint) (*domain.Inventory, error) {
	_, span := tracer.Start(ctx, "repository.FindByID",
		trace.WithAttributes(
			attribute.Int("inventory.id", int(id)),
		),
	)
	defer span.End()

	inventory, err := r.GormInventoryRepository.FindByID(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("inventory.product_id", int(inventory.ProductID)),
		attribute.Int("inventory.quantity", inventory.Quantity),
		attribute.String("inventory.location", inventory.Location),
	)
	return inventory, nil
}

// FindByProductID with tracing
func (r *GormInventoryRepositoryWithTracing) FindByProductIDWithContext(ctx context.Context, productID uint) (*domain.Inventory, error) {
	_, span := tracer.Start(ctx, "repository.FindByProductID",
		trace.WithAttributes(
			attribute.Int("inventory.product_id", int(productID)),
		),
	)
	defer span.End()

	inventory, err := r.GormInventoryRepository.FindByProductID(productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("inventory.id", int(inventory.ID)),
		attribute.Int("inventory.quantity", inventory.Quantity),
		attribute.String("inventory.location", inventory.Location),
	)
	return inventory, nil
}

// FindAll with tracing
func (r *GormInventoryRepositoryWithTracing) FindAllWithContext(ctx context.Context, limit, offset int) ([]domain.Inventory, error) {
	_, span := tracer.Start(ctx, "repository.FindAll",
		trace.WithAttributes(
			attribute.Int("query.limit", limit),
			attribute.Int("query.offset", offset),
		),
	)
	defer span.End()

	inventories, err := r.GormInventoryRepository.FindAll(limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("result.count", len(inventories)))
	return inventories, nil
}

// Update with tracing
func (r *GormInventoryRepositoryWithTracing) UpdateWithContext(ctx context.Context, inventory *domain.Inventory) error {
	_, span := tracer.Start(ctx, "repository.Update",
		trace.WithAttributes(
			attribute.Int("inventory.id", int(inventory.ID)),
			attribute.Int("inventory.product_id", int(inventory.ProductID)),
			attribute.Int("inventory.quantity", inventory.Quantity),
			attribute.String("inventory.location", inventory.Location),
		),
	)
	defer span.End()

	err := r.GormInventoryRepository.Update(inventory)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// Delete with tracing
func (r *GormInventoryRepositoryWithTracing) DeleteWithContext(ctx context.Context, id uint) error {
	_, span := tracer.Start(ctx, "repository.Delete",
		trace.WithAttributes(
			attribute.Int("inventory.id", int(id)),
		),
	)
	defer span.End()

	err := r.GormInventoryRepository.Delete(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// UpdateQuantity with tracing
func (r *GormInventoryRepositoryWithTracing) UpdateQuantityWithContext(ctx context.Context, productID uint, quantity int) error {
	_, span := tracer.Start(ctx, "repository.UpdateQuantity",
		trace.WithAttributes(
			attribute.Int("inventory.product_id", int(productID)),
			attribute.Int("quantity.new_value", quantity),
		),
	)
	defer span.End()

	err := r.GormInventoryRepository.UpdateQuantity(productID, quantity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// Helper function to add database error details to span
func addDBErrorToSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("database error: %v", err))
	}
}
