package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
)

type OrderService struct {
	repo     *repository.OrderRepository
	userRepo *repository.UserRepository
}

func NewOrderService(repo *repository.OrderRepository, userRepo *repository.UserRepository) *OrderService {
	return &OrderService{repo: repo, userRepo: userRepo}
}

func (s *OrderService) GetAllOrders(filter repository.OrderFilter) ([]model.Order, int, error) {
	return s.repo.FindAll(filter)
}

func (s *OrderService) GetOrderByID(id string) (*model.Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("order id is required")
	}

	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	return order, nil
}

type CreateOrderInput struct {
	Invoice         string `json:"invoice"`
	UserID          string `json:"user_id"`
	Items           string `json:"items"`
	Notes           string `json:"notes"`
	TotalQty        int    `json:"total_qty"`
	TotalWeight     int    `json:"total_weight"`
	TotalPrice      int    `json:"total_price"`
	PaymentType     string `json:"payment_type"`
	Status          string `json:"status"`
	ShippingName    string `json:"shipping_name"`
	ShippingService string `json:"shipping_service"`
	ShippingPrice   int    `json:"shipping_price"`
}

func (s *OrderService) CreateOrder(input CreateOrderInput) (*model.Order, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return nil, errors.New("user_id is required")
	}

	user, err := s.userRepo.FindByID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("order_service: verify user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "pending"
	}

	order := &model.Order{
		ID:              uuid.New().String(),
		Invoice:         strings.TrimSpace(input.Invoice),
		UserID:          input.UserID,
		UserName:        user.Name,
		Notes:           strings.TrimSpace(input.Notes),
		TotalQty:        input.TotalQty,
		TotalWeight:     input.TotalWeight,
		TotalPrice:      input.TotalPrice,
		PaymentType:     strings.TrimSpace(input.PaymentType),
		Status:          status,
		ShippingName:    strings.TrimSpace(input.ShippingName),
		ShippingService: strings.TrimSpace(input.ShippingService),
		ShippingPrice:   input.ShippingPrice,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if order.Invoice == "" {
		order.Invoice = fmt.Sprintf("INV-%s", strings.ToUpper(order.ID[:8]))
	}

	items, err := parseOrderItems(input.Items)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(order, items); err != nil {
		return nil, err
	}

	return order, nil
}

func parseOrderItems(itemsText string) ([]repository.OrderItemInput, error) {
	text := strings.TrimSpace(itemsText)
	if text == "" {
		return nil, nil
	}

	rawParts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})

	items := make([]repository.OrderItemInput, 0, len(rawParts))
	for _, raw := range rawParts {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		title := line
		qty := 1
		if idx := strings.LastIndex(strings.ToLower(line), " x"); idx != -1 {
			title = strings.TrimSpace(line[:idx])
			qtyText := strings.TrimSpace(line[idx+2:])
			if qtyText != "" {
				parsedQty, err := strconv.Atoi(qtyText)
				if err != nil || parsedQty <= 0 {
					return nil, fmt.Errorf("invalid quantity in order item %q", line)
				}
				qty = parsedQty
			}
		}

		if title == "" {
			return nil, fmt.Errorf("invalid order item %q", line)
		}

		items = append(items, repository.OrderItemInput{Title: title, Qty: qty})
	}

	return items, nil
}

func (s *OrderService) DeleteOrder(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("order id is required")
	}
	return s.repo.Delete(id)
}
