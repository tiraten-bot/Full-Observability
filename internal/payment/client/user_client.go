package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	pb "github.com/tair/full-observability/api/proto/user"
	"github.com/tair/full-observability/pkg/logger"
)

// UserServiceClient wraps the gRPC client for user service
type UserServiceClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

// NewUserServiceClient creates a new user service gRPC client
func NewUserServiceClient(address string) (*UserServiceClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	client := pb.NewUserServiceClient(conn)

	logger.Logger.Info().
		Str("address", address).
		Msg("Connected to User Service gRPC server")

	return &UserServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetUser gets a user by ID with authentication token
func (c *UserServiceClient) GetUser(ctx context.Context, userID uint, token string) (*pb.User, error) {
	// Add authorization metadata
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Set timeout for this request
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.GetUserRequest{
		Id: uint32(userID),
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return resp.User, nil
}

// ValidateUserRole validates if a user has the required role
func (c *UserServiceClient) ValidateUserRole(ctx context.Context, userID uint, token string, requiredRole string) (bool, error) {
	user, err := c.GetUser(ctx, userID, token)
	if err != nil {
		return false, err
	}

	// Check if user is active
	if !user.IsActive {
		return false, fmt.Errorf("user is not active")
	}

	// Check role
	if user.Role != requiredRole {
		return false, fmt.Errorf("user does not have required role: %s (has: %s)", requiredRole, user.Role)
	}

	return true, nil
}

// CheckUserPermissions checks if user has permission (is active and has valid role)
func (c *UserServiceClient) CheckUserPermissions(ctx context.Context, userID uint, token string, allowedRoles []string) (bool, string, error) {
	user, err := c.GetUser(ctx, userID, token)
	if err != nil {
		return false, "", err
	}

	// Check if user is active
	if !user.IsActive {
		return false, "", fmt.Errorf("user account is disabled")
	}

	// Check if user has one of the allowed roles
	for _, role := range allowedRoles {
		if user.Role == role {
			return true, user.Role, nil
		}
	}

	return false, user.Role, fmt.Errorf("insufficient permissions")
}
