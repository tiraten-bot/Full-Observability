package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tair/full-observability/internal/user/domain"
	"github.com/tair/full-observability/internal/user/usecase/command"
	"github.com/tair/full-observability/internal/user/usecase/query"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	// Command handlers
	registerHandler     *command.RegisterUserHandler
	loginHandler        *command.LoginUserHandler
	updateHandler       *command.UpdateUserHandler
	deleteHandler       *command.DeleteUserHandler
	changeRoleHandler   *command.ChangeRoleHandler
	toggleActiveHandler *command.ToggleActiveHandler

	// Query handlers
	getUserHandler *query.GetUserHandler
	listHandler    *query.ListUsersHandler
	statsHandler   *query.GetStatsHandler

	repo           domain.UserRepository
	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	activeUsers    prometheus.Gauge
}

// NewUserHandler creates a new user handler
func NewUserHandler(repo domain.UserRepository) *UserHandler {
	// Initialize command handlers
	registerHandler := command.NewRegisterUserHandler(repo)
	loginHandler := command.NewLoginUserHandler(repo)
	updateHandler := command.NewUpdateUserHandler(repo)
	deleteHandler := command.NewDeleteUserHandler(repo)
	changeRoleHandler := command.NewChangeRoleHandler(repo)
	toggleActiveHandler := command.NewToggleActiveHandler(repo)

	// Initialize query handlers
	getUserHandler := query.NewGetUserHandler(repo)
	listHandler := query.NewListUsersHandler(repo)
	statsHandler := query.NewGetStatsHandler(repo)

	// Initialize Prometheus metrics
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_requests_total",
			Help: "Total number of requests to user service",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_request_duration_seconds",
			Help:    "Duration of user service requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	activeUsers := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "user_service_active_users",
			Help: "Number of active users in the system",
		},
	)

	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(activeUsers)

	return &UserHandler{
		registerHandler:     registerHandler,
		loginHandler:        loginHandler,
		updateHandler:       updateHandler,
		deleteHandler:       deleteHandler,
		changeRoleHandler:   changeRoleHandler,
		toggleActiveHandler: toggleActiveHandler,
		getUserHandler:      getUserHandler,
		listHandler:         listHandler,
		statsHandler:        statsHandler,
		repo:                repo,
		requestCounter:      requestCounter,
		requestLatency:      requestLatency,
		activeUsers:         activeUsers,
	}
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
func (h *UserHandler) metricsMiddleware(endpoint string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		h.requestLatency.WithLabelValues(r.Method, endpoint).Observe(duration)
		h.requestCounter.WithLabelValues(r.Method, endpoint, strconv.Itoa(rw.statusCode)).Inc()
	}
}

// Register handles POST /auth/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.RegisterUserCommand{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     domain.RoleUser, // Default role
	}

	user, err := h.registerHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.updateActiveUsersMetric()
	h.respondJSON(w, http.StatusCreated, user)
}

// Login handles POST /auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.LoginUserCommand{
		Username: req.Username,
		Password: req.Password,
	}

	response, err := h.loginHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// GetProfile handles GET /users/me (authenticated user)
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	q := query.GetUserQuery{ID: userID}
	user, err := h.getUserHandler.Handle(q)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// UpdateProfile handles PUT /users/me (authenticated user)
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	var req struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.UpdateUserCommand{
		ID:       userID,
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	}

	user, err := h.updateHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// --- ADMIN ENDPOINTS ---

// CreateAdmin handles POST /admin/users (admin only)
func (h *UserHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.RegisterUserCommand{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     req.Role, // Admin can set role
	}

	user, err := h.registerHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.updateActiveUsersMetric()
	h.respondJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /admin/users/{id} (admin only)
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	q := query.GetUserQuery{ID: uint(id)}
	user, err := h.getUserHandler.Handle(q)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// ListUsers handles GET /admin/users (admin only)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	role := r.URL.Query().Get("role")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	q := query.ListUsersQuery{
		Limit:  limit,
		Offset: offset,
		Role:   role,
	}

	users, err := h.listHandler.Handle(q)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.updateActiveUsersMetric()
	h.respondJSON(w, http.StatusOK, users)
}

