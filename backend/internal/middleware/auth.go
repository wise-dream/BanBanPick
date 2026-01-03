package middleware

import (
	"net/http"
	"strings"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	// UserContextKey ключ для хранения пользователя в контексте
	UserContextKey = "user"
)

// AuthMiddleware создает middleware для проверки JWT токена
func AuthMiddleware(jwtService *jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Валидируем токен
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Сохраняем пользователя в контексте
		user := &entities.User{
			ID:       claims.UserID,
			Username: claims.Username,
		}
		c.Set(UserContextKey, user)

		c.Next()
	}
}

// GetUserFromContext извлекает пользователя из контекста
func GetUserFromContext(c *gin.Context) (*entities.User, error) {
	userInterface, exists := c.Get(UserContextKey)
	if !exists {
		return nil, gin.Error{Err: http.ErrNoLocation, Meta: "user not found in context"}
	}

	user, ok := userInterface.(*entities.User)
	if !ok {
		return nil, gin.Error{Err: http.ErrNoLocation, Meta: "invalid user type in context"}
	}

	return user, nil
}