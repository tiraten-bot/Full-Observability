package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tair/full-observability/internal/product/domain"
	"github.com/tair/full-observability/internal/product/usecase/command"
	"github.com/tair/full-observability/internal/product/usecase/query"
	"github.com/tair/full-observability/pkg/logger"
)

// ProductHandler handles HTTP requests for products using CQRS pattern
type ProductHandler struct {
	// Command handlers
	createHandler      *command.CreateProductHandler
	updateHandler      *command.UpdateProductHandler
	deleteHandler      *command.DeleteProductHandler
	updateStockHandler *command.UpdateStockHandler

	// Query handlers
	getProductHandler  *query.GetProductHandler
	listHandler        *query.ListProductsHandler
	statsHandler       *query.GetStatsHandler

	repo           domain.ProductRepository
	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	requestSummary *prometheus.SummaryVec
	totalProducts  prometheus.Gauge
}

// NewProductHandler creates a new product handler with CQRS pattern (manual DI for backwards compatibility)
func NewProductHandler(repo domain.ProductRepository) *ProductHandler {
	// Initialize command handlers
	createHandler := command.NewCreateProductHandler(repo)
	updateHandler := command.NewUpdateProductHandler(repo)
	deleteHandler := command.NewDeleteProductHandler(repo)
	updateStockHandler := command.NewUpdateStockHandler(repo)

	// Initialize query handlers
	getProductHandler := query.NewGetProductHandler(repo)
	listHandler := query.NewListProductsHandler(repo)
	statsHandler := query.NewGetStatsHandler(repo)

	return newProductHandler(
		createHandler, updateHandler, deleteHandler, updateStockHandler,
		getProductHandler, listHandler, statsHandler,
		repo,
	)
}

// NewProductHandlerWithDI creates a new product handler using dependency injection
// This is used by Wire for automatic dependency injection
func NewProductHandlerWithDI(
	createHandler *command.CreateProductHandler,
	updateHandler *command.UpdateProductHandler,
	deleteHandler *command.DeleteProductHandler,
	updateStockHandler *command.UpdateStockHandler,
	getProductHandler *query.GetProductHandler,
	listHandler *query.ListProductsHandler,
	statsHandler *query.GetStatsHandler,
	repo domain.ProductRepository,
) *ProductHandler {
	return newProductHandler(
		createHandler, updateHandler, deleteHandler, updateStockHandler,
		getProductHandler, listHandler, statsHandler,
		repo,
	)
}

// newProductHandler is the internal constructor used by both manual and Wire DI
func newProductHandler(
	createHandler *command.CreateProductHandler,
	updateHandler *command.UpdateProductHandler,
	deleteHandler *command.DeleteProductHandler,
	updateStockHandler *command.UpdateStockHandler,
	getProductHandler *query.GetProductHandler,
	listHandler *query.ListProductsHandler,
	statsHandler *query.GetStatsHandler,
	repo domain.ProductRepository,
) *ProductHandler {
	// Initialize Prometheus metrics
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_service_requests_total",
			Help: "Total number of requests to product service",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_service_request_duration_seconds",
			Help:    "Duration of product service requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Summary metric for percentile calculation (p50, p90, p95, p99)
	requestSummary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "product_service_request_duration_summary",
			Help: "Summary of request durations with percentiles (client-side quantiles)",
			Objectives: map[float64]float64{
				0.5:  0.05,  // p50 (median) with 5% error
				0.9:  0.01,  // p90 with 1% error
				0.95: 0.01,  // p95 with 1% error
				0.99: 0.001, // p99 with 0.1% error
			},
			MaxAge: 10 * time.Minute, // Keep data for 10 minutes
		},
		[]string{"method", "endpoint"},
	)

	totalProducts := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "product_service_total_products",
			Help: "Total number of products in the system",
		},
	)

	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(requestSummary)
	prometheus.MustRegister(totalProducts)

	return &ProductHandler{
		createHandler:      createHandler,
		updateHandler:      updateHandler,
		deleteHandler:      deleteHandler,
		updateStockHandler: updateStockHandler,
		getProductHandler:  getProductHandler,
		listHandler:        listHandler,
		statsHandler:       statsHandler,
		repo:               repo,
		requestCounter:     requestCounter,
		requestLatency:     requestLatency,
		requestSummary:     requestSummary,
		totalProducts:      totalProducts,
	}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// metricsMiddleware wraps handlers with Prometheus metrics
