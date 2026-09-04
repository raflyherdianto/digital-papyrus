package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"

	"github.com/google/uuid"
)

type CoreServiceRepository struct {
	db *sql.DB
}

func NewCoreServiceRepository(db *sql.DB) *CoreServiceRepository {
	return &CoreServiceRepository{db: db}
}

func (r *CoreServiceRepository) FindAll() ([]*model.CoreService, error) {
	query := `SELECT id, title, description, icon, sort_order, is_active, created_at, updated_at 
			  FROM core_services ORDER BY sort_order ASC, created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("core_service repo: query all: %w", err)
	}
	defer rows.Close()

	var coreServices []*model.CoreService
	for rows.Next() {
		s := &model.CoreService{}
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.Icon,
			&s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("core_service repo: scan: %w", err)
		}
		coreServices = append(coreServices, s)
	}
	return coreServices, nil
}

func (r *CoreServiceRepository) FindByID(id string) (*model.CoreService, error) {
	query := `SELECT id, title, description, icon, sort_order, is_active, created_at, updated_at 
			  FROM core_services WHERE id = $1`
	
	s := &model.CoreService{}
	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.Title, &s.Description, &s.Icon,
		&s.SortOrder, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("core_service repo: find by id: %w", err)
	}
	return s, nil
}

func (r *CoreServiceRepository) Create(s *model.CoreService) error {
	s.ID = uuid.New().String()
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	query := `INSERT INTO core_services 
		(id, title, description, icon, sort_order, is_active, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := r.db.Exec(query,
		s.ID, s.Title, s.Description, s.Icon,
		s.SortOrder, s.IsActive, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("core_service repo: create: %w", err)
	}
	return nil
}

func (r *CoreServiceRepository) Update(s *model.CoreService) error {
	s.UpdatedAt = time.Now().UTC()

	query := `UPDATE core_services SET 
		title = $1, description = $2, icon = $3, sort_order = $4, is_active = $5, updated_at = $6 
		WHERE id = $7`
	
	_, err := r.db.Exec(query,
		s.Title, s.Description, s.Icon, s.SortOrder, s.IsActive, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("core_service repo: update: %w", err)
	}
	return nil
}

func (r *CoreServiceRepository) Delete(id string) error {
	query := `DELETE FROM core_services WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("core_service repo: delete: %w", err)
	}
	return nil
}
