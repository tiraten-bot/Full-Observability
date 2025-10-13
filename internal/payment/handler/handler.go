package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tair/full-observability/internal/payment/client"
	"github.com/tair/full-observability/internal/payment/domain"
	"github.com/tair/full-observability/internal/payment/usecase/command"
	"github.com/tair/full-observability/internal/payment/usecase/query"
	"github.com/tair/full-observability/kafka"
	"github.com/tair/full-observability/pkg/logger"
)

// PaymentHandler handles HTTP requests for payments using CQRS pattern
type PaymentHandler struct {
	// Command handlers
	createHandler       *command.CreatePaymentHandler
	updateStatusHandler *command.UpdateStatusHandler

	// Query handlers
	getHandler       *query.GetPaymentHandler
	listHandler      *query.ListPaymentsHandler
	getMyHandler     *query.GetMyPaymentsHandler

	repo             domain.PaymentRepository
	userClient       *client.UserServiceClient
	productClient    *client.ProductServiceClient
	inventoryClient  *client.InventoryServiceClient
	kafkaPublisher   *kafka.Publisher
}

// NewPaymentHandler creates a new payment handler (manual DI)
func NewPaymentHandler(repo domain.PaymentRepository, userClient *client.UserServiceClient, productClient *client.ProductServiceClient, inventoryClient *client.InventoryServiceClient) *PaymentHandler {
	return &PaymentHandler{
		createHandler:       command.NewCreatePaymentHandler(repo),
		updateStatusHandler: command.NewUpdateStatusHandler(repo),
		getHandler:          query.NewGetPaymentHandler(repo),
		listHandler:         query.NewListPaymentsHandler(repo),
		getMyHandler:        query.NewGetMyPaymentsHandler(repo),
		repo:                repo,
		userClient:          userClient,
		productClient:       productClient,
		inventoryClient:     inventoryClient,
	}
}

// NewPaymentHandlerWithDI creates a new payment handler using dependency injection
func NewPaymentHandlerWithDI(
	createHandler *command.CreatePaymentHandler,
	updateStatusHandler *command.UpdateStatusHandler,
	getHandler *query.GetPaymentHandler,
	listHandler *query.ListPaymentsHandler,
	getMyHandler *query.GetMyPaymentsHandler,
	repo domain.PaymentRepository,
	userClient *client.UserServiceClient,
	productClient *client.ProductServiceClient,
	inventoryClient *client.InventoryServiceClient,
	kafkaPublisher *kafka.Publisher,
) *PaymentHandler {
	return &PaymentHandler{
		createHandler:       createHandler,
		updateStatusHandler: updateStatusHandler,
		getHandler:          getHandler,
		listHandler:         listHandler,
		getMyHandler:        getMyHandler,
		repo:                repo,
		userClient:          userClient,
		productClient:       productClient,
		inventoryClient:     inventoryClient,
		kafkaPublisher:      kafkaPublisher,
	}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CreatePayment handles POST /api/payments
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID        uint    `json:"user_id"`
		ProductID     uint    `json:"product_id"`
		Quantity      int32   `json:"quantity"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		PaymentMethod string  `json:"payment_method"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.ProductID == 0 {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Product ID is required",
		})
		return
	}

	if req.Quantity <= 0 {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Quantity must be greater than 0",
		})
		return
	}

	// Check product stock availability via Inventory Service gRPC
	ctx := r.Context()
	available, currentStock, message, err := h.inventoryClient.CheckAvailability(ctx, req.ProductID, req.Quantity)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Uint("product_id", req.ProductID).
			Int32("quantity", req.Quantity).
			Msg("Failed to check product availability")
		respondJSON(w, http.StatusServiceUnavailable, Response{
			Success: false,
			Error:   "Unable to verify product availability. Please try again later.",
		})
		return
	}

	if !available {
		logger.Logger.Warn().
			Uint("product_id", req.ProductID).
			Int32("requested_quantity", req.Quantity).
			Int32("current_stock", currentStock).
			Msg("Insufficient stock for payment")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   message,
			Data: map[string]interface{}{
				"product_id":     req.ProductID,
				"requested":      req.Quantity,
				"available":      currentStock,
			},
		})
		return
	}

	logger.Logger.Info().
		Uint("product_id", req.ProductID).
		Int32("quantity", req.Quantity).
		Int32("current_stock", currentStock).
		Msg("Stock validation passed")

	cmd := command.CreatePaymentCommand{
		UserID:        req.UserID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
	}

	payment, err := h.createHandler.Handle(cmd)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create payment")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Publish product purchased event to Kafka (with tracing)
	if h.kafkaPublisher != nil {
		event := kafka.ProductPurchasedEvent{
			PaymentID:     payment.ID,
			ProductID:     req.ProductID,
			Quantity:      req.Quantity,
			UserID:        req.UserID,
			Amount:        req.Amount,
			Currency:      req.Currency,
			PaymentMethod: req.PaymentMethod,
		}

		if err := h.kafkaPublisher.PublishProductPurchased(ctx, event); err != nil {
			logger.Logger.Error().
				Err(err).
				Uint("payment_id", payment.ID).
				Uint("product_id", req.ProductID).
				Msg("Failed to publish product purchased event")
			// Don't fail the payment, just log the error
		} else {
			logger.Logger.Info().
				Uint("payment_id", payment.ID).
				Uint("product_id", req.ProductID).
				Int32("quantity", req.Quantity).
				Msg("Product purchased event published successfully")
		}
	}

	respondJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: "Payment created successfully",
		Data: map[string]interface{}{
			"payment":    payment,
			"product_id": req.ProductID,
			"quantity":   req.Quantity,
			"stock_info": map[string]interface{}{
				"available":      true,
				"current_stock":  currentStock,
				"reserved_stock": req.Quantity,
			},
		},
	})
}

