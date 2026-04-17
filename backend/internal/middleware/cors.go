package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/digitalpapyrus/backend/internal/config"
)

// CORSMiddleware configures Cross-Origin Resource Sharing for the frontend.
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours preflight cache
	}

	// In development, allow all origins for easier testing
	if !cfg.IsProduction() {
		corsConfig.AllowOrigins = []string{"*"}
		corsConfig.AllowCredentials = false
	}

	return cors.New(corsConfig)
}