// UpdateUser handles PUT /admin/users/{id} (admin only)
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.UpdateUserCommand{
		ID:       uint(id),
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	}

	user, err := h.updateHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /admin/users/{id} (admin only)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	cmd := command.DeleteUserCommand{ID: uint(id)}
	if err := h.deleteHandler.Handle(cmd); err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.updateActiveUsersMetric()
	h.respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// ChangeRole handles PUT /admin/users/{id}/role (admin only)
func (h *UserHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.ChangeRoleCommand{
		UserID: uint(id),
		Role:   req.Role,
	}

	user, err := h.changeRoleHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// ToggleActive handles PUT /admin/users/{id}/active (admin only)
func (h *UserHandler) ToggleActive(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.ToggleActiveCommand{
		UserID:   uint(id),
		IsActive: req.IsActive,
	}

	user, err := h.toggleActiveHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// GetStats handles GET /admin/stats (admin only)
func (h *UserHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	q := query.GetStatsQuery{}
	stats, err := h.statsHandler.Handle(q)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// HealthCheck handles GET /health
func (h *UserHandler) HealthCheck(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Check database connectivity
		if err := db.PingContext(ctx); err != nil {
			h.respondJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		h.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	}
}

// updateActiveUsersMetric updates the active users gauge
func (h *UserHandler) updateActiveUsersMetric() {
	count, err := h.repo.Count()
	if err == nil {
		h.activeUsers.Set(float64(count))
	}
}

// respondJSON sends a JSON response
func (h *UserHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *UserHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

// RegisterRoutes registers all user routes
func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	// Public routes
	router.HandleFunc("/auth/register", h.metricsMiddleware("/auth/register", h.Register)).Methods("POST")
	router.HandleFunc("/auth/login", h.metricsMiddleware("/auth/login", h.Login)).Methods("POST")

	// Authenticated user routes
	router.HandleFunc("/users/me", h.metricsMiddleware("/users/me", AuthMiddleware(h.GetProfile))).Methods("GET")
	router.HandleFunc("/users/me", h.metricsMiddleware("/users/me", AuthMiddleware(h.UpdateProfile))).Methods("PUT")

	// Admin routes
	router.HandleFunc("/admin/users", h.metricsMiddleware("/admin/users", AdminMiddleware(h.CreateAdmin))).Methods("POST")
	router.HandleFunc("/admin/users", h.metricsMiddleware("/admin/users", AdminMiddleware(h.ListUsers))).Methods("GET")
	router.HandleFunc("/admin/users/{id}", h.metricsMiddleware("/admin/users/{id}", AdminMiddleware(h.GetUser))).Methods("GET")
	router.HandleFunc("/admin/users/{id}", h.metricsMiddleware("/admin/users/{id}", AdminMiddleware(h.UpdateUser))).Methods("PUT")
	router.HandleFunc("/admin/users/{id}", h.metricsMiddleware("/admin/users/{id}", AdminMiddleware(h.DeleteUser))).Methods("DELETE")
	router.HandleFunc("/admin/users/{id}/role", h.metricsMiddleware("/admin/users/{id}/role", AdminMiddleware(h.ChangeRole))).Methods("PUT")
	router.HandleFunc("/admin/users/{id}/active", h.metricsMiddleware("/admin/users/{id}/active", AdminMiddleware(h.ToggleActive))).Methods("PUT")
	router.HandleFunc("/admin/stats", h.metricsMiddleware("/admin/stats", AdminMiddleware(h.GetStats))).Methods("GET")
}

// RegisterHealthCheck registers health check endpoint
func (h *UserHandler) RegisterHealthCheck(router *mux.Router, db *sql.DB) {
	router.HandleFunc("/health", h.HealthCheck(db)).Methods("GET")
}
