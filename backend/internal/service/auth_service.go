// Package service implements business logic for the Digital Papyrus application.
package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/repository"
)

var (
	otpStore = sync.Map{} // email -> otpInfo{code, expiresAt}
)

type otpInfo struct {
	code      string
	expiresAt time.Time
}

// Common errors for auth operations.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
)

// JWTClaims defines the structure of our JWT token payload.
type JWTClaims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

// LoginInput represents the request body for login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginOutput represents the response body for a successful login.
type LoginOutput struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      *model.User `json:"user"`
}

// RegisterInput represents the request body for user registration.
type RegisterInput struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	Province    string `json:"province"`
	City        string `json:"city"`
	ZipCode     string `json:"zip_code"`
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(input LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("auth_service: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountDisabled
	}

	expiresAt := time.Now().Add(s.cfg.JWT.ExpiryTime)
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.cfg.App.Name,
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("auth_service: sign token: %w", err)
	}

	return &LoginOutput{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// Register creates a new customer account and returns login info.
func (s *AuthService) Register(input RegisterInput) (*LoginOutput, error) {
	// Check if email already exists
	existing, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("auth_service: register check email: %w", err)
	}
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth_service: register hash password: %w", err)
	}

	// Create user object
	now := time.Now().UTC()
	user := &model.User{
		ID:           uuid.New().String(),
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
		Role:         model.RoleCustomer,
		IsActive:     true,
		PhoneNumber:  input.PhoneNumber,
		Address:      input.Address,
		Province:     input.Province,
		City:         input.City,
		ZipCode:      input.ZipCode,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("auth_service: register create: %w", err)
	}

	// Automatic login after successful registration
	return s.Login(LoginInput{
		Email:    input.Email,
		Password: input.Password,
	})
}

// ValidateToken parses and validates a JWT token string.
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// GetCurrentUser retrieves the full user by ID (for /auth/me).
func (s *AuthService) GetCurrentUser(userID string) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("auth_service: get current user: %w", err)
	}
	return user, nil
}

// SendOTP generates and "sends" a 6-digit OTP to the email.
func (s *AuthService) SendOTP(email string) error {
	// Generate random 6-digit code
	code := fmt.Sprintf("%06d", uuid.New().ID()%1000000)
	
	// Store in memory with 5 minute expiry
	otpStore.Store(email, otpInfo{
		code:      code,
		expiresAt: time.Now().Add(5 * time.Minute),
	})

	// Simulate sending email by logging to console
	fmt.Printf("\n[EMAIL SIMULATOR] To: %s | OTP Code: %s | Expires in 5 minutes\n\n", email, code)
	
	return nil
}

// VerifyOTP checks if the provided code matches the one stored for the email.
func (s *AuthService) VerifyOTP(email, code string) (bool, error) {
	val, ok := otpStore.Load(email)
	if !ok {
		return false, errors.New("OTP not found or expired")
	}

	info := val.(otpInfo)
	if time.Now().After(info.expiresAt) {
		otpStore.Delete(email)
		return false, errors.New("OTP has expired")
	}

	if info.code != code {
		return false, errors.New("invalid OTP code")
	}

	// Success! Delete the OTP so it can't be reused
	otpStore.Delete(email)
	return true, nil
}
