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
	Category    *handler.CategoryHandler
	Upload      *handler.UploadHandler
	CoreService *handler.CoreServiceHandler
	User        *handler.UserHandler
	Review      *handler.ReviewHandler
	Order       *handler.OrderHandler
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
			auth.POST("/register", h.Auth.Register)
			auth.POST("/send-otp", h.Auth.SendOTP)
			auth.POST("/verify-otp", h.Auth.VerifyOTP)
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

		// CoreService routes (public)
		coreServices := v1.Group("/core-services")
		{
			coreServices.GET("", h.CoreService.GetAll)
			coreServices.GET("/:id", h.CoreService.GetByID)
		}

		// CoreService routes (protected: admin only)
		coreServicesAdmin := v1.Group("/core-services")
		coreServicesAdmin.Use(middleware.AuthMiddleware(authService))
		coreServicesAdmin.Use(middleware.RequireAdmin())
		{
			coreServicesAdmin.POST("", h.CoreService.Create)
			coreServicesAdmin.PUT("/:id", h.CoreService.Update)
			coreServicesAdmin.DELETE("/:id", h.CoreService.Delete)
		}

		// User routes (protected: admin only)
		usersAdmin := v1.Group("/users")
		usersAdmin.Use(middleware.AuthMiddleware(authService))
		usersAdmin.Use(middleware.RequireAdmin())
		{
			usersAdmin.GET("", h.User.ListUsers)
			usersAdmin.POST("", h.User.CreateUser)
			usersAdmin.PUT("/:id", h.User.UpdateUser)
			usersAdmin.DELETE("/:id", h.User.DeleteUser)
		}

		// Orders routes (protected: admin only)
		ordersAdmin := v1.Group("/orders")
		ordersAdmin.Use(middleware.AuthMiddleware(authService))
		ordersAdmin.Use(middleware.RequireAdmin())
		{
			ordersAdmin.GET("", h.Order.ListOrders)
			ordersAdmin.GET("/:id", h.Order.GetOrder)
			ordersAdmin.POST("", h.Order.CreateOrder)
			ordersAdmin.DELETE("/:id", h.Order.DeleteOrder)
		}

		// Review routes (public get, protected modifications)
		reviews := v1.Group("/reviews")
		{
			reviews.GET("", h.Review.GetAllReviews)
			reviews.GET("/:id", h.Review.GetReview)
		}

		reviewsProtected := v1.Group("/reviews")
		reviewsProtected.Use(middleware.AuthMiddleware(authService))
		// Anyone authenticated can create a review (we could limit to customer role if needed)
		{
			reviewsProtected.POST("", h.Review.CreateReview)
		}

		reviewsAdmin := v1.Group("/reviews")
		reviewsAdmin.Use(middleware.AuthMiddleware(authService))
		reviewsAdmin.Use(middleware.RequireAdmin())
		{
			reviewsAdmin.PUT("/:id", h.Review.UpdateReview)
			reviewsAdmin.DELETE("/:id", h.Review.DeleteReview)
		}
	}

	return r
}
