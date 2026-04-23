package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

// BookRepository handles book database operations.
type BookRepository struct {
	db *sql.DB
}

// NewBookRepository creates a new BookRepository.
func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

// BookFilter defines query parameters for listing books.
type BookFilter struct {
	Status     string
	CategoryID string
	Search     string
	Page       int
	PerPage    int
}

// FindAll retrieves a paginated list of books with optional filters.
func (r *BookRepository) FindAll(f BookFilter) ([]model.Book, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 12
	}

	where := []string{"1=1"}
	args := []interface{}{}

	if f.Status != "" {
		where = append(where, "b.status = ?")
		args = append(args, f.Status)
	}
	if f.CategoryID != "" {
		where = append(where, "b.category_id = ?")
		args = append(args, f.CategoryID)
	}
	if f.Search != "" {
		where = append(where, "(b.title LIKE ? OR b.author LIKE ? OR b.isbn LIKE ?)")
		searchTerm := "%" + f.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM books b WHERE %s", whereClause)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("book_repo: count: %w", err)
	}

	// Fetch page
	offset := (f.Page - 1) * f.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT b.id, b.title, b.author, b.isbn, b.price, b.rating, b.review_count,
    b.description, b.synopsis, b.image_url, b.category_id, COALESCE(c.name, ''), COALESCE(c.slug, ''), b.status, b.stock,
        b.publisher, b.publication_date, b.pages, b.format, b.language, b.dimensions, b.weight,
        b.created_at, b.updated_at
 FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE %s ORDER BY b.created_at DESC LIMIT ? OFFSET ?`, whereClause)
	dataArgs := append(args, f.PerPage, offset)

	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("book_repo: query: %w", err)
	}
	defer rows.Close()

	books := make([]model.Book, 0)
	for rows.Next() {
		var b model.Book
		var categoryID sql.NullString
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Price, &b.Rating, &b.ReviewCount,
			&b.Description, &b.Synopsis, &b.ImageURL, &categoryID, &b.CategoryName, &b.CategorySlug, &b.Status, &b.Stock,
			&b.Publisher, &b.PublicationDate, &b.Pages, &b.Format, &b.Language, &b.Dimensions, &b.Weight,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("book_repo: scan: %w", err)
		}
		if categoryID.Valid {
			b.CategoryID = categoryID.String
		}
		books = append(books, b)
	}
	return books, total, rows.Err()
}

// FindByID retrieves a single book by its ID.
func (r *BookRepository) FindByID(id string) (*model.Book, error) {
	b := &model.Book{}
	var categoryID sql.NullString
	err := r.db.QueryRow(
		`SELECT b.id, b.title, b.author, b.isbn, b.price, b.rating, b.review_count,
        b.description, b.synopsis, b.image_url, b.category_id, COALESCE(c.name, ''), COALESCE(c.slug, ''), b.status, b.stock,
        b.publisher, b.publication_date, b.pages, b.format, b.language, b.dimensions, b.weight,
        b.created_at, b.updated_at
 FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.id = ?`, id,
	).Scan(
		&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Price, &b.Rating, &b.ReviewCount,
		&b.Description, &b.Synopsis, &b.ImageURL, &categoryID, &b.CategoryName, &b.CategorySlug, &b.Status, &b.Stock,
		&b.Publisher, &b.PublicationDate, &b.Pages, &b.Format, &b.Language, &b.Dimensions, &b.Weight,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("book_repo: find by id: %w", err)
	}
	if categoryID.Valid {
		b.CategoryID = categoryID.String
	}
	return b, nil
}

// Create inserts a new book record.
func (r *BookRepository) Create(b *model.Book) error {
	now := time.Now().UTC()
	b.CreatedAt = now
	b.UpdatedAt = now

	var categoryID interface{} = b.CategoryID
	if b.CategoryID == "" {
		categoryID = nil
	}

	_, err := r.db.Exec(
		`INSERT INTO books (
id, title, author, isbn, price, rating, review_count,
description, synopsis, image_url, category_id, status, stock,
publisher, publication_date, pages, format, language, dimensions, weight,
created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.Title, b.Author, b.ISBN, b.Price, b.Rating, b.ReviewCount,
		b.Description, b.Synopsis, b.ImageURL, categoryID, b.Status, b.Stock,
		b.Publisher, b.PublicationDate, b.Pages, b.Format, b.Language, b.Dimensions, b.Weight,
		b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("book_repo: create: %w", err)
	}
	return nil
}

// Update modifies an existing book record.
func (r *BookRepository) Update(b *model.Book) error {
	b.UpdatedAt = time.Now().UTC()

	var categoryID interface{} = b.CategoryID
	if b.CategoryID == "" {
		categoryID = nil
	}

	_, err := r.db.Exec(
		`UPDATE books SET
title = ?, author = ?, isbn = ?, price = ?, rating = ?, review_count = ?,
description = ?, synopsis = ?, image_url = ?, category_id = ?, status = ?, stock = ?,
publisher = ?, publication_date = ?, pages = ?, format = ?, language = ?,
dimensions = ?, weight = ?, updated_at = ?
 WHERE id = ?`,
		b.Title, b.Author, b.ISBN, b.Price, b.Rating, b.ReviewCount,
		b.Description, b.Synopsis, b.ImageURL, categoryID, b.Status, b.Stock,
		b.Publisher, b.PublicationDate, b.Pages, b.Format, b.Language,
		b.Dimensions, b.Weight, b.UpdatedAt, b.ID,
	)
	if err != nil {
		return fmt.Errorf("book_repo: update: %w", err)
	}
	return nil
}

// Delete removes a book by its ID.
func (r *BookRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("book_repo: delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("book_repo: book not found")
	}
	return nil
}
