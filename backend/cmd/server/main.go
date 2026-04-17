// Digital Papyrus API — Main entry point.
//
// A production-grade RESTful backend for the Digital Papyrus
// book publishing platform, built with Gin and SQLite.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/database"
	"github.com/digitalpapyrus/backend/internal/handler"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/internal/router"
	"github.com/digitalpapyrus/backend/internal/service"
)

func main() {
	// Load .env file (optional, ignored in production containers)
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	log.Printf("[APP] Starting %s (env=%s, port=%s)", cfg.App.Name, cfg.App.Env, cfg.App.Port)

	// Initialize database
	db, err := database.New(cfg.DB.Path)
	if err != nil {
		log.Fatalf("[FATAL] Database connection failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("[FATAL] Database migration failed: %v", err)
	}

	// Seed initial data
	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("[FATAL] Database seeding failed: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
        categoryRepo := repository.NewCategoryRepository(db)

        // Initialize services
        authService := service.NewAuthService(userRepo, cfg)
        bookService := service.NewBookService(bookRepo)
        serviceService := service.NewServiceService(serviceRepo)
        categoryService := service.NewCategoryService(categoryRepo)

        // Initialize handlers
        handlers := router.Handlers{
                Health:   handler.NewHealthHandler(),
                Auth:     handler.NewAuthHandler(authService),
                Book:     handler.NewBookHandler(bookService),
                Service:  handler.NewServiceHandler(serviceService),
                Category: handler.NewCategoryHandler(categoryService),
                Upload:   handler.NewUploadHandler(),
        }

        // Configure router
        engine := router.Setup(cfg, authService, handlers)

	// Create HTTP server with production-grade timeouts
	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           engine,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// Graceful shutdown
	go func() {
		log.Printf("[APP] Server listening on :%s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[APP] Received signal %v, shutting down gracefully...", sig)

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[FATAL] Server forced shutdown: %v", err)
	}

	log.Println("[APP] Server exited gracefully")
}