// GetPayment handles GET /api/payments/{id}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid payment ID",
		})
		return
	}

	q := query.GetPaymentQuery{ID: uint(id)}
	payment, err := h.getHandler.Handle(q)
	if err != nil {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Payment not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    payment,
	})
}

// ListPayments handles GET /api/payments
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	q := query.ListPaymentsQuery{
		Limit:  limit,
		Offset: offset,
	}

	payments, err := h.listHandler.Handle(q)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to list payments")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to list payments",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"payments": payments,
			"total":    len(payments),
		},
	})
}

// UpdatePaymentStatus handles PATCH /api/payments/{id}/status
func (h *PaymentHandler) UpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid payment ID",
		})
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	cmd := command.UpdateStatusCommand{
		PaymentID: uint(id),
		Status:    req.Status,
	}

	if err := h.updateStatusHandler.Handle(cmd); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update payment status")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Payment status updated successfully",
	})
}

// GetMyPayments handles GET /api/payments/my (authenticated user)
func (h *PaymentHandler) GetMyPayments(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, Response{
			Success: false,
			Error:   "User ID not found in context",
		})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	q := query.GetMyPaymentsQuery{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	payments, err := h.getMyHandler.Handle(q)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get user payments")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to get payments",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"payments": payments,
			"total":    len(payments),
		},
	})
}

// GetMiddlewareConfig returns middleware configuration
func (h *PaymentHandler) GetMiddlewareConfig() MiddlewareConfig {
	return DefaultMiddlewareConfig(h.userClient)
}

// RegisterRoutes registers all payment routes
func (h *PaymentHandler) RegisterRoutes(router *mux.Router) {
	middlewareConfig := h.GetMiddlewareConfig()

	// Public routes (no auth - for demo/testing)
	// In production, you might want to remove these or make them admin-only

	// Authenticated user routes (any logged-in user)
	router.HandleFunc("/api/payments/my", middlewareConfig.GetAuthMiddleware()(h.GetMyPayments)).Methods("GET")
	router.HandleFunc("/api/payments", middlewareConfig.GetAuthMiddleware()(h.CreatePayment)).Methods("POST")

	// Admin routes (require admin role)
	router.HandleFunc("/api/payments", middlewareConfig.GetAdminMiddleware()(h.ListPayments)).Methods("GET")
	router.HandleFunc("/api/payments/{id}", middlewareConfig.GetAdminMiddleware()(h.GetPayment)).Methods("GET")
	router.HandleFunc("/api/payments/{id}/status", middlewareConfig.GetAdminMiddleware()(h.UpdatePaymentStatus)).Methods("PATCH")
}

// RegisterHealthCheck registers health check endpoint
func (h *PaymentHandler) RegisterHealthCheck(router *mux.Router, db *sql.DB) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			respondJSON(w, http.StatusServiceUnavailable, Response{
				Success: false,
				Error:   "Database unavailable",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Message: "Payment service is healthy",
		})
	}).Methods("GET")
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

