package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
)

type OrderService struct {
	repo     *repository.OrderRepository
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewOrderService(repo *repository.OrderRepository, userRepo *repository.UserRepository, cfg *config.Config) *OrderService {
	svc := &OrderService{repo: repo, userRepo: userRepo, cfg: cfg}

	// Background worker to check and cancel expired orders (exceeding 24 hours without confirmation)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			_ = svc.CheckAndCancelExpiredOrders()
		}
	}()

	return svc
}

func (s *OrderService) CheckAndCancelExpiredOrders() error {
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	expiredOrders, err := s.repo.FindExpiredOrders(cutoff)
	if err != nil {
		return err
	}

	for _, order := range expiredOrders {
		if err := s.repo.UpdateStatus(order.ID, "cancelled"); err == nil {
			if fullOrder, err := s.repo.FindByID(order.ID); err == nil && fullOrder != nil {
				s.SendInvoiceEmail(fullOrder, "Dibatalkan")
			}
		}
	}
	return nil
}

func (s *OrderService) GetAllOrders(filter repository.OrderFilter) ([]model.Order, int, error) {
	_ = s.CheckAndCancelExpiredOrders()
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

	// If pending or waiting_confirmation and exceeded 24 hours, auto-cancel
	st := strings.ToLower(order.Status)
	if (st == "pending" || st == "waiting_confirmation") && time.Since(order.CreatedAt) > 24*time.Hour {
		if err := s.repo.UpdateStatus(order.ID, "cancelled"); err == nil {
			order.Status = "cancelled"
			s.SendInvoiceEmail(order, "Dibatalkan")
		}
	}

	return order, nil
}

type CreateOrderInput struct {
	Invoice          string `json:"invoice"`
	UserID           string `json:"user_id"`
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
	Notes            string `json:"notes"`
	TotalQty         int    `json:"total_qty"`
	TotalWeight      int    `json:"total_weight"`
	TotalPrice       int    `json:"total_price"`
	PaymentType      string `json:"payment_type"`
	Status           string `json:"status"`
	ShippingName     string `json:"shipping_name"`
	ShippingService  string `json:"shipping_service"`
	ShippingPrice    int    `json:"shipping_price"`
	Tax              int    `json:"tax"`
	ServiceFee       int    `json:"service_fee"`
	Discount         int    `json:"discount"`
}

