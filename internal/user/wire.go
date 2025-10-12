//go:build wireinject
// +build wireinject

package user

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/tair/full-observability/internal/user/delivery/grpc"
	"github.com/tair/full-observability/internal/user/delivery/http"
	"github.com/tair/full-observability/internal/user/domain"
	"github.com/tair/full-observability/internal/user/repository"
	"github.com/tair/full-observability/internal/user/usecase/command"
	"github.com/tair/full-observability/internal/user/usecase/query"
)

// ProvideUserRepository provides the user repository
func ProvideUserRepository(db *gorm.DB) domain.UserRepository {
	return repository.NewGormUserRepository(db)
}

// Command Handlers Providers
func ProvideRegisterUserHandler(repo domain.UserRepository) *command.RegisterUserHandler {
	return command.NewRegisterUserHandler(repo)
}

func ProvideLoginUserHandler(repo domain.UserRepository) *command.LoginUserHandler {
	return command.NewLoginUserHandler(repo)
}

func ProvideUpdateUserHandler(repo domain.UserRepository) *command.UpdateUserHandler {
	return command.NewUpdateUserHandler(repo)
}

func ProvideDeleteUserHandler(repo domain.UserRepository) *command.DeleteUserHandler {
	return command.NewDeleteUserHandler(repo)
}

func ProvideChangeRoleHandler(repo domain.UserRepository) *command.ChangeRoleHandler {
	return command.NewChangeRoleHandler(repo)
}

func ProvideToggleActiveHandler(repo domain.UserRepository) *command.ToggleActiveHandler {
	return command.NewToggleActiveHandler(repo)
}

// Query Handlers Providers
func ProvideGetUserHandler(repo domain.UserRepository) *query.GetUserHandler {
	return query.NewGetUserHandler(repo)
}

func ProvideListUsersHandler(repo domain.UserRepository) *query.ListUsersHandler {
	return query.NewListUsersHandler(repo)
}

func ProvideGetStatsHandler(repo domain.UserRepository) *query.GetStatsHandler {
	return query.NewGetStatsHandler(repo)
}

// CommandHandlers is a struct that holds all command handlers
type CommandHandlers struct {
	RegisterHandler     *command.RegisterUserHandler
	LoginHandler        *command.LoginUserHandler
	UpdateHandler       *command.UpdateUserHandler
	DeleteHandler       *command.DeleteUserHandler
	ChangeRoleHandler   *command.ChangeRoleHandler
	ToggleActiveHandler *command.ToggleActiveHandler
}

// QueryHandlers is a struct that holds all query handlers
type QueryHandlers struct {
	GetUserHandler *query.GetUserHandler
	ListHandler    *query.ListUsersHandler
	StatsHandler   *query.GetStatsHandler
}

// ProvideCommandHandlers provides all command handlers
func ProvideCommandHandlers(
	registerHandler *command.RegisterUserHandler,
	loginHandler *command.LoginUserHandler,
	updateHandler *command.UpdateUserHandler,
	deleteHandler *command.DeleteUserHandler,
	changeRoleHandler *command.ChangeRoleHandler,
	toggleActiveHandler *command.ToggleActiveHandler,
) *CommandHandlers {
	return &CommandHandlers{
		RegisterHandler:     registerHandler,
		LoginHandler:        loginHandler,
		UpdateHandler:       updateHandler,
		DeleteHandler:       deleteHandler,
		ChangeRoleHandler:   changeRoleHandler,
		ToggleActiveHandler: toggleActiveHandler,
	}
}

// ProvideQueryHandlers provides all query handlers
func ProvideQueryHandlers(
	getUserHandler *query.GetUserHandler,
	listHandler *query.ListUsersHandler,
	statsHandler *query.GetStatsHandler,
) *QueryHandlers {
	return &QueryHandlers{
		GetUserHandler: getUserHandler,
		ListHandler:    listHandler,
		StatsHandler:   statsHandler,
	}
}

// Wire sets
var RepositorySet = wire.NewSet(
	ProvideUserRepository,
)

var CommandHandlerSet = wire.NewSet(
	ProvideRegisterUserHandler,
	ProvideLoginUserHandler,
	ProvideUpdateUserHandler,
	ProvideDeleteUserHandler,
	ProvideChangeRoleHandler,
	ProvideToggleActiveHandler,
	ProvideCommandHandlers,
)

var QueryHandlerSet = wire.NewSet(
	ProvideGetUserHandler,
	ProvideListUsersHandler,
	ProvideGetStatsHandler,
	ProvideQueryHandlers,
)

var AllHandlersSet = wire.NewSet(
	RepositorySet,
	CommandHandlerSet,
	QueryHandlerSet,
)

// InitializeHTTPHandler initializes HTTP handler with all dependencies
func InitializeHTTPHandler(db *gorm.DB) (*http.UserHandler, error) {
	wire.Build(
		AllHandlersSet,
		http.NewUserHandlerWithDI,
	)
	return nil, nil
}

// InitializeGRPCServer initializes gRPC server with all dependencies
func InitializeGRPCServer(db *gorm.DB) (*grpc.UserServer, error) {
	wire.Build(
		AllHandlersSet,
		grpc.NewUserServerWithDI,
	)
	return nil, nil
}

