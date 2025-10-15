package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/tair/full-observability/api/proto/inventory"
	"github.com/tair/full-observability/internal/inventory/domain"
	"github.com/tair/full-observability/internal/inventory/usecase/command"
	"github.com/tair/full-observability/internal/inventory/usecase/query"
	"github.com/tair/full-observability/pkg/logger"
)

// InventoryGRPCServer implements the InventoryService gRPC server
type InventoryGRPCServer struct {
	pb.UnimplementedInventoryServiceServer

	// Command handlers
	createHandler         *command.CreateInventoryHandler
	updateQuantityHandler *command.UpdateQuantityHandler
	deleteHandler         *command.DeleteInventoryHandler

	// Query handlers
	getHandler  *query.GetInventoryHandler
	listHandler *query.ListInventoryHandler

	repo domain.InventoryRepository
}

// NewInventoryGRPCServer creates a new gRPC server
func NewInventoryGRPCServer(
	createHandler *command.CreateInventoryHandler,
	updateQuantityHandler *command.UpdateQuantityHandler,
	deleteHandler *command.DeleteInventoryHandler,
	getHandler *query.GetInventoryHandler,
	listHandler *query.ListInventoryHandler,
	repo domain.InventoryRepository,
) *InventoryGRPCServer {
	return &InventoryGRPCServer{
		createHandler:         createHandler,
		updateQuantityHandler: updateQuantityHandler,
		deleteHandler:         deleteHandler,
		getHandler:            getHandler,
		listHandler:           listHandler,
		repo:                  repo,
	}
}

// GetRepository returns the repository (for Kafka consumer)
func (s *InventoryGRPCServer) GetRepository() domain.InventoryRepository {
	return s.repo
}

// CreateInventory creates a new inventory record
func (s *InventoryGRPCServer) CreateInventory(ctx context.Context, req *pb.CreateInventoryRequest) (*pb.InventoryResponse, error) {
	logger.Logger.Info().
		Uint32("product_id", req.ProductId).
		Int32("quantity", req.Quantity).
		Str("location", req.Location).
		Msg("gRPC: CreateInventory called")

	cmd := command.CreateInventoryCommand{
		ProductID: uint(req.ProductId),
		Quantity:  int(req.Quantity),
		Location:  req.Location,
	}

	inventory, err := s.createHandler.Handle(cmd)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to create inventory")
		return nil, status.Errorf(codes.Internal, "failed to create inventory: %v", err)
	}

	return &pb.InventoryResponse{
		Success:   true,
		Message:   "Inventory created successfully",
		Inventory: domainToProto(inventory),
	}, nil
}

// GetInventory retrieves an inventory record by ID
func (s *InventoryGRPCServer) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.InventoryResponse, error) {
	logger.Logger.Info().
		Uint32("id", req.Id).
		Msg("gRPC: GetInventory called")

	q := query.GetInventoryQuery{
		ID: uint(req.Id),
	}

	inventory, err := s.getHandler.Handle(q)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to get inventory")
		return nil, status.Errorf(codes.NotFound, "inventory not found: %v", err)
	}

	return &pb.InventoryResponse{
		Success:   true,
		Message:   "Inventory retrieved successfully",
		Inventory: domainToProto(inventory),
	}, nil
}

// UpdateQuantity updates the quantity of an inventory record
func (s *InventoryGRPCServer) UpdateQuantity(ctx context.Context, req *pb.UpdateQuantityRequest) (*pb.InventoryResponse, error) {
	logger.Logger.Info().
		Uint32("id", req.Id).
		Int32("quantity", req.Quantity).
		Msg("gRPC: UpdateQuantity called")

	// First get the inventory to find product_id
	inventory, err := s.repo.FindByID(uint(req.Id))
	if err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to find inventory")
		return nil, status.Errorf(codes.NotFound, "inventory not found: %v", err)
	}

	cmd := command.UpdateQuantityCommand{
		ProductID: inventory.ProductID,
		Quantity:  int(req.Quantity),
	}

	if err := s.updateQuantityHandler.Handle(cmd); err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to update quantity")
		return nil, status.Errorf(codes.Internal, "failed to update quantity: %v", err)
	}

	// Fetch updated inventory
	inventory, err = s.repo.FindByID(uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch updated inventory: %v", err)
	}

	return &pb.InventoryResponse{
		Success:   true,
		Message:   "Quantity updated successfully",
		Inventory: domainToProto(inventory),
	}, nil
}

// DeleteInventory deletes an inventory record
func (s *InventoryGRPCServer) DeleteInventory(ctx context.Context, req *pb.DeleteInventoryRequest) (*pb.DeleteInventoryResponse, error) {
	logger.Logger.Info().
		Uint32("id", req.Id).
		Msg("gRPC: DeleteInventory called")

	cmd := command.DeleteInventoryCommand{
		ID: uint(req.Id),
	}

	if err := s.deleteHandler.Handle(cmd); err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to delete inventory")
		return nil, status.Errorf(codes.Internal, "failed to delete inventory: %v", err)
	}

	return &pb.DeleteInventoryResponse{
		Success: true,
		Message: "Inventory deleted successfully",
	}, nil
}

