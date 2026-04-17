// Package router wires all routes, middleware, and handlers together.
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/handler"
	"github.com/digitalpapyrus/backend/internal/middleware"
	"github.com/digitalpapyrus/backend/internal/service"
)

// Handlers holds all handler instances for route registration.
type Handlers struct {
	Health   *handler.HealthHandler
	Auth     *handler.AuthHandler
	Book     *handler.BookHandler
	Service  *handler.ServiceHandler
	Category *handler.CategoryHandler
	Upload   *handler.UploadHandler
}

// Setup creates and configures the Gin engine with all routes.
func Setup(cfg *config.Config, authService *service.AuthService, h Handlers) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Trust only proxy headers from Cloudflare Tunnel / reverse proxy
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1", "172.16.0.0/12", "10.0.0.0/8"})

	// Global middleware stack (order matters)
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.SecurityHeaders(cfg))
	r.Use(middleware.CORSMiddleware(cfg))
	r.Use(middleware.RateLimitMiddleware(cfg))
	r.Use(gin.Logger())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Health check (public)
		v1.GET("/health", h.Health.HealthCheck)

		// Auth routes (public with stricter rate limiting)
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimitMiddleware(cfg))
		{
			auth.POST("/login", h.Auth.Login)
		}

		// Auth routes (protected)
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(authService))
		{
			authProtected.GET("/me", h.Auth.Me)
			authProtected.POST("/logout", h.Auth.Logout)
		}

		// Book routes (public)
		books := v1.Group("/books")
		{
			books.GET("", h.Book.ListBooks)
			books.GET("/:id", h.Book.GetBook)
		}

		// Book routes (protected: admin + author)
		booksProtected := v1.Group("/books")
		booksProtected.Use(middleware.AuthMiddleware(authService))
		booksProtected.Use(middleware.RequireAdminOrAuthor())
		{
			booksProtected.POST("", h.Book.CreateBook)
			booksProtected.PUT("/:id", h.Book.UpdateBook)
		}

		// Book delete (protected: admin only)
		booksAdmin := v1.Group("/books")
		booksAdmin.Use(middleware.AuthMiddleware(authService))
		booksAdmin.Use(middleware.RequireAdmin())
		{
			booksAdmin.DELETE("/:id", h.Book.DeleteBook)
		}

		// Category routes (public)
		categories := v1.Group("/categories")
		{
			categories.GET("", h.Category.ListCategories)
			categories.GET("/:id", h.Category.GetCategory)
		}

		// Category routes (protected: admin only)
		categoriesAdmin := v1.Group("/categories")
		categoriesAdmin.Use(middleware.AuthMiddleware(authService))
		categoriesAdmin.Use(middleware.RequireAdmin())
		{
			categoriesAdmin.POST("", h.Category.CreateCategory)
			categoriesAdmin.PUT("/:id", h.Category.UpdateCategory)
			categoriesAdmin.DELETE("/:id", h.Category.DeleteCategory)
		}

		// Upload route (protected: admin + author)
		upload := v1.Group("/upload")
		upload.Use(middleware.AuthMiddleware(authService))
		upload.Use(middleware.RequireAdminOrAuthor())
		{
			upload.POST("", h.Upload.UploadImage)
		}

		// Service routes (public)
		services := v1.Group("/services")
		{
			services.GET("", h.Service.ListServices)
			services.GET("/:id", h.Service.GetService)
		}

		// Service routes (protected: admin only)
		servicesAdmin := v1.Group("/services")
		servicesAdmin.Use(middleware.AuthMiddleware(authService))
		servicesAdmin.Use(middleware.RequireAdmin())
		{
			servicesAdmin.POST("", h.Service.CreateService)
			servicesAdmin.PUT("/:id", h.Service.UpdateService)
			servicesAdmin.DELETE("/:id", h.Service.DeleteService)
		}
	}

	return r
}
