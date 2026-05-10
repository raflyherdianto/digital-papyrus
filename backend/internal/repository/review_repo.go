package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create inserts a new review.
func (r *ReviewRepository) Create(rev *model.Review) error {
	serviceIDJSON, err := json.Marshal(rev.ServiceID)
	if err != nil {
		return err
	}
	bookIDJSON, err := json.Marshal(rev.BookID)
	if err != nil {
		return err
	}
	detailsJSON, err := json.Marshal(rev.Details)
	if err != nil {
		return err
	}
	ratingJSON, err := json.Marshal(rev.Rating)
	if err != nil {
		return err
	}

	query := `INSERT INTO reviews (id, user_id, order_id, service_id, book_id, details, rating, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, rev.ID, rev.UserID, rev.OrderID, string(serviceIDJSON), string(bookIDJSON), string(detailsJSON), string(ratingJSON), rev.CreatedAt, rev.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository: failed to insert review: %w", err)
	}

	return nil
}

// GetByID retrieves a single review.
func (r *ReviewRepository) GetByID(id string) (*model.Review, error) {
	query := `SELECT r.id, r.user_id, r.order_id, COALESCE(o.invoice, ''), r.service_id, r.book_id, r.details, r.rating, r.created_at, r.updated_at
			  FROM reviews r 
			  LEFT JOIN orders o ON r.order_id = o.id
			  WHERE r.id = ?`
	row := r.db.QueryRow(query, id)

	var rev model.Review
	var serviceIDStr, bookIDStr, detailsStr, ratingStr string

	err := row.Scan(&rev.ID, &rev.UserID, &rev.OrderID, &rev.Invoice, &serviceIDStr, &bookIDStr, &detailsStr, &ratingStr, &rev.CreatedAt, &rev.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("review not found")
		}
		return nil, err
	}

	json.Unmarshal([]byte(serviceIDStr), &rev.ServiceID)
	json.Unmarshal([]byte(bookIDStr), &rev.BookID)
	json.Unmarshal([]byte(detailsStr), &rev.Details)
	json.Unmarshal([]byte(ratingStr), &rev.Rating)

	return &rev, nil
}

// GetAll retrieves all reviews, optionally paginated.
func (r *ReviewRepository) GetAll(page, limit int) ([]*model.Review, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM reviews`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `SELECT r.id, r.user_id, r.order_id, COALESCE(o.invoice, ''), r.service_id, r.book_id, r.details, r.rating, r.created_at, r.updated_at
			  FROM reviews r 
			  LEFT JOIN orders o ON r.order_id = o.id
			  ORDER BY r.created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []*model.Review
	for rows.Next() {
		var rev model.Review
		var serviceIDStr, bookIDStr, detailsStr, ratingStr string

		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.OrderID, &rev.Invoice, &serviceIDStr, &bookIDStr, &detailsStr, &ratingStr, &rev.CreatedAt, &rev.UpdatedAt); err != nil {
			return nil, 0, err
		}

		json.Unmarshal([]byte(serviceIDStr), &rev.ServiceID)
		json.Unmarshal([]byte(bookIDStr), &rev.BookID)
		json.Unmarshal([]byte(detailsStr), &rev.Details)
		json.Unmarshal([]byte(ratingStr), &rev.Rating)

		reviews = append(reviews, &rev)
	}

	return reviews, total, nil
}

// Update modifies an existing review.
func (r *ReviewRepository) Update(rev *model.Review) error {
	serviceIDJSON, _ := json.Marshal(rev.ServiceID)
	bookIDJSON, _ := json.Marshal(rev.BookID)
	detailsJSON, _ := json.Marshal(rev.Details)
	ratingJSON, _ := json.Marshal(rev.Rating)

	rev.UpdatedAt = time.Now()

	query := `UPDATE reviews SET service_id = ?, book_id = ?, details = ?, rating = ?, updated_at = ?
			  WHERE id = ?`
	res, err := r.db.Exec(query, string(serviceIDJSON), string(bookIDJSON), string(detailsJSON), string(ratingJSON), rev.UpdatedAt, rev.ID)
	if err != nil {
		return fmt.Errorf("repository: failed to update review: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("review not found")
	}

	return nil
}

// Delete removes a review by ID.
func (r *ReviewRepository) Delete(id string) error {
	query := `DELETE FROM reviews WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("review not found")
	}

	return nil
}

// CheckOrderCompleted verifies if an order exists for the user and is delivered/completed.
func (r *ReviewRepository) CheckOrderCompleted(userID, orderID string) error {
	var status string
	err := r.db.QueryRow(`SELECT status FROM orders WHERE id = ? AND user_id = ?`, orderID, userID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("order not found or does not belong to user")
		}
		return err
	}

	if status != "completed" && status != "delivered" {
		return errors.New("order is not completed yet")
	}

	return nil
}
