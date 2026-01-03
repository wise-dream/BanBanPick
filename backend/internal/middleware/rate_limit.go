package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter управляет rate limiting
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // Количество запросов
	window   time.Duration // Временное окно
	cleanup  *time.Ticker
}

// Visitor представляет посетителя с его запросами
type Visitor struct {
	lastSeen time.Time
	count    int
	mu       sync.Mutex
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	// Очистка старых записей каждые 5 минут
	rl.cleanup = time.NewTicker(5 * time.Minute)
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors удаляет старых посетителей
func (rl *RateLimiter) cleanupVisitors() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, visitor := range rl.visitors {
			visitor.mu.Lock()
			if now.Sub(visitor.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
			visitor.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// Allow проверяет, разрешен ли запрос
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			lastSeen: time.Now(),
			count:    1,
		}
		rl.visitors[ip] = visitor
		rl.mu.Unlock()
		return true
	}
	rl.mu.Unlock()

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	now := time.Now()

	// Если прошло больше окна, сбрасываем счетчик
	if now.Sub(visitor.lastSeen) > rl.window {
		visitor.count = 1
		visitor.lastSeen = now
		return true
	}

	// Проверяем лимит
	if visitor.count >= rl.rate {
		return false
	}

	visitor.count++
	visitor.lastSeen = now
	return true
}

// RateLimitMiddleware создает middleware для rate limiting
func RateLimitMiddleware(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		// Пытаемся получить user ID из контекста (если пользователь авторизован)
		// Это позволяет разным пользователям иметь свои лимиты
		var identifier string
		user, err := GetUserFromContext(c)
		if err == nil && user != nil {
			// Используем user ID для авторизованных пользователей
			identifier = fmt.Sprintf("user:%d", user.ID)
		} else {
			// Fallback на IP для неавторизованных пользователей
			identifier = fmt.Sprintf("ip:%s", c.ClientIP())
		}

		if !limiter.Allow(identifier) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DefaultRateLimitMiddleware создает middleware с настройками по умолчанию
func DefaultRateLimitMiddleware() gin.HandlerFunc {
	// 100 запросов в минуту по умолчанию
	return RateLimitMiddleware(100, time.Minute)
}

// StrictRateLimitMiddleware создает middleware с более строгими лимитами
func StrictRateLimitMiddleware() gin.HandlerFunc {
	// 20 запросов в минуту для строгих endpoints
	return RateLimitMiddleware(20, time.Minute)
}
