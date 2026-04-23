package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

// CategoryRepository handles category database operations.
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FindAll retrieves all categories.
func (r *CategoryRepository) FindAll() ([]model.Category, error) {
	rows, err := r.db.Query(`SELECT id, name, slug, created_at, updated_at FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("category_repo: find all: %w", err)
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("category_repo: scan: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// FindByID retrieves a single category by its ID.
func (r *CategoryRepository) FindByID(id string) (*model.Category, error) {
	var c model.Category
	err := r.db.QueryRow(`SELECT id, name, slug, created_at, updated_at FROM categories WHERE id = ?`, id).
		Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("category_repo: find by id: %w", err)
	}
	return &c, nil
}

// FindBySlug retrieves a single category by slug.
func (r *CategoryRepository) FindBySlug(slug string) (*model.Category, error) {
	var c model.Category
	err := r.db.QueryRow(`SELECT id, name, slug, created_at, updated_at FROM categories WHERE slug = ?`, slug).
		Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("category_repo: find by slug: %w", err)
	}
	return &c, nil
}

// Create inserts a new category record.
func (r *CategoryRepository) Create(c *model.Category) error {
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err := r.db.Exec(`INSERT INTO categories (id, name, slug, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Slug, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("category_repo: create: %w", err)
	}
	return nil
}

// Update modifies an existing category record.
func (r *CategoryRepository) Update(c *model.Category) error {
	c.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(`UPDATE categories SET name = ?, slug = ?, updated_at = ? WHERE id = ?`,
		c.Name, c.Slug, c.UpdatedAt, c.ID)
	if err != nil {
		return fmt.Errorf("category_repo: update: %w", err)
	}
	return nil
}

// Delete removes a category by its ID.
func (r *CategoryRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("category_repo: delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("category_repo: not found")
	}
	return nil
}
