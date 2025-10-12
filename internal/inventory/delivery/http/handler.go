package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tair/full-observability/internal/inventory/client"
	"github.com/tair/full-observability/internal/inventory/domain"
	"github.com/tair/full-observability/internal/inventory/usecase/command"
	"github.com/tair/full-observability/internal/inventory/usecase/query"
	"github.com/tair/full-observability/pkg/logger"
)

// InventoryHandler handles HTTP requests for inventory using CQRS pattern
type InventoryHandler struct {
	// Command handlers
	createHandler        *command.CreateInventoryHandler
	updateQuantityHandler *command.UpdateQuantityHandler
	deleteHandler        *command.DeleteInventoryHandler

	// Query handlers
	getHandler  *query.GetInventoryHandler
	listHandler *query.ListInventoryHandler

	repo       domain.InventoryRepository
	userClient *client.UserServiceClient
}

// NewInventoryHandler creates a new inventory handler (manual DI)
func NewInventoryHandler(repo domain.InventoryRepository, userClient *client.UserServiceClient) *InventoryHandler {
	return &InventoryHandler{
		createHandler:        command.NewCreateInventoryHandler(repo),
		updateQuantityHandler: command.NewUpdateQuantityHandler(repo),
		deleteHandler:        command.NewDeleteInventoryHandler(repo),
		getHandler:          query.NewGetInventoryHandler(repo),
		listHandler:         query.NewListInventoryHandler(repo),
		repo:                repo,
		userClient:          userClient,
	}
}

// NewInventoryHandlerWithDI creates a new inventory handler using dependency injection
func NewInventoryHandlerWithDI(
	createHandler *command.CreateInventoryHandler,
	updateQuantityHandler *command.UpdateQuantityHandler,
	deleteHandler *command.DeleteInventoryHandler,
	getHandler *query.GetInventoryHandler,
	listHandler *query.ListInventoryHandler,
	repo domain.InventoryRepository,
	userClient *client.UserServiceClient,
) *InventoryHandler {
	return &InventoryHandler{
		createHandler:        createHandler,
		updateQuantityHandler: updateQuantityHandler,
		deleteHandler:        deleteHandler,
		getHandler:          getHandler,
		listHandler:         listHandler,
		repo:                repo,
		userClient:          userClient,
	}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CreateInventory handles POST /api/inventory
func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID uint   `json:"product_id"`
		Quantity  int    `json:"quantity"`
		Location  string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	cmd := command.CreateInventoryCommand{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Location:  req.Location,
	}

	inventory, err := h.createHandler.Handle(cmd)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create inventory")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: "Inventory created successfully",
		Data:    inventory,
	})
}

// GetInventory handles GET /api/inventory/{id}
func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid inventory ID",
		})
		return
	}

	q := query.GetInventoryQuery{ID: uint(id)}
	inventory, err := h.getHandler.Handle(q)
	if err != nil {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Inventory not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    inventory,
	})
}

// ListInventory handles GET /api/inventory
func (h *InventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	q := query.ListInventoryQuery{
		Limit:  limit,
		Offset: offset,
	}

	inventories, err := h.listHandler.Handle(q)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to list inventories")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to list inventories",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    inventories,
	})
}

// UpdateQuantity handles PATCH /api/inventory/{product_id}/quantity
func (h *InventoryHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseUint(vars["product_id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid product ID",
		})
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	cmd := command.UpdateQuantityCommand{
		ProductID: uint(productID),
		Quantity:  req.Quantity,
	}

	if err := h.updateQuantityHandler.Handle(cmd); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update quantity")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Quantity updated successfully",
	})
}

// GetByProductID handles GET /api/inventory/product/{product_id} (authenticated user)
func (h *InventoryHandler) GetByProductID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseUint(vars["product_id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid product ID",
		})
		return
	}

	inventory, err := h.repo.FindByProductID(uint(productID))
	if err != nil {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Inventory not found for this product",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    inventory,
	})
}

// CheckAvailability handles GET /api/inventory/check/{product_id} (authenticated user)
func (h *InventoryHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseUint(vars["product_id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid product ID",
		})
		return
	}

	// Get requested quantity from query param (default: 1)
	requestedQty, _ := strconv.Atoi(r.URL.Query().Get("quantity"))
	if requestedQty <= 0 {
		requestedQty = 1
	}

	inventory, err := h.repo.FindByProductID(uint(productID))
	if err != nil {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Product not found in inventory",
		})
		return
	}

	available := inventory.Quantity >= requestedQty
	message := "Product is available"
	if !available {
		message = fmt.Sprintf("Insufficient quantity. Available: %d, Requested: %d", inventory.Quantity, requestedQty)
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"product_id":  productID,
			"available":   available,
			"quantity":    inventory.Quantity,
			"requested":   requestedQty,
			"location":    inventory.Location,
			"message":     message,
		},
	})
}

// RegisterRoutes registers all inventory routes
func (h *InventoryHandler) RegisterRoutes(router *mux.Router) {
	// Public routes (no auth)
	router.HandleFunc("/api/inventory", h.ListInventory).Methods("GET")
	router.HandleFunc("/api/inventory/{id}", h.GetInventory).Methods("GET")

	// Authenticated user routes (any logged-in user)
	router.HandleFunc("/api/inventory/product/{product_id}", AuthMiddleware(h.userClient)(h.GetByProductID)).Methods("GET")
	router.HandleFunc("/api/inventory/check/{product_id}", AuthMiddleware(h.userClient)(h.CheckAvailability)).Methods("GET")

	// Admin routes (require admin role)
	router.HandleFunc("/api/inventory", AdminMiddleware(h.userClient)(h.CreateInventory)).Methods("POST")
	router.HandleFunc("/api/inventory/{product_id}/quantity", AdminMiddleware(h.userClient)(h.UpdateQuantity)).Methods("PATCH")
}

// GetUserClient returns the user service client
func (h *InventoryHandler) GetUserClient() *client.UserServiceClient {
	return h.userClient
}

// RegisterHealthCheck registers health check endpoint
func (h *InventoryHandler) RegisterHealthCheck(router *mux.Router, db *sql.DB) {
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
			Message: "Inventory service is healthy",
		})
	}).Methods("GET")
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

