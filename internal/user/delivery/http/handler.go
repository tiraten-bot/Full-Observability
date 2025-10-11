package http

import (
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
	createHandler  *command.CreateUserHandler
	updateHandler  *command.UpdateUserHandler
	deleteHandler  *command.DeleteUserHandler
	getUserHandler *query.GetUserHandler
	listHandler    *query.ListUsersHandler
	repo           domain.UserRepository
	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	activeUsers    prometheus.Gauge
}

// NewUserHandler creates a new user handler
func NewUserHandler(repo domain.UserRepository) *UserHandler {
	// Initialize command handlers
	createHandler := command.NewCreateUserHandler(repo)
	updateHandler := command.NewUpdateUserHandler(repo)
	deleteHandler := command.NewDeleteUserHandler(repo)

	// Initialize query handlers
	getUserHandler := query.NewGetUserHandler(repo)
	listHandler := query.NewListUsersHandler(repo)

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
		createHandler:  createHandler,
		updateHandler:  updateHandler,
		deleteHandler:  deleteHandler,
		getUserHandler: getUserHandler,
		listHandler:    listHandler,
		repo:           repo,
		requestCounter: requestCounter,
		requestLatency: requestLatency,
		activeUsers:    activeUsers,
	}
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

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.CreateUserCommand{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
	}

	user, err := h.createHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.updateActiveUsersMetric()
	h.respondJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	q := query.GetUserQuery{ID: id}
	user, err := h.getUserHandler.Handle(q)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := query.ListUsersQuery{}
	users, err := h.listHandler.Handle(q)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.updateActiveUsersMetric()
	h.respondJSON(w, http.StatusOK, users)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cmd := command.UpdateUserCommand{
		ID:       id,
		Email:    req.Email,
		FullName: req.FullName,
	}

	user, err := h.updateHandler.Handle(cmd)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	cmd := command.DeleteUserCommand{ID: id}
	if err := h.deleteHandler.Handle(cmd); err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.updateActiveUsersMetric()
	h.respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// HealthCheck handles GET /health
func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
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
	router.HandleFunc("/users", h.metricsMiddleware("/users", h.CreateUser)).Methods("POST")
	router.HandleFunc("/users", h.metricsMiddleware("/users", h.ListUsers)).Methods("GET")
	router.HandleFunc("/users/{id}", h.metricsMiddleware("/users/{id}", h.GetUser)).Methods("GET")
	router.HandleFunc("/users/{id}", h.metricsMiddleware("/users/{id}", h.UpdateUser)).Methods("PUT")
	router.HandleFunc("/users/{id}", h.metricsMiddleware("/users/{id}", h.DeleteUser)).Methods("DELETE")
	router.HandleFunc("/health", h.metricsMiddleware("/health", h.HealthCheck)).Methods("GET")
}

