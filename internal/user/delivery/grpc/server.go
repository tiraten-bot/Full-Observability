package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/tair/full-observability/api/proto/user"
	"github.com/tair/full-observability/internal/user/domain"
	"github.com/tair/full-observability/internal/user/usecase/command"
	"github.com/tair/full-observability/internal/user/usecase/query"
)

// UserServer implements the gRPC UserService
type UserServer struct {
	pb.UnimplementedUserServiceServer
	
	// Command handlers
	registerHandler     *command.RegisterUserHandler
	loginHandler        *command.LoginUserHandler
	updateHandler       *command.UpdateUserHandler
	deleteHandler       *command.DeleteUserHandler
	changeRoleHandler   *command.ChangeRoleHandler
	toggleActiveHandler *command.ToggleActiveHandler

	// Query handlers
	getUserHandler *query.GetUserHandler
	listHandler    *query.ListUsersHandler
	statsHandler   *query.GetStatsHandler
}

// NewUserServer creates a new gRPC user server (manual DI for backwards compatibility)
func NewUserServer(repo domain.UserRepository) *UserServer {
	return &UserServer{
		registerHandler:     command.NewRegisterUserHandler(repo),
		loginHandler:        command.NewLoginUserHandler(repo),
		updateHandler:       command.NewUpdateUserHandler(repo),
		deleteHandler:       command.NewDeleteUserHandler(repo),
		changeRoleHandler:   command.NewChangeRoleHandler(repo),
		toggleActiveHandler: command.NewToggleActiveHandler(repo),
		getUserHandler:      query.NewGetUserHandler(repo),
		listHandler:         query.NewListUsersHandler(repo),
		statsHandler:        query.NewGetStatsHandler(repo),
	}
}

// NewUserServerWithDI creates a new gRPC user server using dependency injection
// This is used by Wire for automatic dependency injection
func NewUserServerWithDI(
	registerHandler *command.RegisterUserHandler,
	loginHandler *command.LoginUserHandler,
	updateHandler *command.UpdateUserHandler,
	deleteHandler *command.DeleteUserHandler,
	changeRoleHandler *command.ChangeRoleHandler,
	toggleActiveHandler *command.ToggleActiveHandler,
	getUserHandler *query.GetUserHandler,
	listHandler *query.ListUsersHandler,
	statsHandler *query.GetStatsHandler,
) *UserServer {
	return &UserServer{
		registerHandler:     registerHandler,
		loginHandler:        loginHandler,
		updateHandler:       updateHandler,
		deleteHandler:       deleteHandler,
		changeRoleHandler:   changeRoleHandler,
		toggleActiveHandler: toggleActiveHandler,
		getUserHandler:      getUserHandler,
		listHandler:         listHandler,
		statsHandler:        statsHandler,
	}
}

// Register handles user registration
func (s *UserServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	cmd := command.RegisterUserCommand{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     domain.RoleUser,
	}

	user, err := s.registerHandler.Handle(cmd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.RegisterResponse{
		User: domainUserToProto(user),
	}, nil
}

// Login handles user authentication
func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	cmd := command.LoginUserCommand{
		Username: req.Username,
		Password: req.Password,
	}

	response, err := s.loginHandler.Handle(cmd)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	return &pb.LoginResponse{
		Token: response.Token,
		User:  domainUserToProto(response.User),
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	q := query.GetUserQuery{ID: uint(req.Id)}

	user, err := s.getUserHandler.Handle(q)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	return &pb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// UpdateUser updates user information
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	cmd := command.UpdateUserCommand{
		ID:       uint(req.Id),
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	}

	user, err := s.updateHandler.Handle(cmd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// DeleteUser deletes a user
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	cmd := command.DeleteUserCommand{ID: uint(req.Id)}

	if err := s.deleteHandler.Handle(cmd); err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

// ListUsers lists users with pagination
func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	q := query.ListUsersQuery{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
		Role:   req.Role,
	}

	users, err := s.listHandler.Handle(q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	protoUsers := make([]*pb.User, len(users))
	for i, user := range users {
		protoUsers[i] = domainUserToProto(&user)
	}

	return &pb.ListUsersResponse{
		Users: protoUsers,
		Total: int32(len(users)),
	}, nil
}

// ChangeRole changes user role (admin only)
func (s *UserServer) ChangeRole(ctx context.Context, req *pb.ChangeRoleRequest) (*pb.UserResponse, error) {
	cmd := command.ChangeRoleCommand{
		UserID: uint(req.UserId),
		Role:   req.Role,
	}

	user, err := s.changeRoleHandler.Handle(cmd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// ToggleActive toggles user active status (admin only)
func (s *UserServer) ToggleActive(ctx context.Context, req *pb.ToggleActiveRequest) (*pb.UserResponse, error) {
	cmd := command.ToggleActiveCommand{
		UserID:   uint(req.UserId),
		IsActive: req.IsActive,
	}

	user, err := s.toggleActiveHandler.Handle(cmd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.UserResponse{
		User: domainUserToProto(user),
	}, nil
}

// GetStats returns user statistics (admin only)
func (s *UserServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.StatsResponse, error) {
	q := query.GetStatsQuery{}

	stats, err := s.statsHandler.Handle(q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.StatsResponse{
		TotalUsers:  stats.TotalUsers,
		AdminCount:  stats.AdminCount,
		UserCount:   stats.UserCount,
		ActiveUsers: stats.ActiveUsers,
	}, nil
}

// Helper function to convert domain user to proto user
func domainUserToProto(user *domain.User) *pb.User {
	return &pb.User{
		Id:        uint32(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

