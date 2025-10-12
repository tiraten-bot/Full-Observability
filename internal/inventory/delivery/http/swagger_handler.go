package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterSwaggerDocs registers Swagger documentation routes
// @Summary Swagger documentation
// @Description Swagger API documentation for Inventory Service
// @Tags Swagger
// @Success 200 {string} string "Swagger UI"
// @Router /swagger/ [get]
func RegisterSwaggerDocs(router *mux.Router, swaggerHandler http.Handler) {
	// Swagger UI
	router.PathPrefix("/swagger/").Handler(swaggerHandler)
}

// CreateInventory godoc
// @Summary Create new inventory
// @Description Create a new inventory record (Admin only)
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{product_id=int,quantity=int,location=string} true "Inventory data"
// @Success 201 {object} object{success=bool,message=string,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Router /api/inventory [post]
func (h *InventoryHandler) CreateInventoryDoc() {}

// ListInventory godoc
// @Summary List all inventory
// @Description Get a list of all inventory records with pagination
// @Tags Inventory
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} object{success=bool,data=array}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/inventory [get]
func (h *InventoryHandler) ListInventoryDoc() {}

// GetInventory godoc
// @Summary Get inventory by ID
// @Description Get a specific inventory record by its ID
// @Tags Inventory
// @Produce json
// @Param id path int true "Inventory ID"
// @Success 200 {object} object{success=bool,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 404 {object} object{success=bool,error=string}
// @Router /api/inventory/{id} [get]
func (h *InventoryHandler) GetInventoryDoc() {}

// UpdateQuantity godoc
// @Summary Update inventory quantity
// @Description Update the quantity of an inventory record (Admin only)
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param request body object{quantity=int} true "Quantity data"
// @Success 200 {object} object{success=bool,message=string}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Router /api/inventory/{product_id}/quantity [patch]
func (h *InventoryHandler) UpdateQuantityDoc() {}

// GetByProductID godoc
// @Summary Get inventory by product ID
// @Description Get inventory record for a specific product (Authenticated users)
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param product_id path int true "Product ID"
// @Success 200 {object} object{success=bool,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 401 {object} object{success=bool,error=string}
// @Failure 404 {object} object{success=bool,error=string}
// @Router /api/inventory/product/{product_id} [get]
func (h *InventoryHandler) GetByProductIDDoc() {}

// CheckAvailability godoc
// @Summary Check product availability
// @Description Check if a product is available in the requested quantity (Authenticated users)
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param product_id path int true "Product ID"
// @Param quantity query int false "Requested quantity (default: 1)"
// @Success 200 {object} object{success=bool,data=object{product_id=int,available=bool,quantity=int,requested=int,location=string,message=string}}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 401 {object} object{success=bool,error=string}
// @Failure 404 {object} object{success=bool,error=string}
// @Router /api/inventory/check/{product_id} [get]
func (h *InventoryHandler) CheckAvailabilityDoc() {}

// HealthCheck godoc
// @Summary Health check
// @Description Check service health and database connectivity
// @Tags Health
// @Produce json
// @Success 200 {object} object{success=bool,message=string}
// @Failure 503 {object} object{success=bool,error=string}
// @Router /health [get]
func (h *InventoryHandler) HealthCheckDoc() {}
