package veto

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type SelectSideUseCase struct {
	sessionRepo repositories.VetoSessionRepository
}

type SelectSideInput struct {
	SessionID uint
	Side      string // "attack" или "defence"
}

type SelectSideOutput struct {
	Session *entities.VetoSession
}

func NewSelectSideUseCase(sessionRepo repositories.VetoSessionRepository) *SelectSideUseCase {
	return &SelectSideUseCase{
		sessionRepo: sessionRepo,
	}
}

func (uc *SelectSideUseCase) Execute(input SelectSideInput) (*SelectSideOutput, error) {
	// Получаем сессию
	session, err := uc.sessionRepo.GetByID(input.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Проверяем, что сессия завершена
	if session.Status != entities.VetoStatusFinished {
		return nil, ErrInvalidAction
	}

	// Проверяем, что выбрана карта
	if session.SelectedMapID == nil {
		return nil, ErrInvalidAction
	}

	// Валидируем сторону
	if input.Side != "attack" && input.Side != "defence" {
		return nil, ErrInvalidAction
	}

	// Устанавливаем сторону
	session.SelectedSide = &input.Side

	// Обновляем сессию
	if err := uc.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	// Получаем обновленную сессию
	updatedSession, err := uc.sessionRepo.GetByID(session.ID)
	if err != nil {
		return nil, err
	}

	return &SelectSideOutput{
		Session: updatedSession,
	}, nil
}