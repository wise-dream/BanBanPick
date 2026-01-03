package user

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetSessionsUseCase struct {
	sessionRepo repositories.VetoSessionRepository
}

type GetSessionsOutput struct {
	Sessions []entities.VetoSession
}

func NewGetSessionsUseCase(sessionRepo repositories.VetoSessionRepository) *GetSessionsUseCase {
	return &GetSessionsUseCase{
		sessionRepo: sessionRepo,
	}
}

func (uc *GetSessionsUseCase) Execute(userID uint) (*GetSessionsOutput, error) {
	sessions, err := uc.sessionRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &GetSessionsOutput{
		Sessions: sessions,
	}, nil
}