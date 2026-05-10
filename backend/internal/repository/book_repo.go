package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

// normalizeISBN returns nil for placeholder/empty ISBN values so that
// the UNIQUE constraint allows multiple books without a real ISBN.
func normalizeISBN(isbn string) interface{} {
	trimmed := strings.TrimSpace(isbn)
	switch strings.ToLower(trimmed) {
	case "", "-", "--", "—", "n/a", "tbd", "belum terbit", "pending":
		return nil
	}
	return trimmed
}

func normalizeBadge(badge string) string {
	trimmed := strings.TrimSpace(badge)
	if trimmed == "" {
		return "Regular"
	}
	return trimmed
}

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
	Badge      string
	Search     string
	Page       int
	PerPage    int
	MinPrice   int
	MaxPrice   int    // 0 means no upper limit
	MaxRating  float64 // 0 means no filter
	Sort       string  // "newest", "popular" (rating), "" = newest
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
	if f.Badge != "" {
		// Normalize badge to ensure case-insensitive matching
		normalizedBadge := strings.TrimSpace(f.Badge)
		if normalizedBadge == "" {
			normalizedBadge = "Regular"
		}
		where = append(where, "COALESCE(b.badge, 'Regular') = ?")
		args = append(args, normalizedBadge)
	}
	if f.Search != "" {
		where = append(where, "(b.title LIKE ? OR b.author LIKE ? OR b.isbn LIKE ?)")
		searchTerm := "%" + f.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if f.MinPrice > 0 {
		where = append(where, "b.price >= ?")
		args = append(args, f.MinPrice)
	}
	if f.MaxPrice > 0 {
		where = append(where, "b.price <= ?")
		args = append(args, f.MaxPrice)
	}
	if f.MaxRating > 0 {
		where = append(where, `
			COALESCE((SELECT AVG(CAST(j.value AS REAL)) FROM reviews rv,
				json_each(CASE WHEN json_valid(rv.rating) THEN rv.rating ELSE '{}' END) j
				WHERE j.key = 'book_' || b.id), 0) <= ?`)
		args = append(args, f.MaxRating)
	}

	ratingExpr := `COALESCE((SELECT AVG(CAST(j.value AS REAL)) FROM reviews rv,
		json_each(CASE WHEN json_valid(rv.rating) THEN rv.rating ELSE '{}' END) j
		WHERE j.key = 'book_' || b.id), 0)`
	reviewCountExpr := `(SELECT COUNT(*) FROM reviews rv, json_each(CASE WHEN json_valid(rv.book_id) THEN rv.book_id ELSE '[]' END) j WHERE j.value = b.id)`

	whereClause := strings.Join(where, " AND ")

	// Determine ORDER BY
	orderClause := "b.updated_at DESC"
	if f.Sort == "popular" {
		where = append(where, fmt.Sprintf("%s >= 4.5", ratingExpr))
		whereClause = strings.Join(where, " AND ")
		orderClause = fmt.Sprintf("%s DESC, %s DESC, b.updated_at DESC", ratingExpr, reviewCountExpr)
	}

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE %s", whereClause)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("book_repo: count: %w", err)
	}

	// Fetch page — use COALESCE to guarantee non-NULL values for every column
	offset := (f.Page - 1) * f.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT b.id, b.title, b.author, COALESCE(b.isbn, ''), COALESCE(b.badge, 'Regular'), COALESCE(b.ggkey, ''), COALESCE(b.qrcbn, ''), COALESCE(b.price, 0), 
        COALESCE((SELECT AVG(CAST(j.value AS REAL)) FROM reviews r, json_each(CASE WHEN json_valid(r.rating) THEN r.rating ELSE '{}' END) j WHERE j.key = 'book_' || b.id), 0), 
        (SELECT COUNT(*) FROM reviews r, json_each(CASE WHEN json_valid(r.book_id) THEN r.book_id ELSE '[]' END) j WHERE j.value = b.id),
    COALESCE(b.description, ''), COALESCE(b.synopsis, ''), COALESCE(b.image_url, ''), b.category_id, COALESCE(c.name, ''), COALESCE(c.slug, ''), COALESCE(b.status, 'draft'), COALESCE(b.stock, 0),
        COALESCE(b.publisher, ''), COALESCE(b.publication_date, ''), COALESCE(b.pages, 0), COALESCE(b.format, ''), COALESCE(b.language, ''), COALESCE(b.dimensions, ''), COALESCE(b.weight, ''),
        COALESCE(b.created_at, datetime('now')), COALESCE(b.updated_at, datetime('now'))
 FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE %s ORDER BY %s LIMIT ? OFFSET ?`, whereClause, orderClause)
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
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Badge, &b.GGKEY, &b.QRCBN, &b.Price, &b.Rating, &b.ReviewCount,
			&b.Description, &b.Synopsis, &b.ImageURL, &categoryID, &b.CategoryName, &b.CategorySlug, &b.Status, &b.Stock,
			&b.Publisher, &b.PublicationDate, &b.Pages, &b.Format, &b.Language, &b.Dimensions, &b.Weight,
			&createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, 0, fmt.Errorf("book_repo: scan: %w", err)
		}
		if categoryID.Valid {
			b.CategoryID = categoryID.String
		}
		b.CreatedAt = parseDateTime(createdAtStr)
		b.UpdatedAt = parseDateTime(updatedAtStr)
		books = append(books, b)
	}
	return books, total, rows.Err()
}

// FindByID retrieves a single book by its ID.
func (r *BookRepository) FindByID(id string) (*model.Book, error) {
	b := &model.Book{}
	var categoryID sql.NullString
	var createdAtStr, updatedAtStr string
	err := r.db.QueryRow(
		`SELECT b.id, b.title, b.author, COALESCE(b.isbn, ''), COALESCE(b.badge, 'Regular'), COALESCE(b.ggkey, ''), COALESCE(b.qrcbn, ''), COALESCE(b.price, 0), 
        COALESCE((SELECT AVG(CAST(j.value AS REAL)) FROM reviews r, json_each(CASE WHEN json_valid(r.rating) THEN r.rating ELSE '{}' END) j WHERE j.key = 'book_' || b.id), 0), 
        (SELECT COUNT(*) FROM reviews r, json_each(CASE WHEN json_valid(r.book_id) THEN r.book_id ELSE '[]' END) j WHERE j.value = b.id),
        COALESCE(b.description, ''), COALESCE(b.synopsis, ''), COALESCE(b.image_url, ''), b.category_id, COALESCE(c.name, ''), COALESCE(c.slug, ''), COALESCE(b.status, 'draft'), COALESCE(b.stock, 0),
        COALESCE(b.publisher, ''), COALESCE(b.publication_date, ''), COALESCE(b.pages, 0), COALESCE(b.format, ''), COALESCE(b.language, ''), COALESCE(b.dimensions, ''), COALESCE(b.weight, ''),
        COALESCE(b.created_at, datetime('now')), COALESCE(b.updated_at, datetime('now'))
 FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.id = ?`, id,
	).Scan(
		&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Badge, &b.GGKEY, &b.QRCBN, &b.Price, &b.Rating, &b.ReviewCount,
		&b.Description, &b.Synopsis, &b.ImageURL, &categoryID, &b.CategoryName, &b.CategorySlug, &b.Status, &b.Stock,
		&b.Publisher, &b.PublicationDate, &b.Pages, &b.Format, &b.Language, &b.Dimensions, &b.Weight,
		&createdAtStr, &updatedAtStr,
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
	b.CreatedAt = parseDateTime(createdAtStr)
	b.UpdatedAt = parseDateTime(updatedAtStr)
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

	isbn := normalizeISBN(b.ISBN)
	badge := normalizeBadge(b.Badge)

	_, err := r.db.Exec(
		`INSERT INTO books (
id, title, author, isbn, badge, ggkey, qrcbn, price, rating, review_count,
description, synopsis, image_url, category_id, status, stock,
publisher, publication_date, pages, format, language, dimensions, weight,
created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.Title, b.Author, isbn, badge, b.GGKEY, b.QRCBN, b.Price, b.Rating, b.ReviewCount,
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

	isbn := normalizeISBN(b.ISBN)
	badge := normalizeBadge(b.Badge)

	_, err := r.db.Exec(
		`UPDATE books SET
title = ?, author = ?, isbn = ?, badge = ?, ggkey = ?, qrcbn = ?, price = ?, rating = ?, review_count = ?,
description = ?, synopsis = ?, image_url = ?, category_id = ?, status = ?, stock = ?,
publisher = ?, publication_date = ?, pages = ?, format = ?, language = ?,
dimensions = ?, weight = ?, updated_at = ?
 WHERE id = ?`,
		b.Title, b.Author, isbn, badge, b.GGKEY, b.QRCBN, b.Price, b.Rating, b.ReviewCount,
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

// parseDateTime parses SQLite datetime strings into time.Time.
// Supports multiple formats that SQLite's datetime() function may produce.
func parseDateTime(s string) time.Time {
	if s == "" {
		return time.Now().UTC()
	}
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Now().UTC()
}
