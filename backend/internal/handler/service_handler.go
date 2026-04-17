package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
)

// ServiceHandler handles service CRUD endpoints.
type ServiceHandler struct {
	serviceService *service.ServiceService
}

// NewServiceHandler creates a new ServiceHandler.
func NewServiceHandler(serviceService *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{serviceService: serviceService}
}

// ListServices handles GET /api/v1/services
func (h *ServiceHandler) ListServices(c *gin.Context) {
	// Public route shows only active services
	activeOnly := c.DefaultQuery("active_only", "true") == "true"

	services, err := h.serviceService.ListServices(activeOnly)
	if err != nil {
		response.InternalError(c, "Failed to retrieve services")
		return
	}

	response.OK(c, "Services retrieved successfully", services)
}

// GetService handles GET /api/v1/services/:id
func (h *ServiceHandler) GetService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Service ID is required", nil)
		return
	}

	svc, err := h.serviceService.GetService(id)
	if err != nil {
		response.InternalError(c, "Failed to retrieve service")
		return
	}
	if svc == nil {
		response.NotFound(c, "Service not found")
		return
	}

	response.OK(c, "Service retrieved successfully", svc)
}

// CreateService handles POST /api/v1/services
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var input service.CreateServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	if errs := input.Validate(); len(errs) > 0 {
		response.BadRequest(c, "Validation failed", errs)
		return
	}

	svc, err := h.serviceService.CreateService(input)
	if err != nil {
		response.InternalError(c, "Failed to create service")
		return
	}

	response.Created(c, "Service created successfully", svc)
}

// UpdateService handles PUT /api/v1/services/:id
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Service ID is required", nil)
		return
	}

	var input service.UpdateServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	svc, err := h.serviceService.UpdateService(id, input)
	if err != nil {
		response.InternalError(c, "Failed to update service")
		return
	}
	if svc == nil {
		response.NotFound(c, "Service not found")
		return
	}

	response.OK(c, "Service updated successfully", svc)
}

// DeleteService handles DELETE /api/v1/services/:id
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Service ID is required", nil)
		return
	}

	if err := h.serviceService.DeleteService(id); err != nil {
		response.NotFound(c, "Service not found")
		return
	}

	response.OK(c, "Service deleted successfully", nil)
}
