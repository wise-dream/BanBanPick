package password

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost для bcrypt хеширования (10 - хороший баланс между безопасностью и производительностью)
	DefaultCost = 10
)

// HashPassword хеширует пароль с использованием bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword проверяет, соответствует ли пароль хешу
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}