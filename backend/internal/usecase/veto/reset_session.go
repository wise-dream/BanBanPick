package veto

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type ResetSessionUseCase struct {
	sessionRepo repositories.VetoSessionRepository
	actionRepo  repositories.VetoActionRepository
}

type ResetSessionInput struct {
	SessionID uint
}

type ResetSessionOutput struct {
	Session *entities.VetoSession
}

func NewResetSessionUseCase(
	sessionRepo repositories.VetoSessionRepository,
	actionRepo repositories.VetoActionRepository,
) *ResetSessionUseCase {
	return &ResetSessionUseCase{
		sessionRepo: sessionRepo,
		actionRepo:  actionRepo,
	}
}

func (uc *ResetSessionUseCase) Execute(input ResetSessionInput) (*ResetSessionOutput, error) {
	// Получаем сессию
	session, err := uc.sessionRepo.GetByID(input.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Удаляем все действия
	if err := uc.actionRepo.DeleteBySessionID(input.SessionID); err != nil {
		return nil, err
	}

	// Сбрасываем состояние сессии
	session.Status = entities.VetoStatusNotStarted
	session.CurrentTeam = "A"
	session.SelectedMapID = nil
	session.SelectedSide = nil
	session.FinishedAt = nil
	session.Actions = []entities.VetoAction{}

	// Обновляем сессию
	if err := uc.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	// Получаем обновленную сессию
	updatedSession, err := uc.sessionRepo.GetByID(session.ID)
	if err != nil {
		return nil, err
	}

	return &ResetSessionOutput{
		Session: updatedSession,
	}, nil
}