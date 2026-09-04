// Package testutil provides shared test setup utilities.
package testutil

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/digitalpapyrus/backend/internal/config"
	"github.com/digitalpapyrus/backend/internal/database"
	"github.com/digitalpapyrus/backend/internal/handler"
	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/internal/router"
	"github.com/digitalpapyrus/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// TestEnv holds all test dependencies.
type TestEnv struct {
	DB             *sql.DB
	Config         *config.Config
	Router         *gin.Engine
	AuthService    *service.AuthService
	BookService    *service.BookService
	ServiceService *service.ServiceService
	UserRepo       *repository.UserRepository
	BookRepo       *repository.BookRepository
	ServiceRepo    *repository.ServiceRepository
	OrderRepo      *repository.OrderRepository
}

// SetupTestEnv creates a clean test environment with a PostgreSQL database.
func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
	}

	dbPort := os.Getenv("TEST_DB_PORT")
	if dbPort == "" {
		dbPort = os.Getenv("DB_PORT")
		if dbPort == "" {
			dbPort = "5432"
		}
	}

	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "postgres"
		}
	}

	dbPassword := os.Getenv("TEST_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			dbPassword = "postgres"
		}
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "digital_papyrus"
		}
	}

	cfg := &config.Config{
		App: config.AppConfig{
			Env:  "test",
			Port: "0",
			Name: "digital-papyrus-test",
		},
		DB: config.DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret-for-unit-tests-only-not-production",
			ExpiryTime: 1 * time.Hour,
		},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
		},
		Rate: config.RateConfig{
			General: 1000,
			Auth:    100,
		},
		Security: config.SecurityConfig{
			BcryptCost: 4, // Low cost for fast tests
		},
		Seed: config.SeedConfig{
			SuperAdminEmail:    "test-admin@test.com",
			SuperAdminPassword: "TestAdmin@2026!",
			SuperAdminName:     "Test Admin",
		},
	}

	db, err := database.New(cfg.DB.DSN())
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	if err := database.Seed(db, cfg); err != nil {
		t.Fatalf("failed to seed test database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg)
	bookService := service.NewBookService(bookRepo, userRepo, orderRepo, cfg)
	serviceService := service.NewServiceService(serviceRepo)

	// Initialize handlers
	handlers := router.Handlers{
		Health:  handler.NewHealthHandler(),
		Auth:    handler.NewAuthHandler(authService),
		Book:    handler.NewBookHandler(bookService),
		Service: handler.NewServiceHandler(serviceService),
	}

	engine := router.Setup(cfg, authService, handlers)

	t.Cleanup(func() {
		db.Close()
	})

	return &TestEnv{
		DB:             db,
		Config:         cfg,
		Router:         engine,
		AuthService:    authService,
		BookService:    bookService,
		ServiceService: serviceService,
		UserRepo:       userRepo,
		BookRepo:       bookRepo,
		ServiceRepo:    serviceRepo,
		OrderRepo:      orderRepo,
	}
}
