package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

// ServiceRepository handles service database operations.
type ServiceRepository struct {
	db *sql.DB
}

// NewServiceRepository creates a new ServiceRepository.
func NewServiceRepository(db *sql.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// FindAll retrieves all active services ordered by sort_order.
func (r *ServiceRepository) FindAll(activeOnly bool) ([]model.Service, error) {
	query := `SELECT id, title, description, icon, tier, price, price_label,
	                 features, is_featured, badge, sort_order, is_active,
	                 created_at, updated_at
	          FROM services`
	if activeOnly {
		query += " WHERE is_active = 1"
	}
	query += " ORDER BY sort_order ASC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("service_repo: find all: %w", err)
	}
	defer rows.Close()

	services := make([]model.Service, 0)
	for rows.Next() {
		var s model.Service
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.Icon, &s.Tier,
			&s.Price, &s.PriceLabel, &s.Features, &s.IsFeatured,
			&s.Badge, &s.SortOrder, &s.IsActive,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("service_repo: scan: %w", err)
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

// FindByID retrieves a single service by its ID.
func (r *ServiceRepository) FindByID(id string) (*model.Service, error) {
	s := &model.Service{}
	err := r.db.QueryRow(
		`SELECT id, title, description, icon, tier, price, price_label,
		        features, is_featured, badge, sort_order, is_active,
		        created_at, updated_at
		 FROM services WHERE id = ?`, id,
	).Scan(
		&s.ID, &s.Title, &s.Description, &s.Icon, &s.Tier,
		&s.Price, &s.PriceLabel, &s.Features, &s.IsFeatured,
		&s.Badge, &s.SortOrder, &s.IsActive,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("service_repo: find by id: %w", err)
	}
	return s, nil
}

// Create inserts a new service record.
func (r *ServiceRepository) Create(s *model.Service) error {
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	_, err := r.db.Exec(
		`INSERT INTO services (
			id, title, description, icon, tier, price, price_label,
			features, is_featured, badge, sort_order, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Title, s.Description, s.Icon, s.Tier,
		s.Price, s.PriceLabel, s.Features, s.IsFeatured,
		s.Badge, s.SortOrder, s.IsActive,
		s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("service_repo: create: %w", err)
	}
	return nil
}

// Update modifies an existing service record.
func (r *ServiceRepository) Update(s *model.Service) error {
	s.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE services SET
			title = ?, description = ?, icon = ?, tier = ?, price = ?, price_label = ?,
			features = ?, is_featured = ?, badge = ?, sort_order = ?, is_active = ?,
			updated_at = ?
		 WHERE id = ?`,
		s.Title, s.Description, s.Icon, s.Tier, s.Price, s.PriceLabel,
		s.Features, s.IsFeatured, s.Badge, s.SortOrder, s.IsActive,
		s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("service_repo: update: %w", err)
	}
	return nil
}

// Delete removes a service by its ID.
func (r *ServiceRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM services WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("service_repo: delete: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("service_repo: service not found")
	}
	return nil
}
