// Package service implements business logic for the Digital Papyrus application.
package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
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
	otpStore        = sync.Map{} // email -> otpInfo{code, expiresAt}
	resetTokenStore = sync.Map{} // token -> resetInfo{email, expiresAt}
)

type otpInfo struct {
	code      string
	expiresAt time.Time
}

type resetInfo struct {
	email     string
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

// UpdateProfile updates the profile of the current authenticated user.
func (s *AuthService) UpdateProfile(userID string, name, phoneNumber, address, province, city, regency, village, zipCode string) (*model.User, error) {
	u, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("auth_service: find user: %w", err)
	}
	if u == nil {
		return nil, errors.New("user not found")
	}

	u.Name = name
	u.PhoneNumber = phoneNumber
	u.Address = address
	u.Province = province
	u.City = city
	u.Regency = regency
	u.Village = village
	u.ZipCode = zipCode

	if err := s.userRepo.Update(u); err != nil {
		return nil, fmt.Errorf("auth_service: update user: %w", err)
	}

	return u, nil
}

// SendOTP generates and sends a 6-digit OTP to the email.
func (s *AuthService) SendOTP(email string) error {
	// Generate random 6-digit code
	code := fmt.Sprintf("%06d", uuid.New().ID()%1000000)
	
	// Store in database verify_otps table with 5 minute expiry
	expiredAt := time.Now().Add(5 * time.Minute)
	if err := s.userRepo.SaveOTP(email, code, expiredAt); err != nil {
		return fmt.Errorf("failed to save OTP: %w", err)
	}
	otpStore.Store(email, otpInfo{
		code:      code,
		expiresAt: expiredAt,
	})

	// Find logo file synchronously
	var logoBytes []byte
	logoPath := ""
	logoPaths := []string{
		"logo.png",
		"../logo.png",
		"./logo.png",
		"backend/logo.png",
		"frontend/public/logo.png",
		"../frontend/public/logo.png",
	}
	for _, p := range logoPaths {
		if _, err := os.Stat(p); err == nil {
			logoPath = p
			break
		}
	}
	if logoPath != "" {
		logoBytes, _ = os.ReadFile(logoPath)
	}

	// Render the template
	body := strings.Replace(OTPEmailTemplate, "184920", code, 1)
	useCID := len(logoBytes) > 0
	if useCID {
		body = strings.Replace(body, "https://digitalpapyrus.web.id/logo.png", "cid:logo_cid", 1)
	}

	// SMTP credentials and server info
	smtpHost := s.cfg.SMTP.Host
	smtpPort := s.cfg.SMTP.Port
	smtpUsername := s.cfg.SMTP.Username
	smtpPassword := s.cfg.SMTP.Password

	// Send email in a goroutine so it doesn't block the API client
	go func() {
		// Determine the sender domain for Message-ID alignment
		domain := "gmail.com"
		parts := strings.Split(smtpUsername, "@")
		if len(parts) > 1 {
			domain = parts[1]
		}
		messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
		dateStr := time.Now().Format(time.RFC1123Z)

		var msgBuilder strings.Builder
		msgBuilder.WriteString(fmt.Sprintf("From: Digital Papyrus <%s>\r\n", smtpUsername))
		msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", email))
		msgBuilder.WriteString("Subject: Verifikasi Akun Baru Anda\r\n")
		msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", dateStr))
		msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
		msgBuilder.WriteString("MIME-Version: 1.0\r\n")
		msgBuilder.WriteString("Auto-Submitted: auto-generated\r\n")
		msgBuilder.WriteString("X-Auto-Response-Suppress: OOF, AutoReply\r\n")

		if useCID {
			// Boundary for multipart/related
			boundary := "----=_NextPart_" + uuid.New().String()
			msgBuilder.WriteString(fmt.Sprintf("Content-Type: multipart/related; type=\"text/html\"; boundary=\"%s\"\r\n\r\n", boundary))

			// HTML Body part
			msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
			msgBuilder.WriteString(body)
			msgBuilder.WriteString("\r\n")

			// Image Attachment part (inline CID)
			msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msgBuilder.WriteString("Content-Type: image/png\r\n")
			msgBuilder.WriteString("Content-Disposition: inline\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: base64\r\n")
			msgBuilder.WriteString("Content-ID: <logo_cid>\r\n\r\n")

			// Base64 encode in standard RFC MIME chunks (76 chars per line)
			b64Bytes := make([]byte, base64.StdEncoding.EncodedLen(len(logoBytes)))
			base64.StdEncoding.Encode(b64Bytes, logoBytes)
			
			for i := 0; i < len(b64Bytes); i += 76 {
				end := i + 76
				if end > len(b64Bytes) {
					end = len(b64Bytes)
				}
				msgBuilder.Write(b64Bytes[i:end])
				msgBuilder.WriteString("\r\n")
			}

			// Close boundary
			msgBuilder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
		} else {
			// Simple non-multipart HTML email
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
			msgBuilder.WriteString(body)
		}

		auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		err := smtp.SendMail(addr, auth, smtpUsername, []string{email}, []byte(msgBuilder.String()))
		if err != nil {
			log.Printf("[SMTP ERROR] Failed to send OTP to %s: %v", email, err)
		} else {
			log.Printf("[SMTP SUCCESS] Sent OTP to %s (Code: %s)", email, code)
		}
	}()
	
	return nil
}

