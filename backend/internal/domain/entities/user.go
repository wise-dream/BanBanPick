package entities

import (
	"errors"
	"time"
)

type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // не возвращается в JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate проверяет валидность данных пользователя
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if len(u.Username) > 50 {
		return errors.New("username must be no more than 50 characters")
	}
	if u.Password == "" {
		return errors.New("password is required")
	}
	return nil
}
