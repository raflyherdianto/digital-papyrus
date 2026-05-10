package model

import "time"

type CoreService struct {
	ID          string    `json:"id"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Icon        string    `json:"icon"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
