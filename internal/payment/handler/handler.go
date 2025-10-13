package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tair/full-observability/internal/payment/domain"
	"github.com/tair/full-observability/pkg/logger"
)

// PaymentHandler handles HTTP requests for payments
type PaymentHandler struct {
	repo domain.PaymentRepository
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(repo domain.PaymentRepository) *PaymentHandler {
	return &PaymentHandler{repo: repo}
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

	// Generate unique order ID
	orderID := fmt.Sprintf("ORD-%s", uuid.New().String()[:8])

	payment := &domain.Payment{
		UserID:        req.UserID,
		OrderID:       orderID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        domain.StatusPending,
		PaymentMethod: req.PaymentMethod,
		TransactionID: fmt.Sprintf("TXN-%s", uuid.New().String()[:12]),
	}

	if err := h.repo.Create(payment); err != nil {
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

	payment, err := h.repo.FindByID(uint(id))
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

	if limit == 0 {
		limit = 10
	}

	payments, err := h.repo.FindAll(limit, offset)
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

	if err := h.repo.UpdateStatus(uint(id), req.Status); err != nil {
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

// RegisterRoutes registers all payment routes
func (h *PaymentHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/payments", h.ListPayments).Methods("GET")
	router.HandleFunc("/api/payments", h.CreatePayment).Methods("POST")
	router.HandleFunc("/api/payments/{id}", h.GetPayment).Methods("GET")
	router.HandleFunc("/api/payments/{id}/status", h.UpdatePaymentStatus).Methods("PATCH")
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