func (s *OrderService) CreateOrder(input CreateOrderInput) (*model.Order, error) {
	var user *model.User
	var err error

	if strings.TrimSpace(input.UserID) != "" {
		user, err = s.userRepo.FindByID(input.UserID)
		if err != nil {
			return nil, fmt.Errorf("order_service: verify user: %w", err)
		}
		if user == nil {
			return nil, errors.New("user not found")
		}
	} else if strings.TrimSpace(input.CustomerName) != "" {
		custName := strings.TrimSpace(input.CustomerName)
		custEmail := strings.TrimSpace(input.CustomerEmail)
		if custEmail == "" {
			cleanName := strings.ToLower(strings.ReplaceAll(custName, " ", "."))
			cleanName = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' {
					return r
				}
				return -1
			}, cleanName)
			if cleanName == "" {
				cleanName = "customer"
			}
			custEmail = fmt.Sprintf("%s_%s@guest.digitalpapyrus.web.id", cleanName, uuid.New().String()[:6])
		}
		user, _ = s.userRepo.FindByEmail(custEmail)
		if user == nil {
			hash, err := bcrypt.GenerateFromPassword([]byte("Customer123!"), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("order_service: hash password: %w", err)
			}
			user = &model.User{
				ID:           uuid.New().String(),
				Email:        custEmail,
				PasswordHash: string(hash),
				Name:         custName,
				Role:         "customer",
				IsActive:     true,
				PhoneNumber:  strings.TrimSpace(input.CustomerPhone),
				Address:      strings.TrimSpace(input.CustomerAddress),
				Province:     strings.TrimSpace(input.CustomerProvince),
				City:         strings.TrimSpace(input.CustomerCity),
				Regency:      strings.TrimSpace(input.CustomerDistrict),
				District:     strings.TrimSpace(input.CustomerDistrict),
				Village:      strings.TrimSpace(input.CustomerVillage),
				ZipCode:      strings.TrimSpace(input.CustomerZipCode),
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
			}
			if err := s.userRepo.Create(user); err != nil {
				return nil, fmt.Errorf("order_service: create manual customer: %w", err)
			}
		} else {
			if strings.TrimSpace(input.CustomerAddress) != "" {
				user.Address = strings.TrimSpace(input.CustomerAddress)
			}
			if strings.TrimSpace(input.CustomerProvince) != "" {
				user.Province = strings.TrimSpace(input.CustomerProvince)
			}
			if strings.TrimSpace(input.CustomerCity) != "" {
				user.City = strings.TrimSpace(input.CustomerCity)
			}
			if strings.TrimSpace(input.CustomerDistrict) != "" {
				user.Regency = strings.TrimSpace(input.CustomerDistrict)
				user.District = strings.TrimSpace(input.CustomerDistrict)
			}
			if strings.TrimSpace(input.CustomerVillage) != "" {
				user.Village = strings.TrimSpace(input.CustomerVillage)
			}
			if strings.TrimSpace(input.CustomerZipCode) != "" {
				user.ZipCode = strings.TrimSpace(input.CustomerZipCode)
			}
			if strings.TrimSpace(input.CustomerPhone) != "" {
				user.PhoneNumber = strings.TrimSpace(input.CustomerPhone)
			}
			_ = s.userRepo.Update(user)
		}
		input.UserID = user.ID
	} else {
		return nil, errors.New("customer (pilih user_id atau isi nama customer) wajib diisi")
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
		UserEmail:       user.Email,
		UserPhone:       user.PhoneNumber,
		UserAddress:     user.Address,
		Notes:           strings.TrimSpace(input.Notes),
		TotalQty:        input.TotalQty,
		TotalWeight:     input.TotalWeight,
		TotalPrice:      input.TotalPrice,
		PaymentType:     strings.TrimSpace(input.PaymentType),
		Status:          status,
		ShippingName:    strings.TrimSpace(input.ShippingName),
		ShippingService: strings.TrimSpace(input.ShippingService),
		ShippingPrice:   input.ShippingPrice,
		Tax:             input.Tax,
		ServiceFee:      input.ServiceFee,
		Discount:        input.Discount,
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

	if fullOrder, err := s.repo.FindByID(order.ID); err == nil && fullOrder != nil {
		return fullOrder, nil
	}

	return order, nil
}

func parseOrderItems(itemsText string) ([]repository.OrderItemInput, error) {
	text := strings.TrimSpace(itemsText)
	if text == "" {
		return nil, nil
	}

	var jsonItems []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Title string `json:"title"`
		Qty   int    `json:"qty"`
		Price int    `json:"price"`
	}
	if err := json.Unmarshal([]byte(text), &jsonItems); err == nil {
		var result []repository.OrderItemInput
		for _, item := range jsonItems {
			result = append(result, repository.OrderItemInput{
				ID:    item.ID,
				Type:  item.Type,
				Title: item.Title,
				Qty:   item.Qty,
				Price: item.Price,
			})
		}
		return result, nil
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

func (s *OrderService) UpdateOrderStatus(id string, status string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("order id is required")
	}
	if strings.TrimSpace(status) == "" {
		return errors.New("status is required")
	}
	return s.repo.UpdateStatus(id, status)
}

func (s *OrderService) GetOrderByIDOrInvoice(idOrInvoice string) (*model.Order, error) {
	if strings.TrimSpace(idOrInvoice) == "" {
		return nil, errors.New("order id or invoice is required")
	}
	return s.repo.FindByIDOrInvoice(idOrInvoice)
}

