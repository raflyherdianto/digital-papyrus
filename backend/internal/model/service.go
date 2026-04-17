package model

import "time"

// ServiceTier constants define the allowed service tiers.
const (
	ServiceTierBasic    = "basic"
	ServiceTierSilver   = "silver"
	ServiceTierGold     = "gold"
	ServiceTierPlatinum = "platinum"
)

// Service represents a publishing service package offered to customers.
type Service struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Description string   `json:"description"`
	Icon       string    `json:"icon,omitempty"`
	Tier       string    `json:"tier"`
	Price      int       `json:"price"`       // price in smallest currency unit
	PriceLabel string    `json:"price_label"` // display label e.g. "Rp 275k"
	Features   string    `json:"features"`    // JSON array of feature strings
	IsFeatured bool      `json:"is_featured"`
	Badge      string    `json:"badge,omitempty"` // e.g. "Starter", "Terpopuler"
	SortOrder  int       `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
