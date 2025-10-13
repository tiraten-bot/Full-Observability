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
	"github.com/tair/full-observability/pkg/logger"
)

// PaymentHandler handles HTTP requests for payments using CQRS pattern
type PaymentHandler struct {
	// Command handlers
	createHandler       *command.CreatePaymentHandler
	updateStatusHandler *command.UpdateStatusHandler

	// Query handlers
	getHandler  *query.GetPaymentHandler
	listHandler *query.ListPaymentsHandler

	repo       domain.PaymentRepository
	userClient *client.UserServiceClient
}

// NewPaymentHandler creates a new payment handler (manual DI)
func NewPaymentHandler(repo domain.PaymentRepository, userClient *client.UserServiceClient) *PaymentHandler {
	return &PaymentHandler{
		createHandler:       command.NewCreatePaymentHandler(repo),
		updateStatusHandler: command.NewUpdateStatusHandler(repo),
		getHandler:          query.NewGetPaymentHandler(repo),
		listHandler:         query.NewListPaymentsHandler(repo),
		repo:                repo,
		userClient:          userClient,
	}
}

// NewPaymentHandlerWithDI creates a new payment handler using dependency injection
func NewPaymentHandlerWithDI(
	createHandler *command.CreatePaymentHandler,
	updateStatusHandler *command.UpdateStatusHandler,
	getHandler *query.GetPaymentHandler,
	listHandler *query.ListPaymentsHandler,
	repo domain.PaymentRepository,
	userClient *client.UserServiceClient,
) *PaymentHandler {
	return &PaymentHandler{
		createHandler:       createHandler,
		updateStatusHandler: updateStatusHandler,
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

// CreatePayment handles POST /api/payments
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID        uint    `json:"user_id"`
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

	respondJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: "Payment created successfully",
		Data:    payment,
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

	if limit == 0 {
		limit = 10
	}

	payments, err := h.repo.FindByUserID(userID, limit, offset)
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

// RegisterRoutes registers all payment routes
func (h *PaymentHandler) RegisterRoutes(router *mux.Router) {
	// Public routes (no auth - for demo/testing)
	// In production, you might want to remove these or make them admin-only

	// Authenticated user routes (any logged-in user)
	router.HandleFunc("/api/payments/my", AuthMiddleware(h.userClient)(h.GetMyPayments)).Methods("GET")
	router.HandleFunc("/api/payments", AuthMiddleware(h.userClient)(h.CreatePayment)).Methods("POST")

	// Admin routes (require admin role)
	router.HandleFunc("/api/payments", AdminMiddleware(h.userClient)(h.ListPayments)).Methods("GET")
	router.HandleFunc("/api/payments/{id}", AdminMiddleware(h.userClient)(h.GetPayment)).Methods("GET")
	router.HandleFunc("/api/payments/{id}/status", AdminMiddleware(h.userClient)(h.UpdatePaymentStatus)).Methods("PATCH")
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

