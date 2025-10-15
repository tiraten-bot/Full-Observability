package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	pb "github.com/tair/full-observability/api/proto/product"
	"github.com/tair/full-observability/pkg/logger"
)

// ProductServiceClient wraps the gRPC client for product service
type ProductServiceClient struct {
	client pb.ProductServiceClient
	conn   *grpc.ClientConn
}

// NewProductServiceClient creates a new product service gRPC client
func NewProductServiceClient(address string) (*ProductServiceClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	client := pb.NewProductServiceClient(conn)

	logger.Logger.Info().
		Str("address", address).
		Msg("Connected to Product Service gRPC server")

	return &ProductServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *ProductServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetProduct gets a product by ID
func (c *ProductServiceClient) GetProduct(ctx context.Context, productID uint) (*pb.Product, error) {
	// Set timeout for this request
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.GetProductRequest{
		Id: uint32(productID),
	}

	resp, err := c.client.GetProduct(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return resp.Product, nil
}

// CheckAvailability checks if a product is available with required quantity
func (c *ProductServiceClient) CheckAvailability(ctx context.Context, productID uint, quantity int32) (bool, int32, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.CheckAvailabilityRequest{
		ProductId: uint32(productID),
		Quantity:  quantity,
	}

	resp, err := c.client.CheckAvailability(ctx, req)
	if err != nil {
		return false, 0, fmt.Errorf("failed to check availability: %w", err)
	}

	return resp.Available, resp.CurrentStock, nil
}

// UpdateStock updates product stock (requires admin token)
func (c *ProductServiceClient) UpdateStock(ctx context.Context, productID uint, stock int32, token string) error {
	// Add authorization metadata
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.UpdateStockRequest{
		ProductId: uint32(productID),
		Stock:     stock,
	}

	_, err := c.client.UpdateStock(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

// ListProducts lists products with optional category filter
func (c *ProductServiceClient) ListProducts(ctx context.Context, limit, offset int32, category string) ([]*pb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.ListProductsRequest{
		Limit:    limit,
		Offset:   offset,
		Category: category,
	}

	resp, err := c.client.ListProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return resp.Products, nil
}
