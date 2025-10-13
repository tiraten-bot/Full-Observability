//go:build wireinject
// +build wireinject

package inventory

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/tair/full-observability/internal/inventory/client"
	grpcDelivery "github.com/tair/full-observability/internal/inventory/delivery/grpc"
	"github.com/tair/full-observability/internal/inventory/delivery/http"
	"github.com/tair/full-observability/internal/inventory/domain"
	"github.com/tair/full-observability/internal/inventory/repository"
	"github.com/tair/full-observability/internal/inventory/usecase/command"
	"github.com/tair/full-observability/internal/inventory/usecase/query"
)

// ProvideInventoryRepository provides the inventory repository
func ProvideInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return repository.NewGormInventoryRepository(db)
}

// Command Handlers Providers
func ProvideCreateInventoryHandler(repo domain.InventoryRepository) *command.CreateInventoryHandler {
	return command.NewCreateInventoryHandler(repo)
}

func ProvideUpdateQuantityHandler(repo domain.InventoryRepository) *command.UpdateQuantityHandler {
	return command.NewUpdateQuantityHandler(repo)
}

func ProvideDeleteInventoryHandler(repo domain.InventoryRepository) *command.DeleteInventoryHandler {
	return command.NewDeleteInventoryHandler(repo)
}

// Query Handlers Providers
func ProvideGetInventoryHandler(repo domain.InventoryRepository) *query.GetInventoryHandler {
	return query.NewGetInventoryHandler(repo)
}

func ProvideListInventoryHandler(repo domain.InventoryRepository) *query.ListInventoryHandler {
	return query.NewListInventoryHandler(repo)
}

// ProvideUserServiceClient provides the user service gRPC client
func ProvideUserServiceClient(userServiceAddr string) (*client.UserServiceClient, error) {
	return client.NewUserServiceClient(userServiceAddr)
}

// Wire sets
var RepositorySet = wire.NewSet(
	ProvideInventoryRepository,
)

var CommandHandlerSet = wire.NewSet(
	ProvideCreateInventoryHandler,
	ProvideUpdateQuantityHandler,
	ProvideDeleteInventoryHandler,
)

var QueryHandlerSet = wire.NewSet(
	ProvideGetInventoryHandler,
	ProvideListInventoryHandler,
)

var AllHandlersSet = wire.NewSet(
	RepositorySet,
	CommandHandlerSet,
	QueryHandlerSet,
)

// InitializeHTTPHandler initializes HTTP handler with all dependencies
func InitializeHTTPHandler(db *gorm.DB, userServiceAddr string) (*http.InventoryHandler, error) {
	wire.Build(
		AllHandlersSet,
		ProvideUserServiceClient,
		http.NewInventoryHandlerWithDI,
	)
	return nil, nil
}

// InitializeGRPCServer initializes gRPC server with all dependencies
func InitializeGRPCServer(db *gorm.DB) (*grpcDelivery.InventoryGRPCServer, error) {
	wire.Build(
		AllHandlersSet,
		grpcDelivery.NewInventoryGRPCServer,
	)
	return nil, nil
}

