package handler

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/middleware"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
)

type OrderHandler struct {
	svc         *service.OrderService
	cfg         *config.Config
	settingRepo *repository.SettingRepo
}

type CreateOrderRequest struct {
	Invoice         string `json:"invoice"`
	UserID          string `json:"user_id"`
	CustomerName     string `json:"customer_name"`
	CustomerEmail    string `json:"customer_email"`
	CustomerPhone    string `json:"customer_phone"`
	CustomerAddress  string `json:"customer_address"`
	CustomerProvince string `json:"customer_province"`
	CustomerCity     string `json:"customer_city"`
	CustomerDistrict string `json:"customer_district"`
	CustomerVillage  string `json:"customer_village"`
	CustomerZipCode  string `json:"customer_zip_code"`
	Items            string `json:"items"`
	Notes           string `json:"notes"`
	TotalQty        int    `json:"total_qty" binding:"required"`
	TotalWeight     int    `json:"total_weight"`
	TotalPrice      int    `json:"total_price" binding:"required"`
	PaymentType     string `json:"payment_type"`
	Status          string `json:"status"`
	ShippingName    string `json:"shipping_name"`
	ShippingService string `json:"shipping_service"`
	ShippingPrice   int    `json:"shipping_price"`
	Tax             int    `json:"tax"`
	ServiceFee      int    `json:"service_fee"`
	Discount        int    `json:"discount"`
}

