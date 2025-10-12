package main

// @title User Service API
// @version 1.0
// @description Microservice for user management with full observability stack (Prometheus, Jaeger, Grafana)
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://github.com/ozturkeniss/Full-Observability
// @contact.email support@example.com

// @license.name MIT
// @license.url https://github.com/ozturkeniss/Full-Observability/blob/main/LICENSE

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Auth
// @tag.description Authentication endpoints

// @tag.name Users
// @tag.description User management endpoints

// @tag.name Admin
// @tag.description Admin-only endpoints

// @tag.name Health
// @tag.description Health check endpoints

