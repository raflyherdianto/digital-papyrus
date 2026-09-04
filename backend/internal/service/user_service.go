package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.userRepo.FindAll()
}

func (s *UserService) GetUserByID(id string) (*model.User, error) {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("user_service: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

type CreateUserInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Regency     string `json:"regency"`
	District    string `json:"district"`
	Village     string `json:"village"`
	ZipCode     string `json:"zip_code"`
}

func (s *UserService) CreateUser(input CreateUserInput) (*model.User, error) {
	if input.Regency == "" && input.District != "" {
		input.Regency = input.District
	}
	existing, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("user_service: check email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("user_service: hash password: %w", err)
	}

	now := time.Now().UTC()
	u := &model.User{
		ID:           uuid.New().String(),
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
		Role:         input.Role,
		IsActive:     input.IsActive,
		PhoneNumber:  input.PhoneNumber,
		Address:      input.Address,
		Province:     input.Province,
		City:         input.City,
		Regency:      input.Regency,
		Village:      input.Village,
		ZipCode:      input.ZipCode,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(u); err != nil {
		return nil, fmt.Errorf("user_service: create: %w", err)
	}

	return u, nil
}

type UpdateUserInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Regency     string `json:"regency"`
	District    string `json:"district"`
	Village     string `json:"village"`
	ZipCode     string `json:"zip_code"`
}

func (s *UserService) UpdateUser(id string, input UpdateUserInput) (*model.User, error) {
	if input.Regency == "" && input.District != "" {
		input.Regency = input.District
	}
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("user_service: %w", err)
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	if input.Email != u.Email {
		existing, err := s.userRepo.FindByEmail(input.Email)
		if err != nil {
			return nil, fmt.Errorf("user_service: check email: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, ErrEmailAlreadyExists
		}
		u.Email = input.Email
	}

	u.Name = input.Name
	u.Role = input.Role
	u.IsActive = input.IsActive
	u.PhoneNumber = input.PhoneNumber
	u.Address = input.Address
	u.Province = input.Province
	u.City = input.City
	u.Regency = input.Regency
	u.Village = input.Village
	u.ZipCode = input.ZipCode

	if err := s.userRepo.Update(u); err != nil {
		return nil, fmt.Errorf("user_service: update: %w", err)
	}

	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("user_service: hash password: %w", err)
		}
		if err := s.userRepo.UpdatePassword(id, string(hash)); err != nil {
			return nil, fmt.Errorf("user_service: update password: %w", err)
		}
		u.PasswordHash = string(hash)
	}

	return u, nil
}

func (s *UserService) DeleteUser(id string) error {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("user_service: check id: %w", err)
	}
	if u == nil {
		return ErrUserNotFound
	}
	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("user_service: delete: %w", err)
	}
	return nil
}
