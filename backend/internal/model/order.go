// Package model defines domain entities for the Digital Papyrus application.
package model

import "time"

// Order represents a customer order row shown in the admin UI.
type Order struct {
	ID             string    `json:"id"`
	Invoice        string    `json:"invoice"`
	UserID         string    `json:"user_id"`
	UserName       string    `json:"user_name"`
	UserEmail      string    `json:"user_email,omitempty"`
	UserPhone      string    `json:"user_phone,omitempty"`
	UserAddress    string    `json:"user_address,omitempty"`
	Items          []OrderItem `json:"items,omitempty"`
	Notes          string    `json:"notes"`
	TotalQty       int       `json:"total_qty"`
	TotalWeight    int       `json:"total_weight"`
	TotalPrice     int       `json:"total_price"`
	PaymentType    string    `json:"payment_type"`
	Status         string    `json:"status"`
	ShippingName   string    `json:"shipping_name"`
	ShippingService string   `json:"shipping_service"`
	ShippingPrice      int       `json:"shipping_price"`
	ActualShippingCost int       `json:"actual_shipping_cost"`
	Tax                int       `json:"tax"`
	ServiceFee         int       `json:"service_fee"`
	Discount           int       `json:"discount"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// OrderItem represents a line item from order_details joined with item names.
type OrderItem struct {
	ID         string `json:"id"`
	OrderID    string `json:"order_id"`
	ServiceID  string `json:"service_id,omitempty"`
	BookID     string `json:"book_id,omitempty"`
	ItemName   string `json:"item_name"`
	Qty        int    `json:"qty"`
	TotalPrice int    `json:"total_price"`
	UnitCogs   int    `json:"unit_cogs"`
	ItemType   string `json:"item_type"`
	Format     string `json:"format,omitempty"`
}
