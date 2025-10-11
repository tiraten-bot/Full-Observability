package repository

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"github.com/tair/full-observability/internal/user/domain"
)

var tracer = otel.Tracer("user-repository")

// GormUserRepositoryWithTracing wraps GormUserRepository with tracing
type GormUserRepositoryWithTracing struct {
	*GormUserRepository
}

// NewGormUserRepositoryWithTracing creates a new repository with tracing
func NewGormUserRepositoryWithTracing(db *gorm.DB) *GormUserRepositoryWithTracing {
	return &GormUserRepositoryWithTracing{
		GormUserRepository: NewGormUserRepository(db),
	}
}

// Create with tracing
func (r *GormUserRepositoryWithTracing) CreateWithContext(ctx context.Context, user *domain.User) error {
	_, span := tracer.Start(ctx, "repository.Create",
		trace.WithAttributes(
			attribute.String("user.username", user.Username),
			attribute.String("user.email", user.Email),
		),
	)
	defer span.End()

	err := r.GormUserRepository.Create(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Int("user.id", int(user.ID)))
	return nil
}

// FindByID with tracing
func (r *GormUserRepositoryWithTracing) FindByIDWithContext(ctx context.Context, id uint) (*domain.User, error) {
	_, span := tracer.Start(ctx, "repository.FindByID",
		trace.WithAttributes(
			attribute.Int("user.id", int(id)),
		),
	)
	defer span.End()

	user, err := r.GormUserRepository.FindByID(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.String("user.username", user.Username))
	return user, nil
}

// FindByUsername with tracing
func (r *GormUserRepositoryWithTracing) FindByUsernameWithContext(ctx context.Context, username string) (*domain.User, error) {
	_, span := tracer.Start(ctx, "repository.FindByUsername",
		trace.WithAttributes(
			attribute.String("user.username", username),
		),
	)
	defer span.End()

	user, err := r.GormUserRepository.FindByUsername(username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("user.id", int(user.ID)))
	return user, nil
}

// FindAll with tracing
func (r *GormUserRepositoryWithTracing) FindAllWithContext(ctx context.Context, limit, offset int) ([]domain.User, error) {
	_, span := tracer.Start(ctx, "repository.FindAll",
		trace.WithAttributes(
			attribute.Int("query.limit", limit),
			attribute.Int("query.offset", offset),
		),
	)
	defer span.End()

	users, err := r.GormUserRepository.FindAll(limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))
	return users, nil
}

// Update with tracing
func (r *GormUserRepositoryWithTracing) UpdateWithContext(ctx context.Context, user *domain.User) error {
	_, span := tracer.Start(ctx, "repository.Update",
		trace.WithAttributes(
			attribute.Int("user.id", int(user.ID)),
			attribute.String("user.email", user.Email),
		),
	)
	defer span.End()

	err := r.GormUserRepository.Update(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// Delete with tracing
func (r *GormUserRepositoryWithTracing) DeleteWithContext(ctx context.Context, id uint) error {
	_, span := tracer.Start(ctx, "repository.Delete",
		trace.WithAttributes(
			attribute.Int("user.id", int(id)),
		),
	)
	defer span.End()

	err := r.GormUserRepository.Delete(id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// Count with tracing
func (r *GormUserRepositoryWithTracing) CountWithContext(ctx context.Context) (int64, error) {
	_, span := tracer.Start(ctx, "repository.Count")
	defer span.End()

	count, err := r.GormUserRepository.Count()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	span.SetAttributes(attribute.Int64("result.count", count))
	return count, nil
}

// Helper function to add database error details to span
func addDBErrorToSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("database error: %v", err))
	}
}

