package jwt

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key"
	expiration := 24 * time.Hour
	service := NewJWTService(secret, expiration)

	userID := uint(1)
	username := "testuser"

	token, err := service.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	// Проверяем, что токен можно распарсить
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, username)
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	expiration := 24 * time.Hour
	service := NewJWTService(secret, expiration)

	userID := uint(1)
	username := "testuser"

	token, err := service.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Валидный токен
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, username)
	}

	// Невалидный токен
	invalidToken := "invalid.token.here"
	_, err = service.ValidateToken(invalidToken)
	if err == nil {
		t.Error("ValidateToken() should return error for invalid token")
	}

	// Токен с неправильным секретом
	wrongService := NewJWTService("wrong-secret", expiration)
	_, err = wrongService.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for token with wrong secret")
	}
}

func TestTokenExpiration(t *testing.T) {
	secret := "test-secret-key"
	expiration := 1 * time.Second // Очень короткое время жизни
	service := NewJWTService(secret, expiration)

	userID := uint(1)
	username := "testuser"

	token, err := service.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Токен должен быть валиден сразу
	_, err = service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v (token should be valid)", err)
	}

	// Ждем истечения токена
	time.Sleep(2 * time.Second)

	// Токен должен быть невалиден после истечения
	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should return error for expired token")
	}
}

func TestTokenWithDifferentUsers(t *testing.T) {
	secret := "test-secret-key"
	expiration := 24 * time.Hour
	service := NewJWTService(secret, expiration)

	// Генерируем токены для разных пользователей
	token1, err := service.GenerateToken(1, "user1")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	token2, err := service.GenerateToken(2, "user2")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Проверяем, что токены разные
	if token1 == token2 {
		t.Error("Tokens for different users should be different")
	}

	// Проверяем, что каждый токен содержит правильные данные
	claims1, err := service.ValidateToken(token1)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims1.UserID != 1 || claims1.Username != "user1" {
		t.Errorf("Token1 claims = %+v, want UserID=1, Username=user1", claims1)
	}

	claims2, err := service.ValidateToken(token2)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims2.UserID != 2 || claims2.Username != "user2" {
		t.Errorf("Token2 claims = %+v, want UserID=2, Username=user2", claims2)
	}
}
