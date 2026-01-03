package user

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetProfileUseCase struct {
	userRepo repositories.UserRepository
}

type GetProfileOutput struct {
	User *entities.User
}

func NewGetProfileUseCase(userRepo repositories.UserRepository) *GetProfileUseCase {
	return &GetProfileUseCase{
		userRepo: userRepo,
	}
}

func (uc *GetProfileUseCase) Execute(userID uint) (*GetProfileOutput, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &GetProfileOutput{
		User: user,
	}, nil
}