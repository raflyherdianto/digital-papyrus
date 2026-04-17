package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/pkg/validator"
)

// ServiceService handles service (publishing package) business logic.
type ServiceService struct {
	serviceRepo *repository.ServiceRepository
}

// NewServiceService creates a new ServiceService.
func NewServiceService(serviceRepo *repository.ServiceRepository) *ServiceService {
	return &ServiceService{serviceRepo: serviceRepo}
}

// CreateServiceInput represents the request body for creating a service.
type CreateServiceInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Tier        string `json:"tier"`
	Price       int    `json:"price"`
	PriceLabel  string `json:"price_label"`
	Features    string `json:"features"` // JSON array string
	IsFeatured  bool   `json:"is_featured"`
	Badge       string `json:"badge"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

// Validate checks all required fields and business rules.
func (i *CreateServiceInput) Validate() map[string]string {
	errs := make(map[string]string)
	i.Title = validator.SanitizeString(i.Title)

	if i.Title == "" {
		errs["title"] = "title is required"
	}
	if i.Tier == "" {
		errs["tier"] = "tier is required"
	}
	validTiers := map[string]bool{
		model.ServiceTierBasic: true, model.ServiceTierSilver: true,
		model.ServiceTierGold: true, model.ServiceTierPlatinum: true,
	}
	if !validTiers[i.Tier] {
		errs["tier"] = "tier must be basic, silver, gold, or platinum"
	}
	if i.Price < 0 {
		errs["price"] = "price must be non-negative"
	}
	return errs
}

// ListServices retrieves all services.
func (s *ServiceService) ListServices(activeOnly bool) ([]model.Service, error) {
	return s.serviceRepo.FindAll(activeOnly)
}

// GetService retrieves a single service by ID.
func (s *ServiceService) GetService(id string) (*model.Service, error) {
	svc, err := s.serviceRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("service_service: %w", err)
	}
	return svc, nil
}

// CreateService creates a new service.
func (s *ServiceService) CreateService(input CreateServiceInput) (*model.Service, error) {
	svc := &model.Service{
		ID:          uuid.New().String(),
		Title:       input.Title,
		Description: input.Description,
		Icon:        input.Icon,
		Tier:        input.Tier,
		Price:       input.Price,
		PriceLabel:  input.PriceLabel,
		Features:    input.Features,
		IsFeatured:  input.IsFeatured,
		Badge:       input.Badge,
		SortOrder:   input.SortOrder,
		IsActive:    input.IsActive,
	}

	if err := s.serviceRepo.Create(svc); err != nil {
		return nil, fmt.Errorf("service_service: create: %w", err)
	}
	return svc, nil
}

// UpdateServiceInput represents the request body for updating a service.
type UpdateServiceInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	Tier        *string `json:"tier"`
	Price       *int    `json:"price"`
	PriceLabel  *string `json:"price_label"`
	Features    *string `json:"features"`
	IsFeatured  *bool   `json:"is_featured"`
	Badge       *string `json:"badge"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

// UpdateService applies partial updates to an existing service.
func (s *ServiceService) UpdateService(id string, input UpdateServiceInput) (*model.Service, error) {
	svc, err := s.serviceRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("service_service: %w", err)
	}
	if svc == nil {
		return nil, nil
	}

	if input.Title != nil {
		svc.Title = validator.SanitizeString(*input.Title)
	}
	if input.Description != nil {
		svc.Description = *input.Description
	}
	if input.Icon != nil {
		svc.Icon = *input.Icon
	}
	if input.Tier != nil {
		svc.Tier = *input.Tier
	}
	if input.Price != nil {
		svc.Price = *input.Price
	}
	if input.PriceLabel != nil {
		svc.PriceLabel = *input.PriceLabel
	}
	if input.Features != nil {
		svc.Features = *input.Features
	}
	if input.IsFeatured != nil {
		svc.IsFeatured = *input.IsFeatured
	}
	if input.Badge != nil {
		svc.Badge = *input.Badge
	}
	if input.SortOrder != nil {
		svc.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		svc.IsActive = *input.IsActive
	}

	if err := s.serviceRepo.Update(svc); err != nil {
		return nil, fmt.Errorf("service_service: update: %w", err)
	}
	return svc, nil
}

// DeleteService removes a service by ID.
func (s *ServiceService) DeleteService(id string) error {
	return s.serviceRepo.Delete(id)
}
