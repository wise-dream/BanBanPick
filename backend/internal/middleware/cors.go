package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware создает middleware для CORS
// Поддерживает несколько origins через запятую в переменной окружения CORS_ORIGIN
// Если CORS_ORIGIN не установлен или равен "*", разрешаются все origins
// Например: CORS_ORIGIN="http://localhost:5173,https://ban.wise-dream.site"
func CORSMiddleware(origin string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	
	// Если origin пустой или "*", разрешаем все origins
	if origin == "" || origin == "*" {
		config.AllowAllOrigins = true
		// При AllowAllOrigins нельзя использовать AllowCredentials
		// Но это нормально для большинства случаев
		config.AllowCredentials = false
	} else {
		// Разбираем origins (может быть несколько через запятую)
		origins := parseOrigins(origin)
		if len(origins) > 0 {
			config.AllowOrigins = origins
			config.AllowOriginFunc = nil
			config.AllowCredentials = true
		} else {
			// Fallback: разрешаем все, если список пустой
			config.AllowAllOrigins = true
			config.AllowCredentials = false
		}
	}
	
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.MaxAge = 12 * time.Hour // Кеширование preflight запросов

	return cors.New(config)
}

// parseOrigins разбирает строку с origins, разделенными запятой
func parseOrigins(originsStr string) []string {
	parts := strings.Split(originsStr, ",")
	origins := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	
	return origins
}