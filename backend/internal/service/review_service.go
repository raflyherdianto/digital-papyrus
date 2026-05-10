package service

import (
	"errors"
	"time"

	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/google/uuid"
)

type ReviewService struct {
	repo *repository.ReviewRepository
}

func NewReviewService(repo *repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) CreateReview(userID, orderID string, serviceIDs []string, bookIDs []string, details map[string]string, rating map[string]int) (*model.Review, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if orderID == "" {
		return nil, errors.New("order_id is required")
	}

	if err := s.repo.CheckOrderCompleted(userID, orderID); err != nil {
		return nil, err
	}

	if serviceIDs == nil {
		serviceIDs = []string{}
	}
	if bookIDs == nil {
		bookIDs = []string{}
	}
	if details == nil {
		details = make(map[string]string)
	}
	if rating == nil {
		rating = make(map[string]int)
	}

	now := time.Now()
	rev := &model.Review{
		ID:        uuid.New().String(),
		UserID:    userID,
		OrderID:   orderID,
		ServiceID: serviceIDs,
		BookID:    bookIDs,
		Details:   details,
		Rating:    rating,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(rev); err != nil {
		return nil, err
	}

	return rev, nil
}

func (s *ReviewService) GetReview(id string) (*model.Review, error) {
	return s.repo.GetByID(id)
}

func (s *ReviewService) GetAllReviews(page, limit int) ([]*model.Review, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return s.repo.GetAll(page, limit)
}

func (s *ReviewService) UpdateReview(id string, serviceIDs []string, bookIDs []string, details map[string]string, rating map[string]int) (*model.Review, error) {
	rev, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if serviceIDs != nil {
		rev.ServiceID = serviceIDs
	}
	if bookIDs != nil {
		rev.BookID = bookIDs
	}
	if details != nil {
		rev.Details = details
	}
	if rating != nil {
		rev.Rating = rating
	}

	if err := s.repo.Update(rev); err != nil {
		return nil, err
	}

	return rev, nil
}

func (s *ReviewService) DeleteReview(id string) error {
	return s.repo.Delete(id)
}