func NewOrderHandler(svc *service.OrderService, cfg *config.Config, settingRepo *repository.SettingRepo) *OrderHandler {
	return &OrderHandler{
		svc:         svc,
		cfg:         cfg,
		settingRepo: settingRepo,
	}
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
		CustomerName:     req.CustomerName,
		CustomerEmail:    req.CustomerEmail,
		CustomerPhone:    req.CustomerPhone,
		CustomerAddress:  req.CustomerAddress,
		CustomerProvince: req.CustomerProvince,
		CustomerCity:     req.CustomerCity,
		CustomerDistrict: req.CustomerDistrict,
		CustomerVillage:  req.CustomerVillage,
		CustomerZipCode:  req.CustomerZipCode,
		Items:            req.Items,
		Notes:           req.Notes,
		TotalQty:        req.TotalQty,
		TotalWeight:     req.TotalWeight,
		TotalPrice:      req.TotalPrice,
		PaymentType:     req.PaymentType,
		Status:          req.Status,
		ShippingName:    req.ShippingName,
		ShippingService: req.ShippingService,
		ShippingPrice:   req.ShippingPrice,
		Tax:             req.Tax,
		ServiceFee:      req.ServiceFee,
		Discount:        req.Discount,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	// Otomatis kirim email invoice ke customer
	if fullOrder, err := h.svc.GetOrderByID(order.ID); err == nil && fullOrder != nil {
		_ = h.svc.SendInvoiceEmail(fullOrder, "Perlu Dibayar")
	} else {
		_ = h.svc.SendInvoiceEmail(order, "Perlu Dibayar")
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

func (h *OrderHandler) ListCustomerOrders(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	status := strings.ToLower(c.Query("status"))

	orders, total, err := h.svc.GetAllOrders(repository.OrderFilter{
		Page:    page,
		PerPage: perPage,
		UserID:  userIDStr,
		Status:  status,
	})
	if err != nil {
		response.InternalError(c, "Failed to retrieve your orders")
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

func (h *OrderHandler) GetCustomerOrder(c *gin.Context) {
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

	if order.UserID != userIDStr {
		response.Forbidden(c, "You do not have permission to access this order")
		return
	}

	response.OK(c, "Order retrieved successfully", order)
}

func (h *OrderHandler) CreateCustomerOrder(c *gin.Context) {
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

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	if req.UserID != userIDStr {
		response.Forbidden(c, "You cannot create an order for another user")
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
		Tax:             req.Tax,
		ServiceFee:      req.ServiceFee,
		Discount:        req.Discount,
	})
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	// Send invoice email with "Perlu Dibayar" status asynchronously
	if fullOrder, err := h.svc.GetOrderByID(order.ID); err == nil && fullOrder != nil {
		h.svc.SendInvoiceEmail(fullOrder, "Perlu Dibayar")
	} else {
		h.svc.SendInvoiceEmail(order, "Perlu Dibayar")
	}

	response.Created(c, "Order created successfully", order)
}

func (h *OrderHandler) CheckPayment(c *gin.Context) {
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
		response.NotFound(c, "Order not found")
		return
	}

	// Use GoBiz API scanner
	found, err := service.CheckGoBizPayment(order.TotalPrice, order.CreatedAt)
	if err != nil {
		response.InternalError(c, "Failed to check GoBiz payment: "+err.Error())
		return
	}

	if !found {
		response.OK(c, "Payment not found yet", map[string]interface{}{
			"status": "unpaid",
			"paid":   false,
		})
		return
	}

	err = h.svc.UpdateOrderStatus(order.ID, "confirmed")
	if err != nil {
		response.InternalError(c, "Payment found but failed to update order status")
		return
	}

	// Send invoice email asynchronously
	h.svc.SendInvoiceEmail(order, "Pembayaran Berhasil")

	response.OK(c, "Payment confirmed successfully", map[string]interface{}{
		"status": "confirmed",
		"paid":   true,
	})
}

func (h *OrderHandler) CancelCustomerOrder(c *gin.Context) {
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

	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Order ID is required", nil)
		return
	}

	order, err := h.svc.GetOrderByID(id)
	if err != nil {
		response.NotFound(c, "Order not found")
		return
	}

	if order.UserID != userIDStr {
		response.Forbidden(c, "You do not have permission to cancel this order")
		return
	}

	if strings.ToLower(order.Status) != "pending" {
		response.BadRequest(c, "Only pending orders can be cancelled", nil)
		return
	}

	err = h.svc.UpdateOrderStatus(order.ID, "cancelled")
	if err != nil {
		response.InternalError(c, "Failed to cancel order: "+err.Error())
		return
	}

	order.Status = "cancelled"
	h.svc.SendInvoiceEmail(order, "Dibatalkan")

	response.OK(c, "Order cancelled successfully", nil)
}

func (h *OrderHandler) ConfirmCustomerPayment(c *gin.Context) {
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

	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Order ID is required", nil)
		return
	}

	order, err := h.svc.GetOrderByID(id)
	if err != nil {
		response.NotFound(c, "Order not found")
		return
	}

	if order.UserID != userIDStr {
		response.Forbidden(c, "You do not have permission to confirm this order")
		return
	}

	currentStatus := strings.ToLower(order.Status)
	if currentStatus != "pending" && currentStatus != "unpaid" {
		response.BadRequest(c, "Only pending orders can be confirmed", nil)
		return
	}

	// Update status to "waiting_confirmation" in database
	if err := h.svc.UpdateOrderStatus(order.ID, "waiting_confirmation"); err != nil {
		response.InternalError(c, "Failed to update order status: "+err.Error())
		return
	}
	order.Status = "waiting_confirmation"

	// Dispatch "Menunggu Konfirmasi" email to customer
	_ = h.svc.SendInvoiceEmail(order, "Menunggu Konfirmasi")

	response.OK(c, "Payment confirmation submitted successfully. Status updated to waiting_confirmation.", order)
}

func (h *OrderHandler) AdminConfirmPayment(c *gin.Context) {
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
		response.NotFound(c, "Order not found")
		return
	}

	err = h.svc.UpdateOrderStatus(order.ID, "confirmed")
	if err != nil {
		response.InternalError(c, "Failed to update order status: "+err.Error())
		return
	}

	order.Status = "confirmed"
	h.svc.SendInvoiceEmail(order, "Lunas")

	response.OK(c, "Payment confirmed successfully", order)
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Order ID is required", nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		response.BadRequest(c, "Invalid order ID", nil)
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	order, err := h.svc.GetOrderByID(id)
	if err != nil {
		response.NotFound(c, "Order not found")
		return
	}

	err = h.svc.UpdateOrderStatus(order.ID, req.Status)
	if err != nil {
		response.InternalError(c, "Failed to update order status: "+err.Error())
		return
	}

	order.Status = req.Status
	switch strings.ToLower(req.Status) {
	case "confirmed", "paid", "lunas":
		_ = h.svc.SendInvoiceEmail(order, "Lunas")
	case "processed", "diproses":
		_ = h.svc.SendInvoiceEmail(order, "Diproses")
	case "shipped", "dikirim":
		_ = h.svc.SendInvoiceEmail(order, "Dikirim")
	case "delivered", "terkirim":
		_ = h.svc.SendInvoiceEmail(order, "Terkirim")
	case "completed", "selesai":
		_ = h.svc.SendInvoiceEmail(order, "Selesai")
	case "cancelled", "dibatalkan":
		_ = h.svc.SendInvoiceEmail(order, "Dibatalkan")
	}

	response.OK(c, "Order status updated successfully", order)
}

