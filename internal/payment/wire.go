//go:build wireinject
// +build wireinject

package payment

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/tair/full-observability/internal/payment/client"
	"github.com/tair/full-observability/internal/payment/domain"
	"github.com/tair/full-observability/internal/payment/handler"
	"github.com/tair/full-observability/internal/payment/repository"
	"github.com/tair/full-observability/internal/payment/usecase/command"
	"github.com/tair/full-observability/internal/payment/usecase/query"
)

// ProvidePaymentRepository provides the payment repository
func ProvidePaymentRepository(db *gorm.DB) domain.PaymentRepository {
	return repository.NewGormPaymentRepository(db)
}

// Command Handlers Providers
func ProvideCreatePaymentHandler(repo domain.PaymentRepository) *command.CreatePaymentHandler {
	return command.NewCreatePaymentHandler(repo)
}

func ProvideUpdateStatusHandler(repo domain.PaymentRepository) *command.UpdateStatusHandler {
	return command.NewUpdateStatusHandler(repo)
}

// Query Handlers Providers
func ProvideGetPaymentHandler(repo domain.PaymentRepository) *query.GetPaymentHandler {
	return query.NewGetPaymentHandler(repo)
}

func ProvideListPaymentsHandler(repo domain.PaymentRepository) *query.ListPaymentsHandler {
	return query.NewListPaymentsHandler(repo)
}

func ProvideGetMyPaymentsHandler(repo domain.PaymentRepository) *query.GetMyPaymentsHandler {
	return query.NewGetMyPaymentsHandler(repo)
}

// ProvideUserServiceClient provides the user service gRPC client
func ProvideUserServiceClient(userServiceAddr string) (*client.UserServiceClient, error) {
	return client.NewUserServiceClient(userServiceAddr)
}

// Wire sets
var RepositorySet = wire.NewSet(
	ProvidePaymentRepository,
)

var CommandHandlerSet = wire.NewSet(
	ProvideCreatePaymentHandler,
	ProvideUpdateStatusHandler,
)

var QueryHandlerSet = wire.NewSet(
	ProvideGetPaymentHandler,
	ProvideListPaymentsHandler,
	ProvideGetMyPaymentsHandler,
)

var AllHandlersSet = wire.NewSet(
	RepositorySet,
	CommandHandlerSet,
	QueryHandlerSet,
)

// InitializeHandler initializes payment handler with all dependencies
func InitializeHandler(db *gorm.DB, userServiceAddr string) (*handler.PaymentHandler, error) {
	wire.Build(
		AllHandlersSet,
		ProvideUserServiceClient,
		handler.NewPaymentHandlerWithDI,
	)
	return nil, nil
}

