package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/middleware"
	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
	"github.com/digitalpapyrus/backend/pkg/validator"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	// Validate email
	input.Email = validator.SanitizeString(input.Email)
	if !validator.ValidateEmail(input.Email) {
		response.BadRequest(c, "Validation failed", map[string]string{
			"email": "valid email address is required",
		})
		return
	}

	if input.Password == "" {
		response.BadRequest(c, "Validation failed", map[string]string{
			"password": "password is required",
		})
		return
	}

	result, err := h.authService.Login(input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		if errors.Is(err, service.ErrAccountDisabled) {
			response.Forbidden(c, "Your account has been disabled. Contact support.")
			return
		}
		response.InternalError(c, "An error occurred during authentication")
		return
	}

	response.OK(c, "Login successful", result)
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body", nil)
		return
	}

	result, err := h.authService.Register(input)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Conflict(c, "Email already registered")
			return
		}
		response.InternalError(c, "An error occurred during registration")
		return
	}

	response.Created(c, "Registration successful", result)
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, err := h.authService.GetCurrentUser(userID.(string))
	if err != nil {
		response.InternalError(c, "Failed to retrieve user information")
		return
	}
	if user == nil {
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, "User retrieved successfully", user)
}

// Logout handles POST /api/v1/auth/logout
// For stateless JWT, logout is handled client-side by discarding the token.
// This endpoint provides a clean API contract.
func (h *AuthHandler) Logout(c *gin.Context) {
	response.OK(c, "Logout successful", nil)
}

// SendOTP handles POST /api/v1/auth/send-otp
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Valid email is required", nil)
		return
	}

	if err := h.authService.SendOTP(input.Email); err != nil {
		response.InternalError(c, "Failed to send OTP")
		return
	}

	response.OK(c, "OTP code has been sent to your email", nil)
}

// VerifyOTP handles POST /api/v1/auth/verify-otp
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Email and 6-digit code are required", nil)
		return
	}

	ok, err := h.authService.VerifyOTP(input.Email, input.Code)
	if err != nil || !ok {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.OK(c, "OTP verified successfully", nil)
}
