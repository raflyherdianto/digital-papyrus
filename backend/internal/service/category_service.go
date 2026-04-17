package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/pkg/validator"
)

// CategoryService handles category business logic.
type CategoryService struct {
	repo *repository.CategoryRepository
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// ListCategories retrieves all categories.
func (s *CategoryService) ListCategories() ([]model.Category, error) {
	return s.repo.FindAll()
}

// GetCategory retrieves a category by ID.
func (s *CategoryService) GetCategory(id string) (*model.Category, error) {
	category, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("category_service: %w", err)
	}
	return category, nil
}

type CreateCategoryInput struct {
	Name string `json:"name"`
}

func (i *CreateCategoryInput) Validate() map[string]string {
	errs := make(map[string]string)
	i.Name = validator.SanitizeString(i.Name)
	if i.Name == "" {
		errs["name"] = "name is required"
	}
	return errs
}

// CreateCategory creates a new category.
func (s *CategoryService) CreateCategory(input CreateCategoryInput) (*model.Category, error) {
	category := &model.Category{
		ID:   uuid.New().String(),
		Name: input.Name,
	}

	if err := s.repo.Create(category); err != nil {
		return nil, fmt.Errorf("category_service: create: %w", err)
	}
	return category, nil
}

type UpdateCategoryInput struct {
	Name *string `json:"name"`
}

// UpdateCategory applies partial updates to an existing category.
func (s *CategoryService) UpdateCategory(id string, input UpdateCategoryInput) (*model.Category, error) {
	category, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("category_service: %w", err)
	}
	if category == nil {
		return nil, nil
	}

	if input.Name != nil {
		category.Name = validator.SanitizeString(*input.Name)
	}

	if err := s.repo.Update(category); err != nil {
		return nil, fmt.Errorf("category_service: update: %w", err)
	}
	return category, nil
}

// DeleteCategory removes a category by ID.
func (s *CategoryService) DeleteCategory(id string) error {
	return s.repo.Delete(id)
}