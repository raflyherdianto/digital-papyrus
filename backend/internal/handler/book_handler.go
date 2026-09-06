package handler

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/middleware"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
)

// BookHandler handles book CRUD endpoints.
type BookHandler struct {
	bookService *service.BookService
}

// NewBookHandler creates a new BookHandler.
func NewBookHandler(bookService *service.BookService) *BookHandler {
	return &BookHandler{bookService: bookService}
}

// ListBooks handles GET /api/v1/books
func (h *BookHandler) ListBooks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "12"))
	minPrice, _ := strconv.Atoi(c.Query("min_price"))
	maxPrice, _ := strconv.Atoi(c.Query("max_price"))
	maxRating, _ := strconv.ParseFloat(c.Query("max_rating"), 64)

	filter := repository.BookFilter{
		Status:     c.Query("status"),
		CategoryID: c.Query("category_id"),
		Badge:      c.Query("badge"),
		Search:     c.Query("search"),
		Page:       page,
		PerPage:    perPage,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		MaxRating:  maxRating,
		Sort:       strings.ToLower(c.Query("sort")),
	}

	books, total, err := h.bookService.ListBooks(filter)
	if err != nil {
		response.InternalError(c, "Failed to retrieve books")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	response.OKWithMeta(c, "Books retrieved successfully", books, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetBook handles GET /api/v1/books/:id
func (h *BookHandler) GetBook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Book ID is required", nil)
		return
	}

	book, err := h.bookService.GetBook(id)
	if err != nil {
		response.InternalError(c, "Failed to retrieve book")
		return
	}
	if book == nil {
		response.NotFound(c, "Book not found")
		return
	}

	response.OK(c, "Book retrieved successfully", book)
}

// CreateBook handles POST /api/v1/books
func (h *BookHandler) CreateBook(c *gin.Context) {
	var input service.CreateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	if errs := input.Validate(); len(errs) > 0 {
		response.BadRequest(c, "Validation failed", errs)
		return
	}

	book, err := h.bookService.CreateBook(input)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			response.BadRequest(c, "ISBN sudah digunakan oleh buku lain", map[string]string{"isbn": "ISBN ini sudah terdaftar, gunakan ISBN yang berbeda"})
			return
		}
		response.InternalError(c, "Failed to create book")
		return
	}

	response.Created(c, "Book created successfully", book)
}

// UpdateBook handles PUT /api/v1/books/:id
func (h *BookHandler) UpdateBook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Book ID is required", nil)
		return
	}

	var input service.UpdateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	book, err := h.bookService.UpdateBook(id, input)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			response.BadRequest(c, "ISBN sudah digunakan oleh buku lain", map[string]string{"isbn": "ISBN ini sudah terdaftar, gunakan ISBN yang berbeda"})
			return
		}
		response.InternalError(c, "Failed to update book")
		return
	}
	if book == nil {
		response.NotFound(c, "Book not found")
		return
	}

	response.OK(c, "Book updated successfully", book)
}

// DeleteBook handles DELETE /api/v1/books/:id
func (h *BookHandler) DeleteBook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Book ID is required", nil)
		return
	}

	if err := h.bookService.DeleteBook(id); err != nil {
		response.NotFound(c, "Book not found")
		return
	}

	response.OK(c, "Book deleted successfully", nil)
}

// CreateCustomerBook handles POST /api/v1/customer/books
func (h *BookHandler) CreateCustomerBook(c *gin.Context) {
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		response.InternalError(c, "Invalid user ID type")
		return
	}

	var input service.CreateCustomerBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	input.UserID = userIDStr

	if errs := input.Validate(); len(errs) > 0 {
		response.BadRequest(c, "Validation failed", errs)
		return
	}

	book, err := h.bookService.CreateCustomerBook(input)
	if err != nil {
		response.InternalError(c, "Failed to submit publishing request: "+err.Error())
		return
	}

	response.Created(c, "Publishing request submitted successfully", book)
}

// UpdateCustomerBookInput represents payload for customer updating their draft.
type UpdateCustomerBookInput struct {
	Title       string `json:"title" binding:"required"`
	Author      string `json:"author" binding:"required"`
	CategoryID  string `json:"category_id" binding:"required"`
	Description string `json:"description"`
	Format      string `json:"format" binding:"required"`
	Language    string `json:"language" binding:"required"`
	DraftURL    string `json:"draft_url" binding:"required"`
}

// UpdateCustomerBook handles PUT /api/v1/customer/books/:id
func (h *BookHandler) UpdateCustomerBook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Book ID is required", nil)
		return
	}

	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		response.InternalError(c, "Invalid user ID type")
		return
	}

	var input UpdateCustomerBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	// Update the book service
	book, err := h.bookService.UpdateCustomerBookDraft(id, userIDStr, service.UpdateCustomerBookInput{
		Title:       input.Title,
		Author:      input.Author,
		CategoryID:  input.CategoryID,
		Description: input.Description,
		Format:      input.Format,
		Language:    input.Language,
		DraftURL:    input.DraftURL,
	})
	if err != nil {
		response.InternalError(c, "Failed to update draft: "+err.Error())
		return
	}
	if book == nil {
		response.NotFound(c, "Book not found or unauthorized")
		return
	}

	response.OK(c, "Draft updated successfully", book)
}

// ValidateBookInput represents payload for validating a book.
type ValidateBookInput struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}

// ValidateBook handles POST /api/v1/books/:id/validate
func (h *BookHandler) ValidateBook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Book ID is required", nil)
		return
	}

	var input ValidateBookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	status := strings.ToLower(input.Status)
	if status != "approved" && status != "rejected" {
		response.BadRequest(c, "Status must be 'approved' or 'rejected'", nil)
		return
	}

	book, err := h.bookService.ValidateBook(id, status, input.Notes)
	if err != nil {
		response.InternalError(c, "Failed to validate book: "+err.Error())
		return
	}
	if book == nil {
		response.NotFound(c, "Book not found")
		return
	}

	response.OK(c, "Book validation updated successfully", book)
}
