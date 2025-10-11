package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/tair/full-observability/pkg/auth"
)

// LoggingInterceptor logs gRPC requests
func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Call the handler
	resp, err := handler(ctx, req)

	// Log the request
	duration := time.Since(start)
	if err != nil {
		log.Printf("gRPC [%s] duration: %v, error: %v", info.FullMethod, duration, err)
	} else {
		log.Printf("gRPC [%s] duration: %v", info.FullMethod, duration)
	}

	return resp, err
}

// AuthInterceptor validates JWT tokens for protected methods
func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Skip auth for public methods
	publicMethods := map[string]bool{
		"/user.v1.UserService/Register": true,
		"/user.v1.UserService/Login":    true,
	}

	if publicMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not provided")
	}

	// Get authorization token
	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token not provided")
	}

	// Validate token
	token := tokens[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Add claims to context
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "username", claims.Username)
	ctx = context.WithValue(ctx, "role", claims.Role)

	// Check admin methods
	adminMethods := map[string]bool{
		"/user.v1.UserService/ChangeRole":   true,
		"/user.v1.UserService/ToggleActive": true,
		"/user.v1.UserService/GetStats":     true,
	}

	if adminMethods[info.FullMethod] && claims.Role != "admin" {
		return nil, status.Errorf(codes.PermissionDenied, "admin access required")
	}

	return handler(ctx, req)
}

