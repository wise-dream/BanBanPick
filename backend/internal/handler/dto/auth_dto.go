package dto

import "github.com/bbp/backend/internal/domain/entities"

// RegisterRequest DTO для регистрации
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest DTO для входа
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse DTO для ответов авторизации
type AuthResponse struct {
	Token string      `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse DTO для данных пользователя в ответах
type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

// ToUserResponse конвертирует entity User в UserResponse
func ToUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}