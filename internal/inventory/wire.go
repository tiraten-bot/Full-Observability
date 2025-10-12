//go:build wireinject
// +build wireinject

package inventory

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/tair/full-observability/internal/inventory/delivery/http"
	"github.com/tair/full-observability/internal/inventory/domain"
	"github.com/tair/full-observability/internal/inventory/repository"
)

// ProvideInventoryRepository provides the inventory repository
func ProvideInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return repository.NewGormInventoryRepository(db)
}

// Wire sets
var RepositorySet = wire.NewSet(
	ProvideInventoryRepository,
)

// InitializeHTTPHandler initializes HTTP handler with all dependencies
func InitializeHTTPHandler(db *gorm.DB) (*http.InventoryHandler, error) {
	wire.Build(
		RepositorySet,
		http.NewInventoryHandler,
	)
	return nil, nil
}

