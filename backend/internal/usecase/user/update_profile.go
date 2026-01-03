package user

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type UpdateProfileUseCase struct {
	userRepo repositories.UserRepository
}

type UpdateProfileInput struct {
	UserID   uint
	Email    *string
	Username *string
}

type UpdateProfileOutput struct {
	User *entities.User
}

func NewUpdateProfileUseCase(userRepo repositories.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		userRepo: userRepo,
	}
}

func (uc *UpdateProfileUseCase) Execute(input UpdateProfileInput) (*UpdateProfileOutput, error) {
	// Получаем текущего пользователя
	user, err := uc.userRepo.GetByID(input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Обновляем поля, если они указаны
	if input.Email != nil {
		// Проверяем уникальность email (если изменился)
		if *input.Email != user.Email {
			existingUser, err := uc.userRepo.GetByEmail(*input.Email)
			if err != nil {
				return nil, err
			}
			if existingUser != nil {
				return nil, ErrEmailAlreadyExists
			}
		}
		user.Email = *input.Email
	}

	if input.Username != nil {
		// Проверяем уникальность username (если изменился)
		if *input.Username != user.Username {
			existingUser, err := uc.userRepo.GetByUsername(*input.Username)
			if err != nil {
				return nil, err
			}
			if existingUser != nil {
				return nil, ErrUsernameAlreadyExists
			}
		}
		user.Username = *input.Username
	}

	// Валидируем пользователя
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем изменения
	if err := uc.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &UpdateProfileOutput{
		User: user,
	}, nil
}