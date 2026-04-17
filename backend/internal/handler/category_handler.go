package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
)

// CategoryHandler handles category CRUD endpoints.
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// ListCategories handles GET /api/v1/categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.ListCategories()
	if err != nil {
		response.InternalError(c, "Failed to retrieve categories")
		return
	}

	response.OK(c, "Categories retrieved successfully", categories)
}

// GetCategory handles GET /api/v1/categories/:id
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Category ID is required", nil)
		return
	}

	category, err := h.categoryService.GetCategory(id)
	if err != nil {
		response.InternalError(c, "Failed to retrieve category")
		return
	}
	if category == nil {
		response.NotFound(c, "Category not found")
		return
	}

	response.OK(c, "Category retrieved successfully", category)
}

// CreateCategory handles POST /api/v1/categories
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var input service.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	if errs := input.Validate(); len(errs) > 0 {
		response.BadRequest(c, "Validation failed", errs)
		return
	}

	category, err := h.categoryService.CreateCategory(input)
	if err != nil {
		response.InternalError(c, "Failed to create category")
		return
	}

	response.Created(c, "Category created successfully", category)
}

// UpdateCategory handles PUT /api/v1/categories/:id
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Category ID is required", nil)
		return
	}

	var input service.UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	category, err := h.categoryService.UpdateCategory(id, input)
	if err != nil {
		response.InternalError(c, "Failed to update category")
		return
	}
	if category == nil {
		response.NotFound(c, "Category not found")
		return
	}

	response.OK(c, "Category updated successfully", category)
}

// DeleteCategory handles DELETE /api/v1/categories/:id
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Category ID is required", nil)
		return
	}

	if err := h.categoryService.DeleteCategory(id); err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	response.OK(c, "Category deleted successfully", nil)
}