// VerifyOTP checks if the provided code matches the one stored for the email in verify_otps table.
// If valid and matches, it deletes the entry from verify_otps.
func (s *AuthService) VerifyOTP(email, code string) (bool, error) {
	ok, err := s.userRepo.VerifyAndConsumeOTP(email, code)
	if err != nil {
		otpStore.Delete(email)
		return false, err
	}

	otpStore.Delete(email)
	return ok, nil
}

// RequestPasswordReset generates a reset token, saves it, and sends the email link.
func (s *AuthService) RequestPasswordReset(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("auth_service: find user: %w", err)
	}
	if user == nil {
		return errors.New("email not found")
	}

	token := uuid.New().String()
	expiredAt := time.Now().Add(15 * time.Minute)

	// Store in database verify_otps table
	if err := s.userRepo.SaveOTP(email, token, expiredAt); err != nil {
		return fmt.Errorf("auth_service: save reset token to verify_otps: %w", err)
	}

	resetTokenStore.Store(token, resetInfo{
		email:     email,
		expiresAt: expiredAt,
	})

	frontendURL := "http://localhost:4321"
	if s.cfg.IsProduction() {
		frontendURL = "https://digitalpapyrus.web.id"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s", frontendURL, token, email)

	// Find logo file synchronously
	var logoBytes []byte
	logoPath := ""
	logoPaths := []string{
		"logo.png",
		"../logo.png",
		"./logo.png",
		"backend/logo.png",
		"frontend/public/logo.png",
		"../frontend/public/logo.png",
	}
	for _, p := range logoPaths {
		if _, err := os.Stat(p); err == nil {
			logoPath = p
			break
		}
	}
	if logoPath != "" {
		logoBytes, _ = os.ReadFile(logoPath)
	}

	// Render body template
	body := strings.Replace(ResetPasswordEmailTemplate, "RESET_LINK_PLACEHOLDER", resetLink, 1)
	useCID := len(logoBytes) > 0
	if useCID {
		body = strings.Replace(body, "https://digitalpapyrus.web.id/logo.png", "cid:logo_cid", 1)
	}

	smtpHost := s.cfg.SMTP.Host
	smtpPort := s.cfg.SMTP.Port
	smtpUsername := s.cfg.SMTP.Username
	smtpPassword := s.cfg.SMTP.Password

	go func() {
		// Determine the sender domain for Message-ID alignment
		domain := "gmail.com"
		parts := strings.Split(smtpUsername, "@")
		if len(parts) > 1 {
			domain = parts[1]
		}
		messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
		dateStr := time.Now().Format(time.RFC1123Z)

		var msgBuilder strings.Builder
		msgBuilder.WriteString(fmt.Sprintf("From: Digital Papyrus <%s>\r\n", smtpUsername))
		msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", email))
		msgBuilder.WriteString("Subject: Atur Ulang Kata Sandi Anda\r\n")
		msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", dateStr))
		msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
		msgBuilder.WriteString("MIME-Version: 1.0\r\n")
		msgBuilder.WriteString("Auto-Submitted: auto-generated\r\n")
		msgBuilder.WriteString("X-Auto-Response-Suppress: OOF, AutoReply\r\n")

		if useCID {
			boundary := "----=_NextPart_" + uuid.New().String()
			msgBuilder.WriteString(fmt.Sprintf("Content-Type: multipart/related; type=\"text/html\"; boundary=\"%s\"\r\n\r\n", boundary))

			// HTML Body part
			msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
			msgBuilder.WriteString(body)
			msgBuilder.WriteString("\r\n")

			// Image Attachment part (inline CID)
			msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msgBuilder.WriteString("Content-Type: image/png\r\n")
			msgBuilder.WriteString("Content-Disposition: inline\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: base64\r\n")
			msgBuilder.WriteString("Content-ID: <logo_cid>\r\n\r\n")

			b64Bytes := make([]byte, base64.StdEncoding.EncodedLen(len(logoBytes)))
			base64.StdEncoding.Encode(b64Bytes, logoBytes)
			
			for i := 0; i < len(b64Bytes); i += 76 {
				end := i + 76
				if end > len(b64Bytes) {
					end = len(b64Bytes)
				}
				msgBuilder.Write(b64Bytes[i:end])
				msgBuilder.WriteString("\r\n")
			}

			msgBuilder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
		} else {
			// Simple non-multipart HTML email
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
			msgBuilder.WriteString(body)
		}

		auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		err := smtp.SendMail(addr, auth, smtpUsername, []string{email}, []byte(msgBuilder.String()))
		if err != nil {
			log.Printf("[SMTP ERROR] Failed to send password reset to %s: %v", email, err)
		} else {
			log.Printf("[SMTP SUCCESS] Sent password reset to %s (Link: %s)", email, resetLink)
		}
	}()

	return nil
}

