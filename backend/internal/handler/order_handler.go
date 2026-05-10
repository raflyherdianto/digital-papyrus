package handler

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
)

type OrderHandler struct {
	svc *service.OrderService
}

type CreateOrderRequest struct {
	Invoice         string `json:"invoice"`
	UserID          string `json:"user_id" binding:"required"`
	Items           string `json:"items"`
	Notes           string `json:"notes"`
	TotalQty        int    `json:"total_qty" binding:"required"`
	TotalWeight     int    `json:"total_weight"`
	TotalPrice      int    `json:"total_price" binding:"required"`
	PaymentType     string `json:"payment_type"`
	Status          string `json:"status" binding:"required"`
	ShippingName    string `json:"shipping_name"`
	ShippingService string `json:"shipping_service"`
	ShippingPrice   int    `json:"shipping_price"`
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	orders, total, err := h.svc.GetAllOrders(repository.OrderFilter{
		Page:    page,
		PerPage: perPage,
		Search:  c.Query("search"),
		Status:  strings.ToLower(c.Query("status")),
	})
	if err != nil {
		response.InternalError(c, "Failed to retrieve orders")
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	response.OKWithMeta(c, "Orders retrieved successfully", orders, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      int64(total),
		TotalPages: totalPages,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Order ID is required", nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		response.BadRequest(c, "Invalid order ID", nil)
		return
	}

	order, err := h.svc.GetOrderByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Order not found")
			return
		}
		response.InternalError(c, "Failed to retrieve order")
		return
	}

	response.OK(c, "Order retrieved successfully", order)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	order, err := h.svc.CreateOrder(service.CreateOrderInput{
		Invoice:         req.Invoice,
		UserID:          req.UserID,
		Items:           req.Items,
		Notes:           req.Notes,
		TotalQty:        req.TotalQty,
		TotalWeight:     req.TotalWeight,
		TotalPrice:      req.TotalPrice,
		PaymentType:     req.PaymentType,
		Status:          req.Status,
		ShippingName:    req.ShippingName,
		ShippingService: req.ShippingService,
		ShippingPrice:   req.ShippingPrice,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Created(c, "Order created successfully", order)
}

func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Order ID is required", nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		response.BadRequest(c, "Invalid order ID", nil)
		return
	}

	if err := h.svc.DeleteOrder(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Order not found")
			return
		}
		response.InternalError(c, "Failed to delete order")
		return
	}

	response.OK(c, "Order deleted successfully", nil)
}
