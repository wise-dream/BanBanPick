package auth

import (
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/pkg/jwt"
	"github.com/bbp/backend/pkg/password"
)

type LoginUseCase struct {
	userRepo   repositories.UserRepository
	jwtService *jwt.JWTService
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token string
	User  *LoginUser
}

type LoginUser struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

func NewLoginUseCase(userRepo repositories.UserRepository, jwtService *jwt.JWTService) *LoginUseCase {
	return &LoginUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (uc *LoginUseCase) Execute(input LoginInput) (*LoginOutput, error) {
	// Находим пользователя по email
	user, err := uc.userRepo.GetByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Проверяем пароль
	if !password.CheckPassword(user.Password, input.Password) {
		return nil, ErrInvalidCredentials
	}

	// Генерируем JWT токен
	token, err := uc.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// Возвращаем данные пользователя (без пароля)
	loginUser := &LoginUser{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return &LoginOutput{
		Token: token,
		User:  loginUser,
	}, nil
}