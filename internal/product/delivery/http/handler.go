package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tair/full-observability/internal/product/domain"
	"github.com/tair/full-observability/internal/product/repository"
	"github.com/tair/full-observability/pkg/logger"
)

type ProductHandler struct {
	repo *repository.GormProductRepository
}

func NewProductHandler(repo *repository.GormProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (h *ProductHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/products", h.CreateProduct).Methods("POST")
	router.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	router.HandleFunc("/api/products/{id}", h.GetProduct).Methods("GET")
	router.HandleFunc("/api/products/{id}", h.UpdateProduct).Methods("PUT")
	router.HandleFunc("/api/products/{id}", h.DeleteProduct).Methods("DELETE")
	router.HandleFunc("/api/products/{id}/stock", h.UpdateStock).Methods("PATCH")
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if err := h.repo.Create(&product); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create product")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to create product",
		})
		return
	}

	respondJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: "Product created successfully",
		Data:    product,
	})
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")

	if limit <= 0 {
		limit = 10
	}

	var products []domain.Product
	var err error

	if category != "" {
		products, err = h.repo.FindByCategory(category, limit, offset)
	} else {
		products, err = h.repo.FindAll(limit, offset)
	}

	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to list products")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to list products",
		})
		return
	}

	count, _ := h.repo.Count()

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"products": products,
			"total":    count,
			"limit":    limit,
			"offset":   offset,
		},
	})
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid product ID",
		})
		return
	}

	product, err := h.repo.FindByID(uint(id))
	if err != nil {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Product not found",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    product,
	})
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid product ID",
		})
		return
	}

	product, err := h.repo.FindByID(uint(id))
	if err != nil {
		respondJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Product not found",
		})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(product); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	product.ID = uint(id) // Preserve the original ID

	if err := h.repo.Update(product); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update product")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to update product",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Product updated successfully",
		Data:    product,
	})
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid product ID",
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to delete product")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to delete product",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Product deleted successfully",
	})
}

func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid product ID",
		})
		return
	}

	var req struct {
		Stock int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if err := h.repo.UpdateStock(uint(id), req.Stock); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update stock")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to update stock",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Stock updated successfully",
	})
}

func (h *ProductHandler) RegisterHealthCheck(router *mux.Router, db *sql.DB) {
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
			Message: "Product service is healthy",
		})
	}).Methods("GET")
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

