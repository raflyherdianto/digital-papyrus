package model

import "time"

// BookStatus constants define the allowed publication states.
const (
	BookStatusDraft     = "draft"
	BookStatusPublished = "published"
	BookStatusArchived  = "archived"
)

// Book represents a published or draft book in the catalog.
type Book struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	ISBN            string    `json:"isbn,omitempty"`
	Price           int       `json:"price"`           // price in smallest currency unit (Rupiah)
	Rating          float64   `json:"rating"`
	ReviewCount     int       `json:"review_count"`
	Description     string    `json:"description,omitempty"`
	Synopsis        string    `json:"synopsis,omitempty"`
	ImageURL        string    `json:"image_url,omitempty"`
	CategoryID      string    `json:"category_id,omitempty"`
	CategoryName    string    `json:"category_name,omitempty"`
	Status          string    `json:"status"`
	Stock           int       `json:"stock"`
	Publisher       string    `json:"publisher,omitempty"`
	PublicationDate string    `json:"publication_date,omitempty"`
	Pages           int       `json:"pages,omitempty"`
	Format          string    `json:"format,omitempty"`
	Language        string    `json:"language,omitempty"`
	Dimensions      string    `json:"dimensions,omitempty"`
	Weight          string    `json:"weight,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PriceFormatted returns the price as a human-readable Rupiah string.
func (b *Book) PriceFormatted() string {
	if b.Price <= 0 {
		return "--"
	}
	// Simple Indonesian Rupiah formatting
	return formatRupiah(b.Price)
}

func formatRupiah(amount int) string {
	s := ""
	a := amount
	for a > 0 {
		if s != "" {
			s = "." + s
		}
		rem := a % 1000
		a = a / 1000
		if a > 0 {
			// Pad with leading zeros for middle groups
			if rem < 10 {
				s = "00" + intToStr(rem) + s
			} else if rem < 100 {
				s = "0" + intToStr(rem) + s
			} else {
				s = intToStr(rem) + s
			}
		} else {
			s = intToStr(rem) + s
		}
	}
	if s == "" {
		s = "0"
	}
	return "Rp " + s
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
