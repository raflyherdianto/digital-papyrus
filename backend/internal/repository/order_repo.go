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
	UserID  string
}

// OrderItemInput describes a single order line item submitted from the UI.
type OrderItemInput struct {
	ID    string
	Type  string
	Title string
	Qty   int
	Price int
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
		`INSERT INTO orders (id, invoice, user_id, notes, total_qty, total_weight, total_price, payment_type, status, shipping_name, shipping_service, shipping_price, tax, service_fee, discount, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		order.ID, order.Invoice, order.UserID, order.Notes, order.TotalQty, order.TotalWeight, order.TotalPrice,
		order.PaymentType, order.Status, order.ShippingName, order.ShippingService, order.ShippingPrice, 
		order.Tax, order.ServiceFee, order.Discount, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("order_repo: create order: %w", err)
	}

	for idx, item := range items {
		if strings.TrimSpace(item.Title) == "" || item.Qty <= 0 {
			return fmt.Errorf("order_repo: invalid order item at position %d", idx+1)
		}

		var bookID, serviceID string
		var unitPrice int
		var err error

		if item.ID != "" && item.Type != "" {
			if strings.ToLower(item.Type) == "service" {
				serviceID = item.ID
				err = tx.QueryRow(`SELECT price FROM services WHERE id = $1 LIMIT 1`, serviceID).Scan(&unitPrice)
				if err != nil {
					bookID, serviceID, unitPrice, err = r.resolveItemByTitle(tx, item.Title)
				}
			} else {
				bookID = item.ID
				err = tx.QueryRow(`SELECT price FROM books WHERE id = $1 LIMIT 1`, bookID).Scan(&unitPrice)
				if err != nil {
					bookID, serviceID, unitPrice, err = r.resolveItemByTitle(tx, item.Title)
				}
			}
		} else {
			bookID, serviceID, unitPrice, err = r.resolveItemByTitle(tx, item.Title)
		}

		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			`INSERT INTO order_details (id, order_id, service_id, book_id, qty, total_price)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
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

	if err := tx.QueryRow(`SELECT id, price FROM books WHERE LOWER(title) = LOWER($1) LIMIT 1`, trimmed).Scan(&bookID, &unitPrice); err == nil {
		return bookID, "", unitPrice, nil
	} else if err != sql.ErrNoRows {
		return "", "", 0, fmt.Errorf("order_repo: resolve book item %q: %w", trimmed, err)
	}

	if err := tx.QueryRow(`SELECT id, price FROM services WHERE LOWER(title) = LOWER($1) LIMIT 1`, trimmed).Scan(&serviceID, &unitPrice); err == nil {
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

	if _, err := tx.Exec(`DELETE FROM reviews WHERE order_id = $1`, id); err != nil {
		return fmt.Errorf("order_repo: delete reviews: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM order_details WHERE order_id = $1`, id); err != nil {
		return fmt.Errorf("order_repo: delete order_details: %w", err)
	}
	res, err := tx.Exec(`DELETE FROM orders WHERE id = $1`, id)
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
	argIdx := 1

	if filter.UserID != "" {
		where = append(where, fmt.Sprintf("o.user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}

	search := strings.TrimSpace(filter.Search)
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		where = append(where, fmt.Sprintf(`(LOWER(o.invoice) LIKE $%d OR LOWER(COALESCE(u.name, '')) LIKE $%d OR LOWER(COALESCE(o.notes, '')) LIKE $%d)`, argIdx, argIdx, argIdx))
		args = append(args, like)
		argIdx++
	}

	status := strings.TrimSpace(strings.ToLower(filter.Status))
	if status != "" {
		where = append(where, fmt.Sprintf("LOWER(o.status) = $%d", argIdx))
		args = append(args, status)
		argIdx++
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
			o.total_price, o.payment_type, o.status, o.shipping_name, o.shipping_service, o.shipping_price, COALESCE(o.actual_shipping_cost, o.shipping_price) AS actual_shipping_cost,
			o.tax, o.service_fee, o.discount, o.created_at, o.updated_at
			FROM orders o LEFT JOIN users u ON o.user_id = u.id WHERE %s ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
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
			&o.TotalPrice, &o.PaymentType, &o.Status, &o.ShippingName, &o.ShippingService, &o.ShippingPrice, &o.ActualShippingCost,
			&o.Tax, &o.ServiceFee, &o.Discount, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("order_repo: scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("order_repo: rows err: %w", err)
	}

	if len(orders) > 0 {
		var orderIDs []string
		for _, o := range orders {
			orderIDs = append(orderIDs, o.ID)
		}

		placeholders := make([]string, len(orderIDs))
		argsQuery := make([]any, len(orderIDs))
		for i, id := range orderIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			argsQuery[i] = id
		}

		detailsQuery := fmt.Sprintf(`
			SELECT od.id, od.order_id, COALESCE(od.service_id, ''), COALESCE(od.book_id, ''),
				COALESCE(s.title, b.title, 'Unknown Item') AS item_name, od.qty, od.total_price, COALESCE(od.unit_cogs, 0) AS unit_cogs,
				CASE WHEN od.book_id IS NOT NULL AND od.book_id != '' THEN 'book' ELSE 'service' END AS item_type,
				COALESCE(b.format, '') AS format
			FROM order_details od
			LEFT JOIN services s ON od.service_id = s.id
			LEFT JOIN books b ON od.book_id = b.id
			WHERE od.order_id IN (%s)
		`, strings.Join(placeholders, ","))

		dRows, err := r.db.Query(detailsQuery, argsQuery...)
		if err != nil {
			return nil, 0, fmt.Errorf("order_repo: query order details: %w", err)
		}
		defer dRows.Close()

		detailsMap := make(map[string][]model.OrderItem)
		for dRows.Next() {
			var d model.OrderItem
			if err := dRows.Scan(&d.ID, &d.OrderID, &d.ServiceID, &d.BookID, &d.ItemName, &d.Qty, &d.TotalPrice, &d.UnitCogs, &d.ItemType, &d.Format); err != nil {
				return nil, 0, fmt.Errorf("order_repo: scan order detail: %w", err)
			}
			detailsMap[d.OrderID] = append(detailsMap[d.OrderID], d)
		}

		for i, o := range orders {
			if items, ok := detailsMap[o.ID]; ok {
				orders[i].Items = items
			} else {
				orders[i].Items = []model.OrderItem{}
			}
		}
	}

	return orders, total, nil
}

func (r *OrderRepository) FindByID(id string) (*model.Order, error) {
	return r.FindByIDOrInvoice(id)
}

func (r *OrderRepository) FindByIDOrInvoice(idOrInvoice string) (*model.Order, error) {
	var o model.Order
	err := r.db.QueryRow(
		`SELECT o.id, o.invoice, o.user_id, COALESCE(u.name, '') AS user_name, COALESCE(u.email, '') AS user_email,
			COALESCE(u.phone_number, '') AS user_phone, COALESCE(u.address, '') AS user_address,
			o.notes, o.total_qty, o.total_weight,
			o.total_price, o.payment_type, o.status, o.shipping_name, o.shipping_service, o.shipping_price, COALESCE(o.actual_shipping_cost, o.shipping_price) AS actual_shipping_cost,
			o.tax, o.service_fee, o.discount, o.created_at, o.updated_at
		 FROM orders o
		 LEFT JOIN users u ON o.user_id = u.id
		 WHERE o.id = $1 OR o.invoice = $1`,
		idOrInvoice,
	).Scan(
		&o.ID, &o.Invoice, &o.UserID, &o.UserName, &o.UserEmail, &o.UserPhone, &o.UserAddress,
		&o.Notes, &o.TotalQty, &o.TotalWeight,
		&o.TotalPrice, &o.PaymentType, &o.Status, &o.ShippingName, &o.ShippingService, &o.ShippingPrice, &o.ActualShippingCost,
		&o.Tax, &o.ServiceFee, &o.Discount, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("order_repo: find by id or invoice: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT od.id, od.order_id, COALESCE(od.service_id, ''), COALESCE(od.book_id, ''),
			COALESCE(s.title, b.title, 'Unknown Item') AS item_name, od.qty, od.total_price, COALESCE(od.unit_cogs, 0) AS unit_cogs,
			CASE WHEN od.book_id IS NOT NULL AND od.book_id != '' THEN 'book' ELSE 'service' END AS item_type,
			COALESCE(b.format, '') AS format
		 FROM order_details od
		 LEFT JOIN services s ON od.service_id = s.id
		 LEFT JOIN books b ON od.book_id = b.id
		 WHERE od.order_id = $1
		 ORDER BY od.id ASC`,
		o.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("order_repo: query order details: %w", err)
	}
	defer rows.Close()

	o.Items = make([]model.OrderItem, 0)
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ServiceID, &item.BookID, &item.ItemName, &item.Qty, &item.TotalPrice, &item.UnitCogs, &item.ItemType, &item.Format); err != nil {
			return nil, fmt.Errorf("order_repo: scan order detail: %w", err)
		}
		o.Items = append(o.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("order_repo: order details rows err: %w", err)
	}

	return &o, nil
}

// UpdateStatus updates the status of an order.
func (r *OrderRepository) UpdateStatus(id string, status string) error {
	res, err := r.db.Exec(`UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("order_repo: update status: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("order not found")
	}
	return nil
}

// FindExpiredOrders returns orders that are in pending status and created before cutoff.
func (r *OrderRepository) FindExpiredOrders(cutoff time.Time) ([]model.Order, error) {
	query := `SELECT id, invoice, user_id, status, total_price, created_at FROM orders 
	          WHERE LOWER(status) IN ('pending', 'waiting_confirmation') 
	            AND created_at < $1`
	rows, err := r.db.Query(query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("order_repo: query expired orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.Invoice, &o.UserID, &o.Status, &o.TotalPrice, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("order_repo: scan expired order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
