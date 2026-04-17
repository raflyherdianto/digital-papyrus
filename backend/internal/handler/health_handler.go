// Package handler provides HTTP request handlers for all API endpoints.
package handler

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/pkg/response"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{startTime: time.Now()}
}

// HealthCheck returns the server health status.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response.OK(c, "Service is healthy", gin.H{
		"status":     "healthy",
		"service":    "digital-papyrus-api",
		"version":    "1.0.0",
		"go_version": runtime.Version(),
		"uptime":     time.Since(h.startTime).String(),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}