// ListInventory lists all inventory records
func (s *InventoryGRPCServer) ListInventory(ctx context.Context, req *pb.ListInventoryRequest) (*pb.ListInventoryResponse, error) {
	logger.Logger.Info().
		Int32("limit", req.Limit).
		Int32("offset", req.Offset).
		Msg("gRPC: ListInventory called")

	limit := int(req.Limit)
	offset := int(req.Offset)

	if limit == 0 {
		limit = 10
	}

	q := query.ListInventoryQuery{
		Limit:  limit,
		Offset: offset,
	}

	inventories, err := s.listHandler.Handle(q)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to list inventories")
		return nil, status.Errorf(codes.Internal, "failed to list inventories: %v", err)
	}

	pbInventories := make([]*pb.Inventory, len(inventories))
	for i, inv := range inventories {
		pbInventories[i] = domainToProto(&inv)
	}

	return &pb.ListInventoryResponse{
		Success:     true,
		Inventories: pbInventories,
		Total:       int32(len(inventories)),
	}, nil
}

// GetByProductID retrieves inventory by product ID
func (s *InventoryGRPCServer) GetByProductID(ctx context.Context, req *pb.GetByProductIDRequest) (*pb.InventoryResponse, error) {
	logger.Logger.Info().
		Uint32("product_id", req.ProductId).
		Msg("gRPC: GetByProductID called")

	inventory, err := s.repo.FindByProductID(uint(req.ProductId))
	if err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to get inventory by product ID")
		return nil, status.Errorf(codes.NotFound, "inventory not found for product: %v", err)
	}

	return &pb.InventoryResponse{
		Success:   true,
		Message:   "Inventory retrieved successfully",
		Inventory: domainToProto(inventory),
	}, nil
}

// CheckAvailability checks if a product has sufficient quantity
func (s *InventoryGRPCServer) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	logger.Logger.Info().
		Uint32("product_id", req.ProductId).
		Int32("required_quantity", req.RequiredQuantity).
		Msg("gRPC: CheckAvailability called")

	inventory, err := s.repo.FindByProductID(uint(req.ProductId))
	if err != nil {
		logger.Logger.Error().Err(err).Msg("gRPC: Failed to check availability")
		return &pb.CheckAvailabilityResponse{
			Available:       false,
			CurrentQuantity: 0,
			Message:         fmt.Sprintf("Product not found: %v", err),
		}, nil
	}

	available := inventory.Quantity >= int(req.RequiredQuantity)
	message := "Available"
	if !available {
		message = fmt.Sprintf("Insufficient stock. Available: %d, Required: %d", inventory.Quantity, req.RequiredQuantity)
	}

	return &pb.CheckAvailabilityResponse{
		Available:       available,
		CurrentQuantity: int32(inventory.Quantity),
		Message:         message,
	}, nil
}

// ReserveStock reserves stock for a product
func (s *InventoryGRPCServer) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	logger.Logger.Info().
		Uint32("product_id", req.ProductId).
		Int32("quantity", req.Quantity).
		Str("reservation_id", req.ReservationId).
		Msg("gRPC: ReserveStock called")

	inventory, err := s.repo.FindByProductID(uint(req.ProductId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	if inventory.Quantity < int(req.Quantity) {
		return &pb.ReserveStockResponse{
			Success:       false,
			Message:       "Insufficient stock",
			ReservationId: "",
		}, nil
	}

	// Update quantity (reduce)
	newQuantity := inventory.Quantity - int(req.Quantity)
	if err := s.repo.UpdateQuantity(uint(req.ProductId), newQuantity); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve stock: %v", err)
	}

	return &pb.ReserveStockResponse{
		Success:       true,
		Message:       "Stock reserved successfully",
		ReservationId: req.ReservationId,
	}, nil
}

// ReleaseStock releases reserved stock
func (s *InventoryGRPCServer) ReleaseStock(ctx context.Context, req *pb.ReleaseStockRequest) (*pb.ReleaseStockResponse, error) {
	logger.Logger.Info().
		Uint32("product_id", req.ProductId).
		Int32("quantity", req.Quantity).
		Str("reservation_id", req.ReservationId).
		Msg("gRPC: ReleaseStock called")

	inventory, err := s.repo.FindByProductID(uint(req.ProductId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	// Update quantity (increase)
	newQuantity := inventory.Quantity + int(req.Quantity)
	if err := s.repo.UpdateQuantity(uint(req.ProductId), newQuantity); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to release stock: %v", err)
	}

	return &pb.ReleaseStockResponse{
		Success: true,
		Message: "Stock released successfully",
	}, nil
}

// domainToProto converts domain Inventory to proto Inventory
func domainToProto(inv *domain.Inventory) *pb.Inventory {
	if inv == nil {
		return nil
	}

	return &pb.Inventory{
		Id:        uint32(inv.ID),
		ProductId: uint32(inv.ProductID),
		Quantity:  int32(inv.Quantity),
		Location:  inv.Location,
		CreatedAt: timestamppb.New(inv.CreatedAt),
		UpdatedAt: timestamppb.New(inv.UpdatedAt),
	}
}