type ShippingCostRequest struct {
	DestinationVillageCode string  `json:"destination_village_code" binding:"required"`
	WeightKG               float64 `json:"weight_kg" binding:"required"`
}

func (h *OrderHandler) CalculateShippingCost(c *gin.Context) {
	var req ShippingCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	apiKey := h.cfg.API.CoIdKey
	if apiKey == "" {
		response.InternalError(c, "API_CO_ID_KEY is not configured")
		return
	}

	settingsMap, err := h.settingRepo.GetAllSettings()
	if err != nil {
		response.InternalError(c, "Failed to retrieve settings")
		return
	}

	originCode := "3575031009"
	if val, ok := settingsMap["origin_village_code"]; ok && val != "" {
		originCode = val
	}

	url := fmt.Sprintf("https://use.api.co.id/expedition/shipping-cost?origin_village_code=%s&destination_village_code=%s&weight=%v",
		originCode, req.DestinationVillageCode, req.WeightKG)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		response.InternalError(c, "Failed to create request: "+err.Error())
		return
	}

	httpReq.Header.Set("x-api-co-id", apiKey)
	httpReq.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		response.InternalError(c, "Failed to execute request: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.InternalError(c, "Failed to read response body: "+err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		response.InternalError(c, fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		response.InternalError(c, "Failed to parse API response: "+err.Error())
		return
	}

	var filteredData []interface{}
	dataMap, ok := result["data"].(map[string]interface{})
	if ok {
		couriers, ok := dataMap["couriers"].([]interface{})
		if ok {
			for _, courierRaw := range couriers {
				cMap, ok := courierRaw.(map[string]interface{})
				if !ok {
					continue
				}
				code, _ := cMap["courier_code"].(string)
				price, _ := cMap["price"].(float64)

				if code == "JNE" || code == "JT" {
					courierKey := "JNE"
					if code == "JT" {
						courierKey = "J&T"
					}
					filteredData = append(filteredData, map[string]interface{}{
						"courier": courierKey,
						"services": []interface{}{
							map[string]interface{}{
								"service": "REG",
								"price":   price,
							},
						},
					})
				}
			}
		}
	}

	var shippingPrice float64 = 0
	if len(filteredData) > 0 {
		firstCourier := filteredData[0].(map[string]interface{})
		services := firstCourier["services"].([]interface{})
		if len(services) > 0 {
			firstService := services[0].(map[string]interface{})
			if price, ok := firstService["price"].(float64); ok {
				shippingPrice = price
			}
		}
	}

	response.OK(c, "Shipping cost calculated", map[string]interface{}{
		"price":    shippingPrice,
		"couriers": filteredData,
	})
}

// GetPublicInvoice retrieves order/invoice details by ID or Invoice number for public sharing.
func (h *OrderHandler) GetPublicInvoice(c *gin.Context) {
	identifier := strings.TrimSpace(c.Param("id"))
	if identifier == "" {
		response.BadRequest(c, "Invoice identifier is required", nil)
		return
	}

	order, err := h.svc.GetOrderByIDOrInvoice(identifier)
	if err != nil {
		response.InternalError(c, "Failed to fetch invoice: "+err.Error())
		return
	}
	if order == nil {
		response.NotFound(c, "Invoice not found")
		return
	}

	response.OK(c, "Invoice fetched successfully", order)
}

