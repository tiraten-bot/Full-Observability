package user

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// Handler handles HTTP requests for users
type Handler struct {
	service        *Service
	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	activeUsers    prometheus.Gauge
}

// NewHandler creates a new user handler with Prometheus metrics
func NewHandler(service *Service) *Handler {
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

	return &Handler{
		service:        service,
		requestCounter: requestCounter,
		requestLatency: requestLatency,
		activeUsers:    activeUsers,
	}
}

// metricsMiddleware wraps handlers with Prometheus metrics
func (h *Handler) metricsMiddleware(endpoint string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
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
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.CreateUser(req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update active users count
	users, _ := h.service.GetAllUsers()
	h.activeUsers.Set(float64(len(users)))

	h.respondJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.service.GetUser(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// GetAllUsers handles GET /users
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update active users count
	h.activeUsers.Set(float64(len(users)))

	h.respondJSON(w, http.StatusOK, users)
}

// UpdateUser handles PUT /users/{id}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.UpdateUser(id, req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.service.DeleteUser(id); err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Update active users count
	users, _ := h.service.GetAllUsers()
	h.activeUsers.Set(float64(len(users)))

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

// RegisterRoutes registers all user routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users", h.metricsMiddleware("/users", h.CreateUser)).Methods("POST")
	router.HandleFunc("/users", h.metricsMiddleware("/users", h.GetAllUsers)).Methods("GET")
	router.HandleFunc("/users/{id}", h.metricsMiddleware("/users/{id}", h.GetUser)).Methods("GET")
	router.HandleFunc("/users/{id}", h.metricsMiddleware("/users/{id}", h.UpdateUser)).Methods("PUT")
	router.HandleFunc("/users/{id}", h.metricsMiddleware("/users/{id}", h.DeleteUser)).Methods("DELETE")
	router.HandleFunc("/health", h.metricsMiddleware("/health", h.HealthCheck)).Methods("GET")
}