// ResetPassword validates the token and updates the user's password.
func (s *AuthService) ResetPassword(email, token, newPassword string) error {
	// First verify and consume from verify_otps database table
	_, dbErr := s.userRepo.VerifyAndConsumeOTP(email, token)
	if dbErr != nil {
		// Fallback check memory store
		val, ok := resetTokenStore.Load(token)
		if !ok {
			return errors.New("token invalid or expired")
		}

		info := val.(resetInfo)
		if time.Now().After(info.expiresAt) {
			resetTokenStore.Delete(token)
			return errors.New("token has expired")
		}

		if info.email != email {
			return errors.New("token does not belong to this email")
		}
	}

	resetTokenStore.Delete(token)

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("auth_service: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.Security.BcryptCost)
	if err != nil {
		return fmt.Errorf("auth_service: hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(user.ID, string(hash)); err != nil {
		return fmt.Errorf("auth_service: update password: %w", err)
	}

	resetTokenStore.Delete(token)

	// Send password change email notification
	_ = s.SendPasswordChangedNotification(email)

	return nil
}

// SendPasswordChangedNotification sends a security alert to the user's email.
func (s *AuthService) SendPasswordChangedNotification(email string) error {
	// Find logo file synchronously
	var logoBytes []byte
	logoPath := ""
	logoPaths := []string{
		"logo.png",
		"../logo.png",
		"./logo.png",
		"backend/logo.png",
		"frontend/public/logo.png",
		"../frontend/public/logo.png",
	}
	for _, p := range logoPaths {
		if _, err := os.Stat(p); err == nil {
			logoPath = p
			break
		}
	}
	if logoPath != "" {
		logoBytes, _ = os.ReadFile(logoPath)
	}

	// Format time in WIB (Western Indonesian Time) which is UTC+7
	loc := time.FixedZone("WIB", 7*60*60)
	timeStr := time.Now().In(loc).Format("02 Jan 2006 15:04:05")

	// Render the template
	body := strings.Replace(PasswordChangedEmailTemplate, "DATE_PLACEHOLDER", timeStr, 1)
	body = strings.Replace(body, "EMAIL_PLACEHOLDER", email, 1)
	
	useCID := len(logoBytes) > 0
	if useCID {
		body = strings.Replace(body, "https://digitalpapyrus.web.id/logo.png", "cid:logo_cid", 1)
	}

	smtpHost := s.cfg.SMTP.Host
	smtpPort := s.cfg.SMTP.Port
	smtpUsername := s.cfg.SMTP.Username
	smtpPassword := s.cfg.SMTP.Password

	go func() {
		// Determine the sender domain for Message-ID alignment
		domain := "gmail.com"
		parts := strings.Split(smtpUsername, "@")
		if len(parts) > 1 {
			domain = parts[1]
		}
		messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
		dateStr := time.Now().Format(time.RFC1123Z)

		var msgBuilder strings.Builder
		msgBuilder.WriteString(fmt.Sprintf("From: Digital Papyrus <%s>\r\n", smtpUsername))
		msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", email))
		msgBuilder.WriteString("Subject: Kata Sandi Anda Berhasil Diperbarui\r\n")
		msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", dateStr))
		msgBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
		msgBuilder.WriteString("MIME-Version: 1.0\r\n")
		msgBuilder.WriteString("Auto-Submitted: auto-generated\r\n")
		msgBuilder.WriteString("X-Auto-Response-Suppress: OOF, AutoReply\r\n")

		if useCID {
			boundary := "----=_NextPart_" + uuid.New().String()
			msgBuilder.WriteString(fmt.Sprintf("Content-Type: multipart/related; type=\"text/html\"; boundary=\"%s\"\r\n\r\n", boundary))

			// HTML Body part
			msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
			msgBuilder.WriteString(body)
			msgBuilder.WriteString("\r\n")

			// Image Attachment part (inline CID)
			msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msgBuilder.WriteString("Content-Type: image/png\r\n")
			msgBuilder.WriteString("Content-Disposition: inline\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: base64\r\n")
			msgBuilder.WriteString("Content-ID: <logo_cid>\r\n\r\n")

			b64Bytes := make([]byte, base64.StdEncoding.EncodedLen(len(logoBytes)))
			base64.StdEncoding.Encode(b64Bytes, logoBytes)
			
			for i := 0; i < len(b64Bytes); i += 76 {
				end := i + 76
				if end > len(b64Bytes) {
					end = len(b64Bytes)
				}
				msgBuilder.Write(b64Bytes[i:end])
				msgBuilder.WriteString("\r\n")
			}

			msgBuilder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
		} else {
			// Simple non-multipart HTML email
			msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			msgBuilder.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
			msgBuilder.WriteString(body)
		}

		auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		err := smtp.SendMail(addr, auth, smtpUsername, []string{email}, []byte(msgBuilder.String()))
		if err != nil {
			log.Printf("[SMTP ERROR] Failed to send password change notification to %s: %v", email, err)
		} else {
			log.Printf("[SMTP SUCCESS] Sent password change notification to %s", email)
		}
	}()

	return nil
}

