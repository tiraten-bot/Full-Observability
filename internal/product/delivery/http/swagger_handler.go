package http

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

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product (Admin only)
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{name=string,description=string,price=number,stock=int,category=string,sku=string,is_active=bool} true "Product data"
// @Success 201 {object} object{success=bool,message=string,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Router /api/products [post]
func (h *ProductHandler) CreateProductDoc() {}

// ListProducts godoc
// @Summary List all products
// @Description Get a list of all products with pagination
// @Tags Products
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param category query string false "Category filter"
// @Success 200 {object} object{success=bool,data=object{products=array,total=int,limit=int,offset=int}}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/products [get]
func (h *ProductHandler) ListProductsDoc() {}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get a specific product by its ID
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} object{success=bool,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 404 {object} object{success=bool,error=string}
// @Router /api/products/{id} [get]
func (h *ProductHandler) GetProductDoc() {}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product (Admin only)
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body object{name=string,description=string,price=number,stock=int,category=string,sku=string,is_active=bool} true "Product data"
// @Success 200 {object} object{success=bool,message=string,data=object}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Router /api/products/{id} [put]
func (h *ProductHandler) UpdateProductDoc() {}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by ID (Admin only)
// @Tags Products
// @Security BearerAuth
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} object{success=bool,message=string}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Router /api/products/{id} [delete]
func (h *ProductHandler) DeleteProductDoc() {}

// UpdateStock godoc
// @Summary Update product stock
// @Description Update the stock quantity of a product (Admin only)
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body object{stock=int} true "Stock data"
// @Success 200 {object} object{success=bool,message=string}
// @Failure 400 {object} object{success=bool,error=string}
// @Failure 403 {object} object{success=bool,error=string}
// @Router /api/products/{id}/stock [patch]
func (h *ProductHandler) UpdateStockDoc() {}

// GetStats godoc
// @Summary Get product statistics
// @Description Get product statistics (total, by category, etc.)
// @Tags Products
// @Produce json
// @Success 200 {object} object{success=bool,data=object}
// @Failure 500 {object} object{success=bool,error=string}
// @Router /api/products/stats [get]
func (h *ProductHandler) GetStatsDoc() {}

// HealthCheck godoc
// @Summary Health check
// @Description Check service health and database connectivity
// @Tags Health
// @Produce json
// @Success 200 {object} object{success=bool,message=string}
// @Failure 503 {object} object{success=bool,error=string}
// @Router /health [get]
func (h *ProductHandler) HealthCheckDoc() {}
