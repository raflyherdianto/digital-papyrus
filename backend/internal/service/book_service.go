package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/pkg/validator"
)

// BookService handles book business logic.
type BookService struct {
	bookRepo *repository.BookRepository
	userRepo *repository.UserRepository
	orderRepo *repository.OrderRepository
	cfg      *config.Config
}

// NewBookService creates a new BookService.
func NewBookService(bookRepo *repository.BookRepository, userRepo *repository.UserRepository, orderRepo *repository.OrderRepository, cfg *config.Config) *BookService {
	return &BookService{
		bookRepo:  bookRepo,
		userRepo:  userRepo,
		orderRepo: orderRepo,
		cfg:       cfg,
	}
}

// CreateBookInput represents the request body for creating a book.
type CreateBookInput struct {
	Title           string  `json:"title"`
	Author          string  `json:"author"`
	ISBN            string  `json:"isbn"`
	Badge           string  `json:"badge"`
	GGKEY           string  `json:"ggkey"`
	QRCBN           string  `json:"qrcbn"`
	Price           int     `json:"price"`
	Rating          float64 `json:"rating"`
	Description     string  `json:"description"`
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
	ValidationStatus string `json:"validation_status"`
	AmazonURL       string  `json:"amazon_url"`
	GPlayBooksURL   string  `json:"gplay_books_url"`
	ProductionCost  int     `json:"production_cost"`
	RoyaltyFee      int     `json:"royalty_fee"`
}

// Validate checks all required fields and business rules.
func (i *CreateBookInput) Validate() map[string]string {
	errs := make(map[string]string)
	i.Title = validator.SanitizeString(i.Title)
	i.Author = validator.SanitizeString(i.Author)
	i.Badge = strings.TrimSpace(i.Badge)
	i.GGKEY = strings.TrimSpace(i.GGKEY)
	i.QRCBN = strings.TrimSpace(i.QRCBN)

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
		Badge:           input.Badge,
		GGKEY:           input.GGKEY,
		QRCBN:           input.QRCBN,
		Price:           input.Price,
		Rating:          input.Rating,
		Description:     input.Description,
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
		ValidationStatus: input.ValidationStatus,
		AmazonURL:       input.AmazonURL,
		GPlayBooksURL:   input.GPlayBooksURL,
		ProductionCost:  input.ProductionCost,
		RoyaltyFee:      input.RoyaltyFee,
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
	Badge           *string  `json:"badge"`
	GGKEY           *string  `json:"ggkey"`
	QRCBN           *string  `json:"qrcbn"`
	Price           *int     `json:"price"`
	Rating          *float64 `json:"rating"`
	ReviewCount     *int     `json:"review_count"`
	Description     *string  `json:"description"`
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
	AmazonURL       *string  `json:"amazon_url"`
	GPlayBooksURL   *string  `json:"gplay_books_url"`
	ProductionCost  *int     `json:"production_cost"`
	RoyaltyFee      *int     `json:"royalty_fee"`
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
	if input.Badge != nil {
		book.Badge = strings.TrimSpace(*input.Badge)
	}
	if input.GGKEY != nil {
		book.GGKEY = strings.TrimSpace(*input.GGKEY)
	}
	if input.QRCBN != nil {
		book.QRCBN = strings.TrimSpace(*input.QRCBN)
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
	if input.AmazonURL != nil {
		book.AmazonURL = *input.AmazonURL
	}
	if input.GPlayBooksURL != nil {
		book.GPlayBooksURL = *input.GPlayBooksURL
	}
	if input.ProductionCost != nil {
		book.ProductionCost = *input.ProductionCost
	}
	if input.RoyaltyFee != nil {
		book.RoyaltyFee = *input.RoyaltyFee
	}

	if err := s.bookRepo.Update(book); err != nil {
		return nil, fmt.Errorf("book_service: update: %w", err)
	}

	// Clean up old image if it was replaced
	if input.ImageURL != nil && *input.ImageURL != oldImageURL && oldImageURL != "" {
		if strings.HasPrefix(oldImageURL, "/uploads/") {
			oldPath := getUploadPath(oldImageURL)
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
		oldPath := getUploadPath(book.ImageURL)
		_ = os.Remove(oldPath)
	}

	// Clean up draft
	if book.DraftURL != "" && strings.HasPrefix(book.DraftURL, "/uploads/drafts/") {
		oldDraftPath := getDraftUploadPath(book.DraftURL)
		_ = os.Remove(oldDraftPath)
	}

	return nil
}

func getUploadPath(imageURL string) string {
	baseName := filepath.Base(imageURL)
	if _, err := os.Stat(filepath.Join("..", "frontend", "public")); err == nil {
		return filepath.Join("..", "frontend", "public", "uploads", baseName)
	}
	return filepath.Join("frontend", "public", "uploads", baseName)
}

func getDraftUploadPath(draftURL string) string {
	baseName := filepath.Base(draftURL)
	if _, err := os.Stat(filepath.Join("..", "frontend", "public")); err == nil {
		return filepath.Join("..", "frontend", "public", "uploads", "drafts", baseName)
	}
	return filepath.Join("frontend", "public", "uploads", "drafts", baseName)
}

// CreateCustomerBookInput represents the input for customer book draft creation.
type CreateCustomerBookInput struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	Publisher   string `json:"publisher"`
	Format      string `json:"format"`
	Language    string `json:"language"`
	Pages       int    `json:"pages"`
	Dimensions  string `json:"dimensions"`
	Weight      string `json:"weight"`
	DraftURL    string `json:"draft_url"`
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"` // injected from auth context
}

// Validate checks fields submitted by a customer.
func (i *CreateCustomerBookInput) Validate() map[string]string {
	errs := make(map[string]string)
	i.Title = validator.SanitizeString(i.Title)
	i.Author = validator.SanitizeString(i.Author)

	if i.Title == "" {
		errs["title"] = "Title is required"
	}
	if i.Author == "" {
		errs["author"] = "Author is required"
	}
	if i.CategoryID == "" {
		errs["category_id"] = "Category is required"
	}
	if i.OrderID == "" {
		errs["order_id"] = "Order/Invoice selection is required"
	}
	if i.DraftURL == "" {
		errs["draft_url"] = "Book draft file is required"
	}
	return errs
}

// CreateCustomerBook creates a new book draft from a customer submission.
func (s *BookService) CreateCustomerBook(input CreateCustomerBookInput) (*model.Book, error) {
	book := &model.Book{
		ID:               uuid.New().String(),
		Title:            input.Title,
		Author:           input.Author,
		Description:      input.Description,
		Publisher:        input.Publisher,
		CategoryID:       input.CategoryID,
		Format:           input.Format,
		Language:         input.Language,
		Pages:            input.Pages,
		Dimensions:       input.Dimensions,
		Weight:           input.Weight,
		DraftURL:         input.DraftURL,
		OrderID:          input.OrderID,
		UserID:           input.UserID,
		Status:           model.BookStatusDraft,
		ValidationStatus: "pending",
		Price:            0,
		Rating:           0.0,
		ReviewCount:      0,
		Stock:            0,
	}

	if err := s.bookRepo.Create(book); err != nil {
		return nil, fmt.Errorf("book_service: create customer book: %w", err)
	}

	// Send email notification asynchronously
	go func() {
		if err := s.SendDraftSubmissionEmail(book); err != nil {
			fmt.Printf("Failed to send draft submission email: %v\n", err)
		}
	}()

	return book, nil
}

// ValidateBook updates the validation status of a book request.
func (s *BookService) ValidateBook(id string, validationStatus string, notes string) (*model.Book, error) {
	book, err := s.bookRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("book_service: %w", err)
	}
	if book == nil {
		return nil, nil
	}

	book.ValidationStatus = validationStatus
	if notes != "" {
		book.Notes = notes
	}

	if err := s.bookRepo.Update(book); err != nil {
		return nil, fmt.Errorf("book_service: validate update: %w", err)
	}

	// Send corresponding email asynchronously
	go func() {
		if validationStatus == "approved" {
			if err := s.SendValidationApproveEmail(book); err != nil {
				fmt.Printf("Failed to send approval email: %v\n", err)
			}
		} else if validationStatus == "rejected" {
			if err := s.SendValidationRejectEmail(book); err != nil {
				fmt.Printf("Failed to send rejection email: %v\n", err)
			}
		}
	}()

	return book, nil
}

