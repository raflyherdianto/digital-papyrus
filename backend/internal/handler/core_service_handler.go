package handler

import (
	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type CoreServiceHandler struct {
	svc *service.CoreServiceService
}

func NewCoreServiceHandler(svc *service.CoreServiceService) *CoreServiceHandler {
	return &CoreServiceHandler{svc: svc}
}

func (h *CoreServiceHandler) GetAll(c *gin.Context) {
	coreServices, err := h.svc.FindAll()
	if err != nil {
		response.InternalError(c, "Failed to retrieve core services")
		return
	}
	response.OK(c, "Core services retrieved successfully", coreServices)
}

func (h *CoreServiceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	s, err := h.svc.FindByID(id)
	if err != nil {
		response.InternalError(c, "Failed to retrieve core service")
		return
	}
	if s == nil {
		response.NotFound(c, "Core service not found")
		return
	}
	response.OK(c, "Core service retrieved successfully", s)
}

func (h *CoreServiceHandler) Create(c *gin.Context) {
	var s model.CoreService
	if err := c.ShouldBindJSON(&s); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	if err := h.svc.Create(&s); err != nil {
		response.InternalError(c, "Failed to create core service")
		return
	}
	response.Created(c, "Core service created successfully", s)
}

func (h *CoreServiceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	
	existing, err := h.svc.FindByID(id)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if existing == nil {
		response.NotFound(c, "Core service not found")
		return
	}

	var s model.CoreService
	if err := c.ShouldBindJSON(&s); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	s.ID = id
	if err := h.svc.Update(&s); err != nil {
		response.InternalError(c, "Failed to update core service")
		return
	}
	response.OK(c, "Core service updated successfully", s)
}

func (h *CoreServiceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(id); err != nil {
		response.InternalError(c, "Failed to delete core service")
		return
	}
	response.OK(c, "Core service deleted successfully", nil)
}
