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

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object{username=string,email=string,password=string,full_name=string} true "User registration data"
// @Success 201 {object} object{id=int,username=string,email=string,full_name=string,role=string,is_active=bool,created_at=string,updated_at=string}
// @Failure 400 {object} object{error=string}
// @Router /auth/register [post]
func (h *UserHandler) RegisterDoc() {}

// Login godoc
// @Summary User login
// @Description Authenticate user and get JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object{username=string,password=string} true "Login credentials"
// @Success 200 {object} object{token=string,user=object}
// @Failure 401 {object} object{error=string}
// @Router /auth/login [post]
func (h *UserHandler) LoginDoc() {}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get authenticated user's profile information
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{id=int,username=string,email=string,full_name=string,role=string,is_active=bool}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /users/me [get]
func (h *UserHandler) GetProfileDoc() {}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Update authenticated user's profile
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{email=string,full_name=string,password=string} true "Update data"
// @Success 200 {object} object{id=int,username=string,email=string,full_name=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Router /users/me [put]
func (h *UserHandler) UpdateProfileDoc() {}

// CreateAdmin godoc
// @Summary Create user (admin)
// @Description Admin endpoint to create a new user with specified role
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{username=string,email=string,password=string,full_name=string,role=string} true "User data"
// @Success 201 {object} object{id=int,username=string,email=string,full_name=string,role=string}
// @Failure 400 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Router /admin/users [post]
func (h *UserHandler) CreateAdminDoc() {}

// ListUsers godoc
// @Summary List all users (admin)
// @Description Admin endpoint to list all users with pagination and filtering
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param role query string false "Role filter (user/admin)"
// @Success 200 {array} object{id=int,username=string,email=string,full_name=string,role=string}
// @Failure 403 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /admin/users [get]
func (h *UserHandler) ListUsersDoc() {}

// GetUser godoc
// @Summary Get user by ID (admin)
// @Description Admin endpoint to get specific user details
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} object{id=int,username=string,email=string,full_name=string,role=string}
// @Failure 400 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUserDoc() {}

// UpdateUser godoc
// @Summary Update user (admin)
// @Description Admin endpoint to update user information
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body object{email=string,full_name=string,password=string} true "Update data"
// @Success 200 {object} object{id=int,username=string,email=string,full_name=string}
// @Failure 400 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Router /admin/users/{id} [put]
func (h *UserHandler) UpdateUserDoc() {}

// DeleteUser godoc
// @Summary Delete user (admin)
// @Description Admin endpoint to delete a user (soft delete)
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} object{message=string}
// @Failure 403 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /admin/users/{id} [delete]
func (h *UserHandler) DeleteUserDoc() {}

// ChangeRole godoc
// @Summary Change user role (admin)
// @Description Admin endpoint to change user's role
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body object{role=string} true "New role"
// @Success 200 {object} object{id=int,username=string,role=string}
// @Failure 400 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Router /admin/users/{id}/role [put]
func (h *UserHandler) ChangeRoleDoc() {}

// ToggleActive godoc
// @Summary Toggle user active status (admin)
// @Description Admin endpoint to activate/deactivate user
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body object{is_active=bool} true "Active status"
// @Success 200 {object} object{id=int,username=string,is_active=bool}
// @Failure 400 {object} object{error=string}
// @Failure 403 {object} object{error=string}
// @Router /admin/users/{id}/active [put]
func (h *UserHandler) ToggleActiveDoc() {}

// GetStats godoc
// @Summary Get user statistics (admin)
// @Description Admin endpoint to get user statistics
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{total_users=int,admin_count=int,user_count=int,active_users=int}
// @Failure 403 {object} object{error=string}
// @Router /admin/stats [get]
func (h *UserHandler) GetStatsDoc() {}

// HealthCheck godoc
// @Summary Health check
// @Description Check service health and database connectivity
// @Tags Health
// @Produce json
// @Success 200 {object} object{status=string}
// @Failure 503 {object} object{status=string,error=string}
// @Router /health [get]
func (h *UserHandler) HealthCheckDoc() {}

