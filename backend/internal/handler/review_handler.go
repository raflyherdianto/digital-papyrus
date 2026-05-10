package handler

import (
	"net/http"
	"strconv"

	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

type CreateReviewRequest struct {
	UserID    string            `json:"user_id" binding:"required"`
	OrderID   string            `json:"order_id" binding:"required"`
	ServiceID []string          `json:"service_id"`
	BookID    []string          `json:"book_id"`
	Details   map[string]string `json:"details"`
	Rating    map[string]int    `json:"rating"`
}

type UpdateReviewRequest struct {
	ServiceID []string          `json:"service_id"`
	BookID    []string          `json:"book_id"`
	Details   map[string]string `json:"details"`
	Rating    map[string]int    `json:"rating"`
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request body", "error": err.Error()})
		return
	}

	rev, err := h.svc.CreateReview(req.UserID, req.OrderID, req.ServiceID, req.BookID, req.Details, req.Rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create review", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Review created successfully", "data": rev})
}

func (h *ReviewHandler) GetReview(c *gin.Context) {
	id := c.Param("id")
	rev, err := h.svc.GetReview(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Review not found", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": rev})
}

func (h *ReviewHandler) GetAllReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	reviews, total, err := h.svc.GetAllReviews(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to retrieve reviews", "error": err.Error()})
		return
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reviews,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	id := c.Param("id")
	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request body", "error": err.Error()})
		return
	}

	rev, err := h.svc.UpdateReview(id, req.ServiceID, req.BookID, req.Details, req.Rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update review", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review updated successfully", "data": rev})
}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteReview(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete review", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review deleted successfully"})
}
