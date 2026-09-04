// Package repository provides database access for all domain entities.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

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
		`SELECT id, email, password_hash, name, role, is_active, phone_number, address, province, city, regency, village, zip_code, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.PhoneNumber, &u.Address, &u.Province, &u.City, &u.Regency, &u.Village, &u.ZipCode, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by email: %w", err)
	}
	u.District = u.Regency
	return u, nil
}

// FindByID retrieves a user by their ID.
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, name, role, is_active, phone_number, address, province, city, regency, village, zip_code, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive, &u.PhoneNumber, &u.Address, &u.Province, &u.City, &u.Regency, &u.Village, &u.ZipCode, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by id: %w", err)
	}
	u.District = u.Regency
	return u, nil
}

// Create inserts a new user record.
func (r *UserRepository) Create(u *model.User) error {
	isActive := 0
	if u.IsActive {
		isActive = 1
	}
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, name, role, is_active, phone_number, address, province, city, regency, village, zip_code, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		u.ID, u.Email, u.PasswordHash, u.Name, u.Role, isActive, u.PhoneNumber, u.Address, u.Province, u.City, u.Regency, u.Village, u.ZipCode, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("user_repo: create: %w", err)
	}
	return nil
}

// Update modifies an existing user record.
func (r *UserRepository) Update(u *model.User) error {
	u.UpdatedAt = time.Now().UTC()
	isActive := 0
	if u.IsActive {
		isActive = 1
	}
	_, err := r.db.Exec(
		`UPDATE users SET email = $1, name = $2, role = $3, is_active = $4, phone_number = $5, address = $6, province = $7, city = $8, regency = $9, village = $10, zip_code = $11, updated_at = $12 WHERE id = $13`,
		u.Email, u.Name, u.Role, isActive, u.PhoneNumber, u.Address, u.Province, u.City, u.Regency, u.Village, u.ZipCode, u.UpdatedAt, u.ID,
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
		`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`,
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
		`SELECT id, email, name, role, is_active, phone_number, address, province, city, regency, village, zip_code, created_at, updated_at
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("user_repo: find all: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsActive, &u.PhoneNumber, &u.Address, &u.Province, &u.City, &u.Regency, &u.Village, &u.ZipCode, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("user_repo: scan user: %w", err)
		}
		u.District = u.Regency
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user_repo: rows err: %w", err)
	}
	return users, nil
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("user_repo: delete: %w", err)
	}
	return nil
}

// GetRegionName resolves a region code to its corresponding name from the database.
func (r *UserRepository) GetRegionName(table string, code string) string {
	if code == "" || code == "-" {
		return "-"
	}
	var name string
	if table != "provinces" && table != "regencies" && table != "districts" && table != "villages" {
		return "-"
	}
	query := fmt.Sprintf("SELECT name FROM %s WHERE code = $1 LIMIT 1", table)
	err := r.db.QueryRow(query, code).Scan(&name)
	if err != nil {
		return code
	}
	return name
}

// SaveOTP stores the generated OTP for the email into verify_otps table.
func (r *UserRepository) SaveOTP(email, otp string, expiredAt time.Time) error {
	// Clean up any existing OTPs for this email to avoid stale entries
	_, _ = r.db.Exec("DELETE FROM verify_otps WHERE email = $1", email)

	id := uuid.New().String()
	_, err := r.db.Exec(
		"INSERT INTO verify_otps (id, email, otp, expired_at) VALUES ($1, $2, $3, $4)",
		id, email, otp, expiredAt,
	)
	if err != nil {
		return fmt.Errorf("user_repo: save otp: %w", err)
	}
	return nil
}

// VerifyAndConsumeOTP checks if the OTP matches and is not expired.
// If valid, it deletes the entry from verify_otps as requested.
func (r *UserRepository) VerifyAndConsumeOTP(email, otp string) (bool, error) {
	var id string
	var expiredAt time.Time
	err := r.db.QueryRow(
		"SELECT id, expired_at FROM verify_otps WHERE email = $1 AND otp = $2 ORDER BY expired_at DESC LIMIT 1",
		email, otp,
	).Scan(&id, &expiredAt)

	if err == sql.ErrNoRows {
		return false, errors.New("OTP tidak ditemukan atau kode salah")
	}
	if err != nil {
		return false, fmt.Errorf("user_repo: verify otp: %w", err)
	}

	if time.Now().After(expiredAt) {
		_, _ = r.db.Exec("DELETE FROM verify_otps WHERE id = $1", id)
		return false, errors.New("OTP sudah kedaluwarsa")
	}

	// Sesuai requirement: jika otp berhasil diverifikasi dan sesuai maka hapus entri yang ada di tabel "verify_otps"
	_, err = r.db.Exec("DELETE FROM verify_otps WHERE id = $1", id)
	if err != nil {
		return false, fmt.Errorf("user_repo: delete consumed otp: %w", err)
	}

	return true, nil
}
