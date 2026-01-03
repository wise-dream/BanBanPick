package veto

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetSessionUseCase struct {
	sessionRepo repositories.VetoSessionRepository
}

type GetSessionOutput struct {
	Session *entities.VetoSession
}

func NewGetSessionUseCase(sessionRepo repositories.VetoSessionRepository) *GetSessionUseCase {
	return &GetSessionUseCase{
		sessionRepo: sessionRepo,
	}
}

func (uc *GetSessionUseCase) Execute(sessionID uint) (*GetSessionOutput, error) {
	session, err := uc.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	return &GetSessionOutput{
		Session: session,
	}, nil
}

// GetSessionByShareToken получает сессию по share token
func (uc *GetSessionUseCase) ExecuteByShareToken(shareToken string) (*GetSessionOutput, error) {
	session, err := uc.sessionRepo.GetByShareToken(shareToken)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	return &GetSessionOutput{
		Session: session,
	}, nil
}