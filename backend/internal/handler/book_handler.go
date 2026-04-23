package handler

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

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

	filter := repository.BookFilter{
		Status:     c.Query("status"),
		CategoryID: c.Query("category_id"),
		Search:     c.Query("search"),
		Page:       page,
		PerPage:    perPage,
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
		if strings.Contains(err.Error(), "UNIQUE") {
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
		if strings.Contains(err.Error(), "UNIQUE") {
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
