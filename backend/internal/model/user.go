// Package model defines domain entities for the Digital Papyrus application.
package model

import "time"

// Role constants define the allowed user roles.
const (
	RoleSuperAdmin = "superadmin"
	RoleAuthor     = "author"
	RoleCustomer   = "customer"
)

// User represents an authenticated user of the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never expose in JSON
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ValidRoles returns all valid user role values.
func ValidRoles() []string {
	return []string{RoleSuperAdmin, RoleAuthor, RoleCustomer}
}

// IsValidRole checks if the provided role string is valid.
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if r == role {
			return true
		}
	}
	return false
}
