package repository

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/tair/full-observability/internal/product/domain"
)

var tracer = otel.Tracer("product-repository")

// GormProductRepositoryWithTracing wraps GormProductRepository with tracing
type GormProductRepositoryWithTracing struct {
	*GormProductRepository
}

// NewGormProductRepositoryWithTracing creates a new repository with tracing
func NewGormProductRepositoryWithTracing(db *gorm.DB) *GormProductRepositoryWithTracing {
	return &GormProductRepositoryWithTracing{
		GormProductRepository: NewGormProductRepository(db),
	}
}

// Create with tracing
func (r *GormProductRepositoryWithTracing) CreateWithContext(ctx context.Context, product *domain.Product) error {
	_, span := tracer.Start(ctx, "repository.Create",
		trace.WithAttributes(
			attribute.String("product.name", product.Name),
			attribute.String("product.sku", product.SKU),
			attribute.String("product.category", product.Category),
			attribute.Float64("product.price", product.Price),
			attribute.Int("product.stock", product.Stock),
		),
	)
	defer span.End()

	err := r.GormProductRepository.Create(product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Int("product.id", int(product.ID)))
	return nil
}

// FindByID with tracing
func (r *GormProductRepositoryWithTracing) FindByIDWithContext(ctx context.Context, id uint) (*domain.Product, error) {
	_, span := tracer.Start(ctx, "repository.FindByID",
		trace.WithAttributes(
			attribute.Int("product.id", int(id)),
		),
	)
	defer span.End()

	product, err := r.GormProductRepository.FindByID(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.String("product.name", product.Name),
		attribute.String("product.sku", product.SKU),
		attribute.Bool("product.is_active", product.IsActive),
	)
	return product, nil
}

// FindBySKU with tracing
func (r *GormProductRepositoryWithTracing) FindBySKUWithContext(ctx context.Context, sku string) (*domain.Product, error) {
	_, span := tracer.Start(ctx, "repository.FindBySKU",
		trace.WithAttributes(
			attribute.String("product.sku", sku),
		),
	)
	defer span.End()

	product, err := r.GormProductRepository.FindBySKU(sku)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("product.id", int(product.ID)),
		attribute.String("product.name", product.Name),
	)
	return product, nil
}

// FindAll with tracing
func (r *GormProductRepositoryWithTracing) FindAllWithContext(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	_, span := tracer.Start(ctx, "repository.FindAll",
		trace.WithAttributes(
			attribute.Int("query.limit", limit),
			attribute.Int("query.offset", offset),
		),
	)
	defer span.End()

	products, err := r.GormProductRepository.FindAll(limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("result.count", len(products)))
	return products, nil
}

// FindByCategory with tracing
func (r *GormProductRepositoryWithTracing) FindByCategoryWithContext(ctx context.Context, category string, limit, offset int) ([]domain.Product, error) {
	_, span := tracer.Start(ctx, "repository.FindByCategory",
		trace.WithAttributes(
			attribute.String("query.category", category),
			attribute.Int("query.limit", limit),
			attribute.Int("query.offset", offset),
		),
	)
	defer span.End()

	products, err := r.GormProductRepository.FindByCategory(category, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("result.count", len(products)))
	return products, nil
}

// Update with tracing
func (r *GormProductRepositoryWithTracing) UpdateWithContext(ctx context.Context, product *domain.Product) error {
	_, span := tracer.Start(ctx, "repository.Update",
		trace.WithAttributes(
			attribute.Int("product.id", int(product.ID)),
			attribute.String("product.name", product.Name),
			attribute.String("product.sku", product.SKU),
			attribute.Float64("product.price", product.Price),
		),
	)
	defer span.End()

	err := r.GormProductRepository.Update(product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// Delete with tracing
func (r *GormProductRepositoryWithTracing) DeleteWithContext(ctx context.Context, id uint) error {
	_, span := tracer.Start(ctx, "repository.Delete",
		trace.WithAttributes(
			attribute.Int("product.id", int(id)),
		),
	)
	defer span.End()

	err := r.GormProductRepository.Delete(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// Count with tracing
func (r *GormProductRepositoryWithTracing) CountWithContext(ctx context.Context) (int64, error) {
	_, span := tracer.Start(ctx, "repository.Count")
	defer span.End()

	count, err := r.GormProductRepository.Count()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	span.SetAttributes(attribute.Int64("result.count", count))
	return count, nil
}

// UpdateStock with tracing
func (r *GormProductRepositoryWithTracing) UpdateStockWithContext(ctx context.Context, id uint, stock int) error {
	_, span := tracer.Start(ctx, "repository.UpdateStock",
		trace.WithAttributes(
			attribute.Int("product.id", int(id)),
			attribute.Int("stock.new_value", stock),
		),
	)
	defer span.End()

	err := r.GormProductRepository.UpdateStock(id, stock)
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

