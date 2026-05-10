package model

import "time"

// Review represents a customer review in the system.
// It can handle multiple books and services in a single review entry.
type Review struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	OrderID   string            `json:"order_id"`
	Invoice   string            `json:"invoice"`    // Populated when retrieving reviews
	ServiceID []string          `json:"service_id"` // JSON array of service IDs
	BookID    []string          `json:"book_id"`    // JSON array of book IDs
	Details   map[string]string `json:"details"`    // Key: id_service/id_book, Value: comment
	Rating    map[string]int    `json:"rating"`     // Key: id_service/id_book, Value: 1-5 rating
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
