package veto

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type StartSessionUseCase struct {
	sessionRepo repositories.VetoSessionRepository
}

type StartSessionInput struct {
	SessionID uint
}

type StartSessionOutput struct {
	Session *entities.VetoSession
}

func NewStartSessionUseCase(
	sessionRepo repositories.VetoSessionRepository,
) *StartSessionUseCase {
	return &StartSessionUseCase{
		sessionRepo: sessionRepo,
	}
}

func (uc *StartSessionUseCase) Execute(input StartSessionInput) (*StartSessionOutput, error) {
	// Получаем сессию
	session, err := uc.sessionRepo.GetByID(input.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Проверяем, что сессия еще не начата
	if session.Status != entities.VetoStatusNotStarted {
		return nil, ErrSessionAlreadyStarted
	}

	// Обновляем статус сессии на in_progress
	session.Status = entities.VetoStatusInProgress

	// Обновляем сессию в БД
	if err := uc.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	// Получаем обновленную сессию
	updatedSession, err := uc.sessionRepo.GetByID(session.ID)
	if err != nil {
		return nil, err
	}

	return &StartSessionOutput{
		Session: updatedSession,
	}, nil
}
