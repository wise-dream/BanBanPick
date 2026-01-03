package auth

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/pkg/jwt"
	"github.com/bbp/backend/pkg/password"
)

type RegisterUseCase struct {
	userRepo  repositories.UserRepository
	jwtService *jwt.JWTService
}

type RegisterInput struct {
	Email    string
	Username string
	Password string
}

type RegisterOutput struct {
	Token string
	User  *entities.User
}

func NewRegisterUseCase(userRepo repositories.UserRepository, jwtService *jwt.JWTService) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (uc *RegisterUseCase) Execute(input RegisterInput) (*RegisterOutput, error) {
	// Проверяем уникальность email
	existingUser, err := uc.userRepo.GetByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Проверяем уникальность username
	existingUser, err = uc.userRepo.GetByUsername(input.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameAlreadyExists
	}

	// Хешируем пароль
	hashedPassword, err := password.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Создаем пользователя
	user := &entities.User{
		Email:    input.Email,
		Username: input.Username,
		Password: hashedPassword,
	}

	// Валидируем пользователя
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем пользователя
	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Генерируем JWT токен
	token, err := uc.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{
		Token: token,
		User:  user,
	}, nil
}