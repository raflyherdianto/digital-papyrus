package service

import (
	"fmt"
	"strings"

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
	slug, err := s.generateUniqueSlug(input.Name, "")
	if err != nil {
		return nil, err
	}

	category := &model.Category{
		ID:   uuid.New().String(),
		Name: input.Name,
		Slug: slug,
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
		if category.Name == "" {
			return nil, fmt.Errorf("category_service: name is required")
		}

		slug, err := s.generateUniqueSlug(category.Name, category.ID)
		if err != nil {
			return nil, err
		}
		category.Slug = slug
	}

	if err := s.repo.Update(category); err != nil {
		return nil, fmt.Errorf("category_service: update: %w", err)
	}
	return category, nil
}

func (s *CategoryService) generateUniqueSlug(name, excludeID string) (string, error) {
	base := slugifyCategoryName(name)
	slug := base
	idx := 2

	for {
		existing, err := s.repo.FindBySlug(slug)
		if err != nil {
			return "", fmt.Errorf("category_service: find slug: %w", err)
		}
		if existing == nil || existing.ID == excludeID {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, idx)
		idx++
	}
}

func slugifyCategoryName(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '-'
		}
	}, slug)

	// Collapse repeated separators and trim edges.
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "category"
	}
	return slug
}

// DeleteCategory removes a category by ID.
func (s *CategoryService) DeleteCategory(id string) error {
	return s.repo.Delete(id)
}
