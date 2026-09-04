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
	argIdx := 1

	if f.Status != "" {
		where = append(where, fmt.Sprintf("b.status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.CategoryID != "" {
		where = append(where, fmt.Sprintf("b.category_id = $%d", argIdx))
		args = append(args, f.CategoryID)
		argIdx++
	}
	if f.Badge != "" {
		normalizedBadge := strings.TrimSpace(f.Badge)
		if normalizedBadge == "" {
			normalizedBadge = "Regular"
		}
		where = append(where, fmt.Sprintf("COALESCE(b.badge, 'Regular') = $%d", argIdx))
		args = append(args, normalizedBadge)
		argIdx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(b.title ILIKE $%d OR b.author ILIKE $%d OR b.isbn ILIKE $%d)", argIdx, argIdx+1, argIdx+2))
		searchTerm := "%" + f.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		argIdx += 3
	}

	if f.MinPrice > 0 {
		where = append(where, fmt.Sprintf("b.price >= $%d", argIdx))
		args = append(args, f.MinPrice)
		argIdx++
	}
	if f.MaxPrice > 0 {
		where = append(where, fmt.Sprintf("b.price <= $%d", argIdx))
		args = append(args, f.MaxPrice)
		argIdx++
	}
	if f.MaxRating > 0 {
		where = append(where, fmt.Sprintf(`
			COALESCE((SELECT AVG((j.value)::numeric) FROM reviews rv,
				jsonb_each_text(CASE WHEN rv.rating LIKE '{%%}' THEN rv.rating::jsonb ELSE '{}'::jsonb END) j
				WHERE j.key = 'book_' || b.id), 0) <= $%d`, argIdx))
		args = append(args, f.MaxRating)
		argIdx++
	}

	ratingExpr := `COALESCE((SELECT AVG((j.value)::numeric) FROM reviews rv,
		jsonb_each_text(CASE WHEN rv.rating LIKE '{%}' THEN rv.rating::jsonb ELSE '{}'::jsonb END) j
		WHERE j.key = 'book_' || b.id), 0)`
	reviewCountExpr := `(SELECT COUNT(*) FROM reviews rv, jsonb_array_elements_text(CASE WHEN rv.book_id LIKE '[%]' THEN rv.book_id::jsonb ELSE '[]'::jsonb END) j WHERE j.value = b.id)`

	whereClause := strings.Join(where, " AND ")

	orderClause := "b.updated_at DESC"
	if f.Sort == "popular" {
		where = append(where, fmt.Sprintf("%s >= 4.5", ratingExpr))
		whereClause = strings.Join(where, " AND ")
		orderClause = fmt.Sprintf("%s DESC, %s DESC, b.updated_at DESC", ratingExpr, reviewCountExpr)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE %s", whereClause)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("book_repo: count: %w", err)
	}

	offset := (f.Page - 1) * f.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT b.id, b.title, b.author, COALESCE(b.isbn, ''), COALESCE(b.badge, 'Regular'), COALESCE(b.ggkey, ''), COALESCE(b.qrcbn, ''), COALESCE(b.price, 0), 
        COALESCE((SELECT AVG((j.value)::numeric) FROM reviews r, jsonb_each_text(CASE WHEN r.rating LIKE '{%%}' THEN r.rating::jsonb ELSE '{}'::jsonb END) j WHERE j.key = 'book_' || b.id), 0), 
        (SELECT COUNT(*) FROM reviews r, jsonb_array_elements_text(CASE WHEN r.book_id LIKE '[%%]' THEN r.book_id::jsonb ELSE '[]'::jsonb END) j WHERE j.value = b.id),
    COALESCE(b.description, ''), COALESCE(b.image_url, ''), b.category_id, COALESCE(c.name, ''), COALESCE(c.slug, ''), COALESCE(b.status, 'draft'), COALESCE(b.stock, 0),
        COALESCE(b.publisher, ''), COALESCE(b.publication_date, ''), COALESCE(b.pages, 0), COALESCE(b.format, ''), COALESCE(b.language, ''), COALESCE(b.dimensions, ''), COALESCE(b.weight, ''),
        b.created_at, b.updated_at, COALESCE(b.user_id, ''), COALESCE(b.order_id, ''), COALESCE(b.draft_url, ''), COALESCE(b.validation_status, 'pending'), COALESCE(b.notes, ''), COALESCE(b.amazon_url, ''), COALESCE(b.gplay_books_url, ''), COALESCE(b.production_cost, 0), COALESCE(b.royalty_fee, 0)
 FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`, whereClause, orderClause, argIdx, argIdx+1)
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
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Badge, &b.GGKEY, &b.QRCBN, &b.Price, &b.Rating, &b.ReviewCount,
			&b.Description, &b.ImageURL, &categoryID, &b.CategoryName, &b.CategorySlug, &b.Status, &b.Stock,
			&b.Publisher, &b.PublicationDate, &b.Pages, &b.Format, &b.Language, &b.Dimensions, &b.Weight,
			&b.CreatedAt, &b.UpdatedAt, &b.UserID, &b.OrderID, &b.DraftURL, &b.ValidationStatus, &b.Notes, &b.AmazonURL, &b.GPlayBooksURL,
			&b.ProductionCost, &b.RoyaltyFee,
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
		`SELECT b.id, b.title, b.author, COALESCE(b.isbn, ''), COALESCE(b.badge, 'Regular'), COALESCE(b.ggkey, ''), COALESCE(b.qrcbn, ''), COALESCE(b.price, 0), 
        COALESCE((SELECT AVG((j.value)::numeric) FROM reviews r, jsonb_each_text(CASE WHEN r.rating LIKE '{%}' THEN r.rating::jsonb ELSE '{}'::jsonb END) j WHERE j.key = 'book_' || b.id), 0), 
        (SELECT COUNT(*) FROM reviews r, jsonb_array_elements_text(CASE WHEN r.book_id LIKE '[%]' THEN r.book_id::jsonb ELSE '[]'::jsonb END) j WHERE j.value = b.id),
        COALESCE(b.description, ''), COALESCE(b.image_url, ''), b.category_id, COALESCE(c.name, ''), COALESCE(c.slug, ''), COALESCE(b.status, 'draft'), COALESCE(b.stock, 0),
        COALESCE(b.publisher, ''), COALESCE(b.publication_date, ''), COALESCE(b.pages, 0), COALESCE(b.format, ''), COALESCE(b.language, ''), COALESCE(b.dimensions, ''), COALESCE(b.weight, ''),
        b.created_at, b.updated_at, COALESCE(b.user_id, ''), COALESCE(b.order_id, ''), COALESCE(b.draft_url, ''), COALESCE(b.validation_status, 'pending'), COALESCE(b.notes, ''), COALESCE(b.amazon_url, ''), COALESCE(b.gplay_books_url, ''), COALESCE(b.production_cost, 0), COALESCE(b.royalty_fee, 0)
 FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.id = $1`, id,
	).Scan(
		&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Badge, &b.GGKEY, &b.QRCBN, &b.Price, &b.Rating, &b.ReviewCount,
		&b.Description, &b.ImageURL, &categoryID, &b.CategoryName, &b.CategorySlug, &b.Status, &b.Stock,
		&b.Publisher, &b.PublicationDate, &b.Pages, &b.Format, &b.Language, &b.Dimensions, &b.Weight,
		&b.CreatedAt, &b.UpdatedAt, &b.UserID, &b.OrderID, &b.DraftURL, &b.ValidationStatus, &b.Notes, &b.AmazonURL, &b.GPlayBooksURL,
		&b.ProductionCost, &b.RoyaltyFee,
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

	isbn := normalizeISBN(b.ISBN)
	badge := normalizeBadge(b.Badge)

	_, err := r.db.Exec(
		`INSERT INTO books (
id, title, author, isbn, badge, ggkey, qrcbn, price, rating, review_count,
description, image_url, category_id, status, stock,
publisher, publication_date, pages, format, language, dimensions, weight,
created_at, updated_at, user_id, order_id, draft_url, validation_status, notes, amazon_url, gplay_books_url, production_cost, royalty_fee
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33)`,
		b.ID, b.Title, b.Author, isbn, badge, b.GGKEY, b.QRCBN, b.Price, b.Rating, b.ReviewCount,
		b.Description, b.ImageURL, categoryID, b.Status, b.Stock,
		b.Publisher, b.PublicationDate, b.Pages, b.Format, b.Language, b.Dimensions, b.Weight,
		b.CreatedAt, b.UpdatedAt, b.UserID, b.OrderID, b.DraftURL, b.ValidationStatus, b.Notes, b.AmazonURL, b.GPlayBooksURL, b.ProductionCost, b.RoyaltyFee,
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
title = $1, author = $2, isbn = $3, badge = $4, ggkey = $5, qrcbn = $6, price = $7, rating = $8, review_count = $9,
description = $10, image_url = $11, category_id = $12, status = $13, stock = $14,
publisher = $15, publication_date = $16, pages = $17, format = $18, language = $19,
dimensions = $20, weight = $21, updated_at = $22, user_id = $23, order_id = $24, draft_url = $25, validation_status = $26, notes = $27, amazon_url = $28, gplay_books_url = $29,
production_cost = $30, royalty_fee = $31
 WHERE id = $32`,
		b.Title, b.Author, isbn, badge, b.GGKEY, b.QRCBN, b.Price, b.Rating, b.ReviewCount,
		b.Description, b.ImageURL, categoryID, b.Status, b.Stock,
		b.Publisher, b.PublicationDate, b.Pages, b.Format, b.Language,
		b.Dimensions, b.Weight, b.UpdatedAt, b.UserID, b.OrderID, b.DraftURL, b.ValidationStatus, b.Notes, b.AmazonURL, b.GPlayBooksURL,
		b.ProductionCost, b.RoyaltyFee, b.ID,
	)
	if err != nil {
		return fmt.Errorf("book_repo: update: %w", err)
	}
	return nil
}

// Delete removes a book by its ID.
func (r *BookRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("book_repo: delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("book_repo: book not found")
	}
	return nil
}

// parseDateTime parses ISO and standard datetime strings into time.Time.
// Supports multiple standard datetime formats.
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