// UpdateCustomerBookInput represents the payload for customer updating their draft.
type UpdateCustomerBookInput struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	Format      string `json:"format"`
	Language    string `json:"language"`
	DraftURL    string `json:"draft_url"`
}

// UpdateCustomerBookDraft updates book draft details and resets validation status.
func (s *BookService) UpdateCustomerBookDraft(id string, userID string, input UpdateCustomerBookInput) (*model.Book, error) {
	book, err := s.bookRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("book_service: %w", err)
	}
	if book == nil {
		return nil, nil
	}

	// Verify ownership
	if book.UserID != userID {
		return nil, fmt.Errorf("unauthorized to update this book draft")
	}

	book.Title = input.Title
	book.Author = input.Author
	book.CategoryID = input.CategoryID
	book.Description = input.Description
	book.Format = input.Format
	book.Language = input.Language

	// Clean up old draft file from storage if a new draft is uploaded
	if input.DraftURL != "" && input.DraftURL != book.DraftURL {
		if book.DraftURL != "" && strings.HasPrefix(book.DraftURL, "/uploads/drafts/") {
			oldDraftPath := getDraftUploadPath(book.DraftURL)
			if err := os.Remove(oldDraftPath); err == nil {
				fmt.Printf("[STORAGE CLEANUP] Successfully deleted old draft: %s\n", oldDraftPath)
			} else {
				fmt.Printf("[STORAGE CLEANUP] Warning: failed to delete old draft %s: %v\n", oldDraftPath, err)
			}
		}
		book.DraftURL = input.DraftURL
	}

	book.ValidationStatus = "pending"
	book.Notes = ""

	if err := s.bookRepo.Update(book); err != nil {
		return nil, fmt.Errorf("book_service: failed to update book draft: %w", err)
	}

	// Send email notification asynchronously
	go func() {
		if err := s.SendDraftSubmissionEmail(book); err != nil {
			fmt.Printf("Failed to send draft submission email: %v\n", err)
		}
	}()

	return book, nil
}
