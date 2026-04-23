package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/pkg/validator"
)

// BookService handles book business logic.
type BookService struct {
	bookRepo *repository.BookRepository
}

// NewBookService creates a new BookService.
func NewBookService(bookRepo *repository.BookRepository) *BookService {
	return &BookService{bookRepo: bookRepo}
}

// CreateBookInput represents the request body for creating a book.
type CreateBookInput struct {
	Title           string  `json:"title"`
	Author          string  `json:"author"`
	ISBN            string  `json:"isbn"`
	Price           int     `json:"price"`
	Rating          float64 `json:"rating"`
	Description     string  `json:"description"`
	Synopsis        string  `json:"synopsis"`
	ImageURL        string  `json:"image_url"`
	CategoryID      string  `json:"category_id"`
	Status          string  `json:"status"`
	Stock           int     `json:"stock"`
	Publisher       string  `json:"publisher"`
	PublicationDate string  `json:"publication_date"`
	Pages           int     `json:"pages"`
	Format          string  `json:"format"`
	Language        string  `json:"language"`
	Dimensions      string  `json:"dimensions"`
	Weight          string  `json:"weight"`
}

// Validate checks all required fields and business rules.
func (i *CreateBookInput) Validate() map[string]string {
	errs := make(map[string]string)
	i.Title = validator.SanitizeString(i.Title)
	i.Author = validator.SanitizeString(i.Author)

	if i.Title == "" {
		errs["title"] = "title is required"
	}
	if i.Author == "" {
		errs["author"] = "author is required"
	}
	if i.Price < 0 {
		errs["price"] = "price must be non-negative"
	}
	if i.Status == "" {
		i.Status = model.BookStatusDraft
	}
	if i.Status != model.BookStatusDraft && i.Status != model.BookStatusPublished && i.Status != model.BookStatusArchived {
		errs["status"] = "status must be draft, published, or archived"
	}
	if i.Rating < 0 || i.Rating > 5 {
		errs["rating"] = "rating must be between 0 and 5"
	}
	if i.Stock < 0 {
		errs["stock"] = "stock must be non-negative"
	}
	return errs
}

// ListBooks retrieves a paginated list of books with optional filters.
func (s *BookService) ListBooks(f repository.BookFilter) ([]model.Book, int64, error) {
	return s.bookRepo.FindAll(f)
}

// GetBook retrieves a single book by ID.
func (s *BookService) GetBook(id string) (*model.Book, error) {
	book, err := s.bookRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("book_service: %w", err)
	}
	return book, nil
}

// CreateBook creates a new book.
func (s *BookService) CreateBook(input CreateBookInput) (*model.Book, error) {
	book := &model.Book{
		ID:              uuid.New().String(),
		Title:           input.Title,
		Author:          input.Author,
		ISBN:            input.ISBN,
		Price:           input.Price,
		Rating:          input.Rating,
		Description:     input.Description,
		Synopsis:        input.Synopsis,
		ImageURL:        input.ImageURL,
		CategoryID:     input.CategoryID,
		Status:          input.Status,
		Stock:           input.Stock,
		Publisher:       input.Publisher,
		PublicationDate: input.PublicationDate,
		Pages:           input.Pages,
		Format:          input.Format,
		Language:        input.Language,
		Dimensions:      input.Dimensions,
		Weight:          input.Weight,
	}

	if err := s.bookRepo.Create(book); err != nil {
		return nil, fmt.Errorf("book_service: create: %w", err)
	}
	return book, nil
}

// UpdateBookInput represents the request body for updating a book.
type UpdateBookInput struct {
	Title           *string  `json:"title"`
	Author          *string  `json:"author"`
	ISBN            *string  `json:"isbn"`
	Price           *int     `json:"price"`
	Rating          *float64 `json:"rating"`
	ReviewCount     *int     `json:"review_count"`
	Description     *string  `json:"description"`
	Synopsis        *string  `json:"synopsis"`
	ImageURL        *string  `json:"image_url"`
	CategoryID      *string  `json:"category_id"`
	Status          *string  `json:"status"`
	Stock           *int     `json:"stock"`
	Publisher       *string  `json:"publisher"`
	PublicationDate *string  `json:"publication_date"`
	Pages           *int     `json:"pages"`
	Format          *string  `json:"format"`
	Language        *string  `json:"language"`
	Dimensions      *string  `json:"dimensions"`
	Weight          *string  `json:"weight"`
}

// UpdateBook applies partial updates to an existing book.
func (s *BookService) UpdateBook(id string, input UpdateBookInput) (*model.Book, error) {
	book, err := s.bookRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("book_service: %w", err)
	}
	if book == nil {
		return nil, nil
	}

	oldImageURL := book.ImageURL

	// Apply partial updates (PATCH semantics)
	if input.Title != nil {
		book.Title = validator.SanitizeString(*input.Title)
	}
	if input.Author != nil {
		book.Author = validator.SanitizeString(*input.Author)
	}
	if input.ISBN != nil {
		book.ISBN = *input.ISBN
	}
	if input.Price != nil {
		book.Price = *input.Price
	}
	if input.Rating != nil {
		book.Rating = *input.Rating
	}
	if input.ReviewCount != nil {
		book.ReviewCount = *input.ReviewCount
	}
	if input.Description != nil {
		book.Description = *input.Description
	}
	if input.Synopsis != nil {
		book.Synopsis = *input.Synopsis
	}
	if input.ImageURL != nil {
		book.ImageURL = *input.ImageURL
	}
	if input.CategoryID != nil {
		book.CategoryID = *input.CategoryID
	}
	if input.Status != nil {
		book.Status = *input.Status
	}
	if input.Stock != nil {
		book.Stock = *input.Stock
	}
	if input.Publisher != nil {
		book.Publisher = *input.Publisher
	}
	if input.PublicationDate != nil {
		book.PublicationDate = *input.PublicationDate
	}
	if input.Pages != nil {
		book.Pages = *input.Pages
	}
	if input.Format != nil {
		book.Format = *input.Format
	}
	if input.Language != nil {
		book.Language = *input.Language
	}
	if input.Dimensions != nil {
		book.Dimensions = *input.Dimensions
	}
	if input.Weight != nil {
		book.Weight = *input.Weight
	}

	if err := s.bookRepo.Update(book); err != nil {
		return nil, fmt.Errorf("book_service: update: %w", err)
	}

	// Clean up old image if it was replaced
	if input.ImageURL != nil && *input.ImageURL != oldImageURL && oldImageURL != "" {
		if strings.HasPrefix(oldImageURL, "/uploads/") {
			oldPath := filepath.Join("frontend", "public", "uploads", filepath.Base(oldImageURL))
			_ = os.Remove(oldPath)
		}
	}

	return book, nil
}

// DeleteBook removes a book by ID.
func (s *BookService) DeleteBook(id string) error {
	book, err := s.bookRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("book_service: find before delete: %w", err)
	}
	if book == nil {
		return nil // already deleted
	}

	if err := s.bookRepo.Delete(id); err != nil {
		return err
	}

	// Clean up image
	if book.ImageURL != "" && strings.HasPrefix(book.ImageURL, "/uploads/") {
		oldPath := filepath.Join("frontend", "public", "uploads", filepath.Base(book.ImageURL))
		_ = os.Remove(oldPath)
	}

	return nil
}
