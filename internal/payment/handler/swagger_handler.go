package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterSwaggerDocs registers Swagger documentation routes
// @Summary Swagger documentation
// @Description Swagger API documentation
// @Tags Swagger
// @Success 200 {string} string "Swagger UI"
// @Router /swagger/ [get]
func RegisterSwaggerDocs(router *mux.Router, swaggerHandler http.Handler) {
	// Swagger UI
	router.PathPrefix("/swagger/").Handler(swaggerHandler)
}

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a new payment with product purchase (Authenticated users)
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{user_id=int,product_id=int,quantity=int,amount=number,currency=string,payment_method=string} true "Payment data"
// @Success 201 {object} object{success=bool,message=string,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 401 {object} object{success=bool,error=string}
// @Failure 503 {object} object{success=bool,error=string}
// @Router /api/payments [post]
func (h *PaymentHandler) CreatePaymentDoc() {}

// GetPayment godoc
// @Summary Get payment by ID
// @Description Get a specific payment by its ID (Admin only)
// @Tags Payments
// @Security BearerAuth
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} object{success=bool,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Failure 404 {object} object{success=bool,error=string}
// @Router /api/payments/{id} [get]
func (h *PaymentHandler) GetPaymentDoc() {}

// ListPayments godoc
// @Summary List all payments
// @Description Get a list of all payments with pagination (Admin only)
// @Tags Payments
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} object{success=bool,data=object{payments=array,total=int}}
// @Failure 403 {object} object{success=bool,error=string}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/payments [get]
func (h *PaymentHandler) ListPaymentsDoc() {}

// UpdatePaymentStatus godoc
// @Summary Update payment status
// @Description Update the status of a payment (Admin only)
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body object{status=string} true "Status data (pending/completed/failed/refunded)"
// @Success 200 {object} object{success=bool,message=string}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Router /api/payments/{id}/status [patch]
func (h *PaymentHandler) UpdatePaymentStatusDoc() {}

// GetMyPayments godoc
// @Summary Get my payments
// @Description Get payments for the authenticated user
// @Tags Payments
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} object{success=bool,data=object{payments=array,total=int}}
// @Failure 401 {object} object{success=bool,error=string}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/payments/my [get]
func (h *PaymentHandler) GetMyPaymentsDoc() {}

// HealthCheck godoc
// @Summary Health check
// @Description Check service health and database connectivity
// @Tags Health
// @Produce json
// @Success 200 {object} object{success=bool,message=string}
// @Failure 503 {object} object{success=bool,error=string}
// @Router /health [get]
func (h *PaymentHandler) HealthCheckDoc() {}

