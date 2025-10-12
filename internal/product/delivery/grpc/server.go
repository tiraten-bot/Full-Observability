package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/tair/full-observability/api/proto/product"
	"github.com/tair/full-observability/internal/product/domain"
	"github.com/tair/full-observability/internal/product/usecase/command"
	"github.com/tair/full-observability/internal/product/usecase/query"
)

// ProductServer implements the gRPC ProductService
type ProductServer struct {
	pb.UnimplementedProductServiceServer

	// Command handlers
	createHandler      *command.CreateProductHandler
	updateHandler      *command.UpdateProductHandler
	deleteHandler      *command.DeleteProductHandler
	updateStockHandler *command.UpdateStockHandler

	// Query handlers
	getProductHandler *query.GetProductHandler
	listHandler       *query.ListProductsHandler
	statsHandler      *query.GetStatsHandler

	// Repository for direct access when needed
	repo domain.ProductRepository
}

// NewProductServer creates a new gRPC product server (manual DI for backwards compatibility)
func NewProductServer(repo domain.ProductRepository) *ProductServer {
	return &ProductServer{
		createHandler:      command.NewCreateProductHandler(repo),
		updateHandler:      command.NewUpdateProductHandler(repo),
		deleteHandler:      command.NewDeleteProductHandler(repo),
		updateStockHandler: command.NewUpdateStockHandler(repo),
		getProductHandler:  query.NewGetProductHandler(repo),
		listHandler:        query.NewListProductsHandler(repo),
		statsHandler:       query.NewGetStatsHandler(repo),
		repo:               repo,
	}
}

// NewProductServerWithDI creates a new gRPC product server using dependency injection
// This is used by Wire for automatic dependency injection
func NewProductServerWithDI(
	createHandler *command.CreateProductHandler,
	updateHandler *command.UpdateProductHandler,
	deleteHandler *command.DeleteProductHandler,
	updateStockHandler *command.UpdateStockHandler,
	getProductHandler *query.GetProductHandler,
	listHandler *query.ListProductsHandler,
	statsHandler *query.GetStatsHandler,
	repo domain.ProductRepository,
) *ProductServer {
	return &ProductServer{
		createHandler:      createHandler,
		updateHandler:      updateHandler,
		deleteHandler:      deleteHandler,
		updateStockHandler: updateStockHandler,
		getProductHandler:  getProductHandler,
		listHandler:        listHandler,
		statsHandler:       statsHandler,
		repo:               repo,
	}
}

// CreateProduct handles product creation
func (s *ProductServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	cmd := command.CreateProductCommand{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		Category:    req.Category,
		SKU:         req.Sku,
		IsActive:    req.IsActive,
	}

	product, err := s.createHandler.Handle(cmd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create product: %v", err)
	}

	return &pb.ProductResponse{
		Product: domainProductToProto(product),
	}, nil
}

// GetProduct retrieves a product by ID
func (s *ProductServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	q := query.GetProductQuery{ID: uint(req.Id)}

	product, err := s.getProductHandler.Handle(q)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	return &pb.ProductResponse{
		Product: domainProductToProto(product),
	}, nil
}

// UpdateProduct updates product information
func (s *ProductServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	cmd := command.UpdateProductCommand{
		ID:          uint(req.Id),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		Category:    req.Category,
		SKU:         req.Sku,
		IsActive:    req.IsActive,
	}

	product, err := s.updateHandler.Handle(cmd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to update product: %v", err)
	}

	return &pb.ProductResponse{
		Product: domainProductToProto(product),
	}, nil
}

// DeleteProduct deletes a product
func (s *ProductServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	cmd := command.DeleteProductCommand{ID: uint(req.Id)}

	if err := s.deleteHandler.Handle(cmd); err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to delete product: %v", err)
	}

	return &pb.DeleteProductResponse{
		Message: "Product deleted successfully",
	}, nil
}

// ListProducts lists products with pagination
func (s *ProductServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	q := query.ListProductsQuery{
		Limit:    int(req.Limit),
		Offset:   int(req.Offset),
		Category: req.Category,
	}

	products, err := s.listHandler.Handle(q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	protoProducts := make([]*pb.Product, len(products))
	for i, product := range products {
		protoProducts[i] = domainProductToProto(&product)
	}

	return &pb.ListProductsResponse{
		Products: protoProducts,
		Total:    int32(len(products)),
	}, nil
}

// UpdateStock updates product stock
func (s *ProductServer) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	cmd := command.UpdateStockCommand{
		ProductID: uint(req.ProductId),
		Stock:     int(req.Stock),
	}

	if err := s.updateStockHandler.Handle(cmd); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to update stock: %v", err)
	}

	return &pb.UpdateStockResponse{
		Message:  "Stock updated successfully",
		NewStock: req.Stock,
	}, nil
}

// CheckAvailability checks product availability
func (s *ProductServer) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	q := query.GetProductQuery{ID: uint(req.ProductId)}

	product, err := s.getProductHandler.Handle(q)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	available := product.Stock >= int(req.Quantity) && product.IsActive
	message := "Product is available"
	if !available {
		if !product.IsActive {
			message = "Product is not active"
		} else {
			message = fmt.Sprintf("Insufficient stock. Requested: %d, Available: %d", req.Quantity, product.Stock)
		}
	}

	return &pb.CheckAvailabilityResponse{
		Available:    available,
		CurrentStock: int32(product.Stock),
		Message:      message,
	}, nil
}

// GetStats returns product statistics
func (s *ProductServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.StatsResponse, error) {
	q := query.GetStatsQuery{}

	stats, err := s.statsHandler.Handle(q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	return &pb.StatsResponse{
		TotalProducts:        stats.TotalProducts,
		ActiveProducts:       stats.ActiveProducts,
		OutOfStock:           stats.OutOfStock,
		LowStock:             stats.LowStock,
		ProductsByCategory:   stats.ProductsByCategory,
	}, nil
}

// Helper function to convert domain product to proto product
func domainProductToProto(product *domain.Product) *pb.Product {
	return &pb.Product{
		Id:          uint32(product.ID),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int32(product.Stock),
		Category:    product.Category,
		Sku:         product.SKU,
		IsActive:    product.IsActive,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}
}

