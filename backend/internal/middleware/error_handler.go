package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware создает middleware для централизованной обработки ошибок
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверяем наличие ошибок
		if len(c.Errors) > 0 {
			// Логируем ошибки
			for _, err := range c.Errors {
				log.Printf("Error: %v", err.Error())
			}

			// Если ответ еще не отправлен, отправляем ошибку
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}
	}
}