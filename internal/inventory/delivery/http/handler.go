package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tair/full-observability/internal/inventory/domain"
	"github.com/tair/full-observability/pkg/logger"
)

// InventoryHandler handles HTTP requests for inventory
type InventoryHandler struct {
	repo domain.InventoryRepository
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(repo domain.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{repo: repo}
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

	inventory := &domain.Inventory{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Location:  req.Location,
	}

	if err := h.repo.Create(inventory); err != nil {
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

	inventory, err := h.repo.FindByID(uint(id))
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

	if limit == 0 {
		limit = 10
	}

	inventories, err := h.repo.FindAll(limit, offset)
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

	if err := h.repo.UpdateQuantity(uint(productID), req.Quantity); err != nil {
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

// RegisterRoutes registers all inventory routes
func (h *InventoryHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/inventory", h.ListInventory).Methods("GET")
	router.HandleFunc("/api/inventory", h.CreateInventory).Methods("POST")
	router.HandleFunc("/api/inventory/{id}", h.GetInventory).Methods("GET")
	router.HandleFunc("/api/inventory/{product_id}/quantity", h.UpdateQuantity).Methods("PATCH")
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

