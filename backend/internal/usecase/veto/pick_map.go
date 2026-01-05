package veto

import (
	"time"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type PickMapUseCase struct {
	sessionRepo  repositories.VetoSessionRepository
	actionRepo   repositories.VetoActionRepository
	mapRepo      repositories.MapRepository
	mapPoolRepo  repositories.MapPoolRepository
	logicService *VetoLogicService
}

type PickMapInput struct {
	SessionID uint
	MapID     uint
	Team      string
}

type PickMapOutput struct {
	Session *entities.VetoSession
	Action  *entities.VetoAction
}

func NewPickMapUseCase(
	sessionRepo repositories.VetoSessionRepository,
	actionRepo repositories.VetoActionRepository,
	mapRepo repositories.MapRepository,
	mapPoolRepo repositories.MapPoolRepository,
	logicService *VetoLogicService,
) *PickMapUseCase {
	return &PickMapUseCase{
		sessionRepo:  sessionRepo,
		actionRepo:   actionRepo,
		mapRepo:      mapRepo,
		mapPoolRepo:  mapPoolRepo,
		logicService: logicService,
	}
}

func (uc *PickMapUseCase) Execute(input PickMapInput) (*PickMapOutput, error) {
	// Получаем сессию
	session, err := uc.sessionRepo.GetByID(input.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Проверяем статус сессии
	if session.Status == entities.VetoStatusFinished || session.Status == entities.VetoStatusCancelled {
		return nil, ErrSessionFinished
	}

	// Проверяем, что это Bo3 или Bo5
	if session.Type != entities.VetoTypeBo3 && session.Type != entities.VetoTypeBo5 {
		return nil, ErrInvalidAction
	}

	// Получаем пул карт
	mapPool, err := uc.mapPoolRepo.GetByID(session.MapPoolID)
	if err != nil {
		return nil, err
	}
	if mapPool == nil {
		return nil, ErrMapPoolNotFound
	}

	// Проверяем, что карта существует
	mapEntity, err := uc.mapRepo.GetByID(input.MapID)
	if err != nil {
		return nil, err
	}
	if mapEntity == nil {
		return nil, ErrMapNotFound
	}

	// Проверяем, что карта в пуле
	mapInPool := false
	for _, m := range mapPool.Maps {
		if m.ID == input.MapID {
			mapInPool = true
			break
		}
	}
	if !mapInPool {
		return nil, ErrMapNotFound
	}

	// Обновляем статус сессии, если она еще не начата (ПЕРЕД проверкой CanPerformAction)
	if session.Status == entities.VetoStatusNotStarted {
		session.Status = entities.VetoStatusInProgress
		// Обновляем сессию в БД, чтобы статус был актуальным для проверки
		if err := uc.sessionRepo.Update(session); err != nil {
			return nil, err
		}
	}

	// Получаем доступные карты
	availableMaps := uc.logicService.GetAvailableMaps(mapPool, session.Actions)

	// Проверяем, что карта доступна
	mapAvailable := false
	for _, m := range availableMaps {
		if m.ID == input.MapID {
			mapAvailable = true
			break
		}
	}
	if !mapAvailable {
		return nil, ErrMapAlreadyPicked
	}

	// Определяем текущий шаг
	currentStep := uc.logicService.GetCurrentStep(session.Actions)

	// Проверяем возможность выполнения действия (теперь статус уже in_progress)
	if !uc.logicService.CanPerformAction(session, entities.VetoActionTypePick, input.Team, session.Actions, len(availableMaps)) {
		return nil, ErrNotYourTurn
	}

	// Создаем действие
	action := &entities.VetoAction{
		VetoSessionID: session.ID,
		MapID:         input.MapID,
		Team:          input.Team,
		ActionType:    entities.VetoActionTypePick,
		StepNumber:    currentStep,
	}

	if err := action.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем действие
	if err := uc.actionRepo.Create(action); err != nil {
		return nil, err
	}

	// Переключаем команду
	if session.CurrentTeam == "A" {
		session.CurrentTeam = "B"
	} else {
		session.CurrentTeam = "A"
	}

	// Добавляем действие в список для проверок (временно, для проверки логики)
	actionsWithNewPick := append(session.Actions, *action)
	availableMapsAfterPick := uc.logicService.GetAvailableMaps(mapPool, actionsWithNewPick)
	
	// ВАЖНО: Проверяем, нужен ли выбор стороны ПЕРЕД проверкой завершения сессии
	// Если нужен выбор стороны, НЕ устанавливаем статус finished
	needsSideSelection := uc.logicService.NeedsSideSelection(session, actionsWithNewPick)
	
	if !needsSideSelection {
		// Выбор стороны не нужен, проверяем завершение сессии
		// Это может быть только для BO1 (нет выбора сторон)
		if uc.logicService.IsVetoFinished(session, actionsWithNewPick, availableMapsAfterPick) {
			// Для BO1: должна остаться только одна карта
			if session.Type == entities.VetoTypeBo1 && len(availableMapsAfterPick) == 1 {
				session.SelectedMapID = &availableMapsAfterPick[0].ID
				now := time.Now()
				session.FinishedAt = &now
			}
			// Для BO3/BO5: десидер выбирается после выбора сторон, не здесь
			session.Status = entities.VetoStatusFinished
		}
	}
	// Если needsSideSelection == true, НЕ устанавливаем finished, так как выбор стороны еще должен произойти

	// Обновляем сессию
	if err := uc.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	// Получаем обновленную сессию с действиями
	updatedSession, err := uc.sessionRepo.GetByID(session.ID)
	if err != nil {
		return nil, err
	}

	return &PickMapOutput{
		Session: updatedSession,
		Action:  action,
	}, nil
}