package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/tair/full-observability/api/proto/inventory"
	"github.com/tair/full-observability/pkg/logger"
)

// InventoryServiceClient wraps the gRPC client for inventory service
type InventoryServiceClient struct {
	client pb.InventoryServiceClient
	conn   *grpc.ClientConn
}

// NewInventoryServiceClient creates a new inventory service gRPC client
func NewInventoryServiceClient(address string) (*InventoryServiceClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	client := pb.NewInventoryServiceClient(conn)

	logger.Logger.Info().
		Str("address", address).
		Msg("Connected to Inventory Service gRPC server")

	return &InventoryServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *InventoryServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetByProductID gets inventory by product ID
func (c *InventoryServiceClient) GetByProductID(ctx context.Context, productID uint) (*pb.Inventory, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.GetByProductIDRequest{
		ProductId: uint32(productID),
	}

	resp, err := c.client.GetByProductID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory by product ID: %w", err)
	}

	return resp.Inventory, nil
}

// CheckAvailability checks if a product is available with required quantity
func (c *InventoryServiceClient) CheckAvailability(ctx context.Context, productID uint, requiredQuantity int32) (bool, int32, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.CheckAvailabilityRequest{
		ProductId:        uint32(productID),
		RequiredQuantity: requiredQuantity,
	}

	resp, err := c.client.CheckAvailability(ctx, req)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to check availability: %w", err)
	}

	return resp.Available, resp.CurrentQuantity, resp.Message, nil
}

// ReserveStock reserves stock for a product
func (c *InventoryServiceClient) ReserveStock(ctx context.Context, productID uint, quantity int32, reservationID string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.ReserveStockRequest{
		ProductId:     uint32(productID),
		Quantity:      quantity,
		ReservationId: reservationID,
	}

	resp, err := c.client.ReserveStock(ctx, req)
	if err != nil {
		return false, "", fmt.Errorf("failed to reserve stock: %w", err)
	}

	return resp.Success, resp.Message, nil
}

// ReleaseStock releases reserved stock
func (c *InventoryServiceClient) ReleaseStock(ctx context.Context, productID uint, quantity int32, reservationID string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.ReleaseStockRequest{
		ProductId:     uint32(productID),
		Quantity:      quantity,
		ReservationId: reservationID,
	}

	resp, err := c.client.ReleaseStock(ctx, req)
	if err != nil {
		return false, "", fmt.Errorf("failed to release stock: %w", err)
	}

	return resp.Success, resp.Message, nil
}

// GetInventory gets inventory by ID
func (c *InventoryServiceClient) GetInventory(ctx context.Context, inventoryID uint) (*pb.Inventory, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.GetInventoryRequest{
		Id: uint32(inventoryID),
	}

	resp, err := c.client.GetInventory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return resp.Inventory, nil
}

// ListInventory lists all inventory records
func (c *InventoryServiceClient) ListInventory(ctx context.Context, limit, offset int32) ([]*pb.Inventory, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &pb.ListInventoryRequest{
		Limit:  limit,
		Offset: offset,
	}

	resp, err := c.client.ListInventory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory: %w", err)
	}

	return resp.Inventories, nil
}

