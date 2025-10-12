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
	
	// Business metrics
	outOfStockProducts prometheus.Gauge
	lowStockProducts   prometheus.Gauge
	productsByCategory *prometheus.GaugeVec
	productErrors      *prometheus.CounterVec
	stockUpdates       *prometheus.CounterVec
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

	// Business-specific metrics
	outOfStockProducts := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "product_service_out_of_stock_total",
			Help: "Number of products that are out of stock",
		},
	)

	lowStockProducts := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "product_service_low_stock_total",
			Help: "Number of products with low stock (stock <= 10)",
		},
	)

	productsByCategory := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "product_service_products_by_category",
			Help: "Number of products per category",
		},
		[]string{"category"},
	)

	productErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_service_errors_total",
			Help: "Total number of product operation errors",
		},
		[]string{"operation", "error_type"},
	)

	stockUpdates := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_service_stock_updates_total",
			Help: "Total number of stock update operations",
		},
		[]string{"operation", "status"}, // operation: increase/decrease, status: success/failed
	)

	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(requestSummary)
	prometheus.MustRegister(totalProducts)
	prometheus.MustRegister(outOfStockProducts)
	prometheus.MustRegister(lowStockProducts)
	prometheus.MustRegister(productsByCategory)
	prometheus.MustRegister(productErrors)
	prometheus.MustRegister(stockUpdates)

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
		outOfStockProducts: outOfStockProducts,
		lowStockProducts:   lowStockProducts,
		productsByCategory: productsByCategory,
		productErrors:      productErrors,
		stockUpdates:       stockUpdates,
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
		h.productErrors.WithLabelValues("create", "validation_error").Inc()
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.updateBusinessMetrics()

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
		h.productErrors.WithLabelValues("update", "validation_error").Inc()
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.updateBusinessMetrics()

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
		h.productErrors.WithLabelValues("delete", "not_found").Inc()
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.updateBusinessMetrics()

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

	// Determine operation type (increase/decrease)
	currentProduct, err := h.getProductHandler.Handle(query.GetProductQuery{ID: uint(id)})
	operation := "set"
	if err == nil {
		if req.Stock > currentProduct.Stock {
			operation = "increase"
		} else if req.Stock < currentProduct.Stock {
			operation = "decrease"
		}
	}

	if err := h.updateStockHandler.Handle(cmd); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to update stock")
		h.stockUpdates.WithLabelValues(operation, "failed").Inc()
		respondJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.stockUpdates.WithLabelValues(operation, "success").Inc()
	h.updateBusinessMetrics()

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

// updateBusinessMetrics updates all business-specific metrics
func (h *ProductHandler) updateBusinessMetrics() {
	// Get all products to calculate metrics
	products, err := h.repo.FindAll(10000, 0)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to fetch products for metrics")
		return
	}

	var outOfStock, lowStock int64
	categoryCount := make(map[string]int64)
	const lowStockThreshold = 10

	for _, product := range products {
		// Count out of stock products
		if product.Stock == 0 {
			outOfStock++
		}
		
		// Count low stock products
		if product.Stock > 0 && product.Stock <= lowStockThreshold {
			lowStock++
		}

		// Count products by category
		if product.Category != "" {
			categoryCount[product.Category]++
		}
	}

	// Update gauges
	h.outOfStockProducts.Set(float64(outOfStock))
	h.lowStockProducts.Set(float64(lowStock))
	
	// Update category counts
	for category, count := range categoryCount {
		h.productsByCategory.WithLabelValues(category).Set(float64(count))
	}
	
	// Update total products
	h.totalProducts.Set(float64(len(products)))
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
