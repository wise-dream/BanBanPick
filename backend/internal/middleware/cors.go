package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware создает middleware для CORS
func CORSMiddleware(origin string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	
	// В продакшене проверяем origin строго
	if origin != "" && origin != "*" {
		config.AllowOrigins = []string{origin}
		config.AllowOriginFunc = nil
	} else {
		// Для разработки разрешаем все origins
		config.AllowAllOrigins = true
	}
	
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour // Кеширование preflight запросов

	return cors.New(config)
}