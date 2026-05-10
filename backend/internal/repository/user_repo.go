// Package repository provides database access for all domain entities.
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

// UserRepository handles user database operations.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail retrieves a user by their email address.
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, name, role, is_active, phone_number, address, province, city, zip_code, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.PhoneNumber, &u.Address, &u.Province, &u.City, &u.ZipCode, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by email: %w", err)
	}
	return u, nil
}

// FindByID retrieves a user by their ID.
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, name, role, is_active, phone_number, address, province, city, zip_code, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.PhoneNumber, &u.Address, &u.Province, &u.City, &u.ZipCode, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by id: %w", err)
	}
	return u, nil
}

// Create inserts a new user record.
func (r *UserRepository) Create(u *model.User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, name, role, is_active, phone_number, address, province, city, zip_code, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Role, u.IsActive, u.PhoneNumber, u.Address, u.Province, u.City, u.ZipCode, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("user_repo: create: %w", err)
	}
	return nil
}

// Update modifies an existing user record.
func (r *UserRepository) Update(u *model.User) error {
	u.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE users SET email = ?, name = ?, role = ?, is_active = ?, phone_number = ?, address = ?, province = ?, city = ?, zip_code = ?, updated_at = ? WHERE id = ?`,
		u.Email, u.Name, u.Role, u.IsActive, u.PhoneNumber, u.Address, u.Province, u.City, u.ZipCode, u.UpdatedAt, u.ID,
	)
	if err != nil {
		return fmt.Errorf("user_repo: update: %w", err)
	}
	return nil
}

// UpdatePassword modifies an existing user's password hash.
func (r *UserRepository) UpdatePassword(id string, passwordHash string) error {
	updatedAt := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("user_repo: update password: %w", err)
	}
	return nil
}

// FindAll retrieves all users.
func (r *UserRepository) FindAll() ([]model.User, error) {
	rows, err := r.db.Query(
		`SELECT id, email, name, role, is_active, phone_number, address, province, city, zip_code, created_at, updated_at
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("user_repo: find all: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsActive, &u.PhoneNumber, &u.Address, &u.Province, &u.City, &u.ZipCode, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("user_repo: scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user_repo: rows err: %w", err)
	}
	return users, nil
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("user_repo: delete: %w", err)
	}
	return nil
}
