// Package router wires all routes, middleware, and handlers together.
package router

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/handler"
	"github.com/digitalpapyrus/backend/internal/middleware"
	"github.com/digitalpapyrus/backend/internal/service"
	"github.com/digitalpapyrus/backend/pkg/response"
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
	Region      *handler.RegionHandler
	Setting     *handler.SettingHandler
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

	// Serve static uploads locally
	uploadDir := filepath.Join("frontend", "public", "uploads")
	if _, err := os.Stat(filepath.Join("..", "frontend", "public", "uploads")); err == nil {
		uploadDir = filepath.Join("..", "frontend", "public", "uploads")
	}
	r.Static("/uploads", uploadDir)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Health check (public)
		v1.GET("/health", h.Health.HealthCheck)

		// Region routes (public)
		regions := v1.Group("/regions")
		{
			regions.GET("/provinces", h.Region.GetProvinces)
			regions.GET("/provinces/:code/regencies", h.Region.GetRegenciesByProvince)
			regions.GET("/regencies/:code/districts", h.Region.GetDistrictsByRegency)
			regions.GET("/districts/:code/villages", h.Region.GetVillagesByDistrict)
		}


		// Auth routes (public with stricter rate limiting)
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimitMiddleware(cfg))
		{
			auth.POST("/login", h.Auth.Login)
			auth.POST("/register", h.Auth.Register)
			auth.POST("/send-otp", h.Auth.SendOTP)
			auth.POST("/verify-otp", h.Auth.VerifyOTP)
			auth.POST("/forgot-password", h.Auth.ForgotPassword)
			auth.POST("/reset-password", h.Auth.ResetPassword)
		}

		// Auth routes (protected)
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(authService))
		{
			authProtected.GET("/me", h.Auth.Me)
			authProtected.PUT("/me", h.Auth.UpdateProfile)
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

		// Book delete & validate (protected: admin only)
		booksAdmin := v1.Group("/books")
		booksAdmin.Use(middleware.AuthMiddleware(authService))
		booksAdmin.Use(middleware.RequireAdmin())
		{
			booksAdmin.DELETE("/:id", h.Book.DeleteBook)
			booksAdmin.POST("/:id/validate", h.Book.ValidateBook)
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

		// Draft upload route (protected: any authenticated user)
		uploadDraft := v1.Group("/upload/draft")
		uploadDraft.Use(middleware.AuthMiddleware(authService))
		{
			uploadDraft.POST("", h.Upload.UploadDraft)
		}

		// Customer book publishing request routes (protected: any authenticated user)
		customerBooks := v1.Group("/customer/books")
		customerBooks.Use(middleware.AuthMiddleware(authService))
		{
			customerBooks.POST("", h.Book.CreateCustomerBook)
			customerBooks.PUT("/:id", h.Book.UpdateCustomerBook)
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
			ordersAdmin.POST("/:id/confirm-payment", h.Order.AdminConfirmPayment)
			ordersAdmin.PUT("/:id/status", h.Order.UpdateOrderStatus)
			ordersAdmin.DELETE("/:id", h.Order.DeleteOrder)
		}

		// Customer orders routes (protected: any authenticated user)
		customerOrders := v1.Group("/customer/orders")
		customerOrders.Use(middleware.AuthMiddleware(authService))
		{
			customerOrders.GET("", h.Order.ListCustomerOrders)
			customerOrders.POST("", h.Order.CreateCustomerOrder)
			customerOrders.GET("/:id", h.Order.GetCustomerOrder)
			customerOrders.POST("/:id/check-payment", h.Order.CheckPayment)
			customerOrders.POST("/:id/confirm-payment", h.Order.ConfirmCustomerPayment)
			customerOrders.POST("/:id/cancel", h.Order.CancelCustomerOrder)
		}
		
		// Customer shipping cost route
		customerShipping := v1.Group("/customer")
		customerShipping.Use(middleware.AuthMiddleware(authService))
		{
			customerShipping.POST("/shipping-cost", h.Order.CalculateShippingCost)
		}

		// Public invoice route (accessible for sharing & PDF saving)
		v1.GET("/invoices/:id", h.Order.GetPublicInvoice)

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

		// Settings routes (public for GET, protected for POST)
		settings := v1.Group("/settings")
		{
			settings.GET("", h.Setting.GetSettings)
		}

		settingsAdmin := v1.Group("/settings")
		settingsAdmin.Use(middleware.AuthMiddleware(authService))
		settingsAdmin.Use(middleware.RequireAdmin())
		{
			settingsAdmin.POST("", h.Setting.UpdateSettings)
		}
	}

	// 404 Handler for undefined routes
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "Endpoint not found")
	})

	return r
}
