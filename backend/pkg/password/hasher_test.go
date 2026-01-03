package password

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hashed == "" {
		t.Error("HashPassword() returned empty hash")
	}

	if hashed == password {
		t.Error("HashPassword() returned the same value as input")
	}

	// Хеш должен быть разным при каждом вызове (из-за соли)
	hashed2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hashed == hashed2 {
		t.Error("HashPassword() should return different hashes for the same password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Правильный пароль
	if !CheckPassword(hashed, password) {
		t.Error("CheckPassword() should return true for correct password")
	}

	// Неправильный пароль
	if CheckPassword(hashed, "wrongpassword") {
		t.Error("CheckPassword() should return false for incorrect password")
	}

	// Пустой пароль
	if CheckPassword(hashed, "") {
		t.Error("CheckPassword() should return false for empty password")
	}
}

func TestCheckPasswordWithDifferentPasswords(t *testing.T) {
	password1 := "password1"
	password2 := "password2"

	hashed1, err := HashPassword(password1)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	hashed2, err := HashPassword(password2)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Хеши должны быть разными
	if hashed1 == hashed2 {
		t.Error("Hashes for different passwords should be different")
	}

	// Каждый хеш должен проверяться только со своим паролем
	if CheckPassword(hashed1, password2) {
		t.Error("CheckPassword() should return false for wrong password")
	}

	if CheckPassword(hashed2, password1) {
		t.Error("CheckPassword() should return false for wrong password")
	}
}

func TestCheckPasswordWithSpecialCharacters(t *testing.T) {
	passwords := []string{
		"password123",
		"password!@#$%",
		"password with spaces",
		"password\nwith\nnewlines",
		"пароль123",
		"verylongpasswordthatexceedsnormallengthandshouldstillworkcorrectly",
	}

	for _, password := range passwords {
		hashed, err := HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword() error = %v for password: %q", err, password)
		}

		if !CheckPassword(hashed, password) {
			t.Errorf("CheckPassword() should return true for password: %q", password)
		}
	}
}
