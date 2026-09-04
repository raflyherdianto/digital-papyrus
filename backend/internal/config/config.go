// Package config provides centralized application configuration
// loaded from environment variables with sensible defaults.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration values.
type Config struct {
	App      AppConfig
	DB       DBConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Rate     RateConfig
	Security SecurityConfig
	Seed     SeedConfig
	SMTP     SMTPConfig
	API      APIConfig
}

// APIConfig holds external API keys.
type APIConfig struct {
	CoIdKey string
}

// SMTPConfig holds SMTP mail server settings.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

// AppConfig holds general application settings.
type AppConfig struct {
	Env  string
	Port string
	Name string
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	URL      string
}

func (c *DBConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret     string
	ExpiryTime time.Duration
}

// CORSConfig holds Cross-Origin Resource Sharing settings.
type CORSConfig struct {
	AllowedOrigins []string
}

// RateConfig holds rate limiting settings.
type RateConfig struct {
	General int
	Auth    int
}

// SecurityConfig holds security-related settings.
type SecurityConfig struct {
	BcryptCost int
}

// SeedConfig holds initial user seeding configuration.
type SeedConfig struct {
	SuperAdminEmail    string
	SuperAdminPassword string
	SuperAdminName     string
	DemoAuthorEmail    string
	DemoAuthorPassword string
	DemoAuthorName     string
	DemoCustomerEmail    string
	DemoCustomerPassword string
	DemoCustomerName     string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	loadEnvFile(".env")
	loadEnvFile("../.env")

	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
			Name: getEnv("APP_NAME", "digital-papyrus-api"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "digital_papyrus"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			URL:      getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "local-dev-secret-key"),
			ExpiryTime: time.Duration(getEnvAsInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"https://digitalpapyrus.web.id", "https://www.digitalpapyrus.web.id", "http://localhost:4321"}, ","),
		},
		Rate: RateConfig{
			General: getEnvAsInt("RATE_LIMIT_GENERAL", 100),
			Auth:    getEnvAsInt("RATE_LIMIT_AUTH", 5),
		},
		Security: SecurityConfig{
			BcryptCost: getEnvAsInt("BCRYPT_COST", 12),
		},
		Seed: SeedConfig{
			SuperAdminEmail:    getEnv("SEED_SUPERADMIN_EMAIL", "admin@local.dev"),
			SuperAdminPassword: getEnv("SEED_SUPERADMIN_PASSWORD", "local-dev-password"),
			SuperAdminName:     getEnv("SEED_SUPERADMIN_NAME", "Local Admin"),
			DemoAuthorEmail:    getEnv("SEED_DEMO_AUTHOR_EMAIL", "author@digitalpapyrus.web.id"),
			DemoAuthorPassword: getEnv("SEED_DEMO_AUTHOR_PASSWORD", "Demo@2026!"),
			DemoAuthorName:     getEnv("SEED_DEMO_AUTHOR_NAME", "Demo Author"),
			DemoCustomerEmail:    getEnv("SEED_DEMO_CUSTOMER_EMAIL", "customer@digitalpapyrus.web.id"),
			DemoCustomerPassword: getEnv("SEED_DEMO_CUSTOMER_PASSWORD", "Demo@2026!"),
			DemoCustomerName:     getEnv("SEED_DEMO_CUSTOMER_NAME", "Demo Customer"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
		},
		API: APIConfig{
			CoIdKey: getEnv("API_CO_ID_KEY", ""),
		},
	}
}

// IsProduction returns true if the application is running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsSlice(key string, defaultVal []string, sep string) []string {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	parts := strings.Split(valStr, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, v)
			}
		}
	}
}