func (h *ProductHandler) metricsMiddleware(endpoint string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		
		// Record metrics
		h.requestCounter.WithLabelValues(r.Method, endpoint, strconv.Itoa(rw.statusCode)).Inc()
		h.requestLatency.WithLabelValues(r.Method, endpoint).Observe(duration)
		h.requestSummary.WithLabelValues(r.Method, endpoint).Observe(duration)
	}
}

func (h *ProductHandler) RegisterRoutes(router *mux.Router) {
	// Public routes (no auth required)
	router.HandleFunc("/api/products", h.metricsMiddleware("/api/products", h.ListProducts)).Methods("GET")
	router.HandleFunc("/api/products/stats", h.metricsMiddleware("/api/products/stats", h.GetStats)).Methods("GET")
	router.HandleFunc("/api/products/{id}", h.metricsMiddleware("/api/products/{id}", h.GetProduct)).Methods("GET")

	// Admin routes (admin role required)
	router.HandleFunc("/api/products", h.metricsMiddleware("/api/products", AdminMiddleware(h.CreateProduct))).Methods("POST")
	router.HandleFunc("/api/products/{id}", h.metricsMiddleware("/api/products/{id}", AdminMiddleware(h.UpdateProduct))).Methods("PUT")
	router.HandleFunc("/api/products/{id}", h.metricsMiddleware("/api/products/{id}", AdminMiddleware(h.DeleteProduct))).Methods("DELETE")
	router.HandleFunc("/api/products/{id}/stock", h.metricsMiddleware("/api/products/{id}/stock", AdminMiddleware(h.UpdateStock))).Methods("PATCH")
}

// CreateProduct handles POST /api/products
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Category    string  `json:"category"`
		SKU         string  `json:"sku"`
		IsActive    bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	cmd := command.CreateProductCommand{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		SKU:         req.SKU,
		IsActive:    req.IsActive,
	}

	product, err := h.createHandler.Handle(cmd)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create product")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.updateProductsMetric()

	respondJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: "Product created successfully",
		Data:    product,
	})
}

// ListProducts handles GET /api/products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")

	q := query.ListProductsQuery{
		Limit:    limit,
		Offset:   offset,
		Category: category,
	}

	products, err := h.listHandler.Handle(q)
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
			"limit":    q.Limit,
			"offset":   offset,
		},
	})
}

// GetProduct handles GET /api/products/{id}
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

	q := query.GetProductQuery{ID: uint(id)}
	product, err := h.getProductHandler.Handle(q)
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

// UpdateProduct handles PUT /api/products/{id}
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

	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Category    string  `json:"category"`
		SKU         string  `json:"sku"`
		IsActive    bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	cmd := command.UpdateProductCommand{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		SKU:         req.SKU,
		IsActive:    req.IsActive,
	}

	product, err := h.updateHandler.Handle(cmd)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update product")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Product updated successfully",
		Data:    product,
	})
}

// DeleteProduct handles DELETE /api/products/{id}
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

	cmd := command.DeleteProductCommand{ID: uint(id)}
	if err := h.deleteHandler.Handle(cmd); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to delete product")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.updateProductsMetric()

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Product deleted successfully",
	})
}

// UpdateStock handles PATCH /api/products/{id}/stock
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

	cmd := command.UpdateStockCommand{
		ProductID: uint(id),
		Stock:     req.Stock,
	}

	if err := h.updateStockHandler.Handle(cmd); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update stock")
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Stock updated successfully",
	})
}

// GetStats handles GET /api/products/stats
func (h *ProductHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	q := query.GetStatsQuery{}
	stats, err := h.statsHandler.Handle(q)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get stats")
		respondJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to get statistics",
		})
		return
	}

	respondJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    stats,
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

// updateProductsMetric updates the total products gauge
func (h *ProductHandler) updateProductsMetric() {
	count, err := h.repo.Count()
	if err == nil {
		h.totalProducts.Set(float64(count))
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
