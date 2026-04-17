// Package response provides standardized API response structures
// following industry-standard JSON:API-inspired patterns.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// Meta provides pagination and summary metadata.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// ErrorInfo provides structured error details.
type ErrorInfo struct {
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// OK sends a 200 success response.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// OKWithMeta sends a 200 success response with pagination metadata.
func OKWithMeta(c *gin.Context, message string, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 response for successful resource creation.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// BadRequest sends a 400 response for client errors.
func BadRequest(c *gin.Context, message string, details map[string]string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "BAD_REQUEST",
			Details: details,
		},
	})
}

// Unauthorized sends a 401 response for authentication failures.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code: "UNAUTHORIZED",
		},
	})
}

// Forbidden sends a 403 response for authorization failures.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code: "FORBIDDEN",
		},
	})
}

// NotFound sends a 404 response.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code: "NOT_FOUND",
		},
	})
}

// Conflict sends a 409 response for duplicate resource conflicts.
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code: "CONFLICT",
		},
	})
}

// InternalError sends a 500 response for server-side errors.
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code: "INTERNAL_ERROR",
		},
	})
}

// TooManyRequests sends a 429 response for rate limiting.
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code: "RATE_LIMIT_EXCEEDED",
		},
	})
}
