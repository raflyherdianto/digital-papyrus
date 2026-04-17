// Package config provides centralized application configuration
// loaded from environment variables with sensible defaults.
package config

import (
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
}

// AppConfig holds general application settings.
type AppConfig struct {
	Env  string
	Port string
	Name string
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Path string
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

// SeedConfig holds initial superadmin seeding configuration.
type SeedConfig struct {
	SuperAdminEmail    string
	SuperAdminPassword string
	SuperAdminName     string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
			Name: getEnv("APP_NAME", "digital-papyrus-api"),
		},
		DB: DBConfig{
			Path: getEnv("DB_PATH", "./data/digital_papyrus.db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "dev-secret-change-in-production-immediately"),
			ExpiryTime: time.Duration(getEnvAsInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"https://digitalpapyrus.web.id"}, ","),
		},
		Rate: RateConfig{
			General: getEnvAsInt("RATE_LIMIT_GENERAL", 100),
			Auth:    getEnvAsInt("RATE_LIMIT_AUTH", 5),
		},
		Security: SecurityConfig{
			BcryptCost: getEnvAsInt("BCRYPT_COST", 12),
		},
		Seed: SeedConfig{
			SuperAdminEmail:    getEnv("SEED_SUPERADMIN_EMAIL", "superadmin@digitalpapyrus.web.id"),
			SuperAdminPassword: getEnv("SEED_SUPERADMIN_PASSWORD", "SuperAdmin@2026!"),
			SuperAdminName:     getEnv("SEED_SUPERADMIN_NAME", "Super Admin"),
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
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
