package auth

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetCurrentUserUseCase struct {
	userRepo repositories.UserRepository
}

type GetCurrentUserOutput struct {
	User *entities.User
}

func NewGetCurrentUserUseCase(userRepo repositories.UserRepository) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{
		userRepo: userRepo,
	}
}

func (uc *GetCurrentUserUseCase) Execute(userID uint) (*GetCurrentUserOutput, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &GetCurrentUserOutput{
		User: user,
	}, nil
}