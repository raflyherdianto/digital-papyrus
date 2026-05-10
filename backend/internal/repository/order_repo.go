// Package repository provides database access for all domain entities.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

// OrderRepository handles order database operations.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// OrderFilter defines supported list filters for orders.
type OrderFilter struct {
	Page    int
	PerPage int
	Search  string
	Status  string
}

// OrderItemInput describes a single order line item submitted from the UI.
type OrderItemInput struct {
	Title string
	Qty   int
}

// Create inserts a new order row and optional order detail rows.
func (r *OrderRepository) Create(order *model.Order, items []OrderItemInput) error {
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	order.UpdatedAt = order.CreatedAt

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("order_repo: begin create tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO orders (id, invoice, user_id, notes, total_qty, total_weight, total_price, payment_type, status, shipping_name, shipping_service, shipping_price, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.ID, order.Invoice, order.UserID, order.Notes, order.TotalQty, order.TotalWeight, order.TotalPrice,
		order.PaymentType, order.Status, order.ShippingName, order.ShippingService, order.ShippingPrice, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("order_repo: create order: %w", err)
	}

	for idx, item := range items {
		if strings.TrimSpace(item.Title) == "" || item.Qty <= 0 {
			return fmt.Errorf("order_repo: invalid order item at position %d", idx+1)
		}

		bookID, serviceID, unitPrice, err := r.resolveItemByTitle(tx, item.Title)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			`INSERT INTO order_details (id, order_id, service_id, book_id, qty, total_price)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("%s-%03d", order.ID, idx+1), order.ID, serviceID, bookID, item.Qty, unitPrice*item.Qty,
		); err != nil {
			return fmt.Errorf("order_repo: insert order detail: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("order_repo: commit create tx: %w", err)
	}
	return nil
}

func (r *OrderRepository) resolveItemByTitle(tx *sql.Tx, title string) (bookID string, serviceID string, unitPrice int, err error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", "", 0, fmt.Errorf("order_repo: empty item title")
	}

	if err := tx.QueryRow(`SELECT id, price FROM books WHERE LOWER(title) = LOWER(?) LIMIT 1`, trimmed).Scan(&bookID, &unitPrice); err == nil {
		return bookID, "", unitPrice, nil
	} else if err != sql.ErrNoRows {
		return "", "", 0, fmt.Errorf("order_repo: resolve book item %q: %w", trimmed, err)
	}

	if err := tx.QueryRow(`SELECT id, price FROM services WHERE LOWER(title) = LOWER(?) LIMIT 1`, trimmed).Scan(&serviceID, &unitPrice); err == nil {
		return "", serviceID, unitPrice, nil
	} else if err != sql.ErrNoRows {
		return "", "", 0, fmt.Errorf("order_repo: resolve service item %q: %w", trimmed, err)
	}

	return "", "", 0, fmt.Errorf("order_repo: item %q not found in books or services", trimmed)
}

// Delete removes an order and its dependent rows.
func (r *OrderRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("order_repo: begin delete tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM reviews WHERE order_id = ?`, id); err != nil {
		return fmt.Errorf("order_repo: delete reviews: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM order_details WHERE order_id = ?`, id); err != nil {
		return fmt.Errorf("order_repo: delete order_details: %w", err)
	}
	res, err := tx.Exec(`DELETE FROM orders WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("order_repo: delete order: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("order not found")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("order_repo: commit delete tx: %w", err)
	}
	return nil
}

func (r *OrderRepository) FindAll(filter OrderFilter) ([]model.Order, int, error) {
	where := []string{"1=1"}
	args := []any{}

	search := strings.TrimSpace(filter.Search)
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		where = append(where, `(LOWER(o.invoice) LIKE ? OR LOWER(COALESCE(u.name, '')) LIKE ? OR LOWER(COALESCE(o.notes, '')) LIKE ?)`)
		args = append(args, like, like, like)
	}

	status := strings.TrimSpace(strings.ToLower(filter.Status))
	if status != "" {
		where = append(where, "LOWER(o.status) = ?")
		args = append(args, status)
	}

	whereClause := strings.Join(where, " AND ")
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM orders o LEFT JOIN users u ON o.user_id = u.id WHERE %s`, whereClause)

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("order_repo: count orders: %w", err)
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	query := fmt.Sprintf(`SELECT o.id, o.invoice, o.user_id, COALESCE(u.name, '') AS user_name, o.notes, o.total_qty, o.total_weight,
			o.total_price, o.payment_type, o.status, o.shipping_name, o.shipping_service, o.shipping_price, o.created_at, o.updated_at
			FROM orders o LEFT JOIN users u ON o.user_id = u.id WHERE %s ORDER BY o.created_at DESC LIMIT ? OFFSET ?`, whereClause)
	rows, err := r.db.Query(query, append(args, perPage, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("order_repo: query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID, &o.Invoice, &o.UserID, &o.UserName, &o.Notes, &o.TotalQty, &o.TotalWeight,
			&o.TotalPrice, &o.PaymentType, &o.Status, &o.ShippingName, &o.ShippingService, &o.ShippingPrice, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("order_repo: scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("order_repo: rows err: %w", err)
	}

	return orders, total, nil
}

func (r *OrderRepository) FindByID(id string) (*model.Order, error) {
	var o model.Order
	err := r.db.QueryRow(
		`SELECT o.id, o.invoice, o.user_id, COALESCE(u.name, '') AS user_name, o.notes, o.total_qty, o.total_weight,
			o.total_price, o.payment_type, o.status, o.shipping_name, o.shipping_service, o.shipping_price, o.created_at, o.updated_at
		 FROM orders o
		 LEFT JOIN users u ON o.user_id = u.id
		 WHERE o.id = ?`,
		id,
	).Scan(
		&o.ID, &o.Invoice, &o.UserID, &o.UserName, &o.Notes, &o.TotalQty, &o.TotalWeight,
		&o.TotalPrice, &o.PaymentType, &o.Status, &o.ShippingName, &o.ShippingService, &o.ShippingPrice, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("order_repo: find by id: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT od.id, od.order_id, COALESCE(od.service_id, ''), COALESCE(od.book_id, ''),
			COALESCE(s.title, b.title, 'Unknown Item') AS item_name, od.qty, od.total_price
		 FROM order_details od
		 LEFT JOIN services s ON od.service_id = s.id
		 LEFT JOIN books b ON od.book_id = b.id
		 WHERE od.order_id = ?
		 ORDER BY od.id ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("order_repo: query order details: %w", err)
	}
	defer rows.Close()

	o.Items = make([]model.OrderItem, 0)
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ServiceID, &item.BookID, &item.ItemName, &item.Qty, &item.TotalPrice); err != nil {
			return nil, fmt.Errorf("order_repo: scan order detail: %w", err)
		}
		o.Items = append(o.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("order_repo: order details rows err: %w", err)
	}

	return &o, nil
}
