// Package middleware provides HTTP middleware for the Gin router.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/model"
	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
)

// contextKey constants for storing user info in request context.
const (
	ContextKeyUserID = "user_id"
	ContextKeyEmail  = "user_email"
	ContextKeyRole   = "user_role"
	ContextKeyName   = "user_name"
)

// AuthMiddleware validates JWT tokens from the Authorization header.
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid authorization header format. Use: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Store claims in context for downstream handlers
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyName, claims.Name)

		c.Next()
	}
}

// RequireRole creates middleware that restricts access to specific roles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || !roleSet[roleStr] {
			response.Forbidden(c, "You do not have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience middleware that allows only superadmin access.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleSuperAdmin)
}

// RequireAdminOrAuthor allows superadmin and author access.
func RequireAdminOrAuthor() gin.HandlerFunc {
	return RequireRole(model.RoleSuperAdmin, model.RoleAuthor)
}
