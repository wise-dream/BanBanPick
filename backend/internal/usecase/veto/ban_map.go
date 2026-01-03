package veto

import (
	"time"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type BanMapUseCase struct {
	sessionRepo    repositories.VetoSessionRepository
	actionRepo     repositories.VetoActionRepository
	mapRepo        repositories.MapRepository
	mapPoolRepo    repositories.MapPoolRepository
	logicService   *VetoLogicService
}

type BanMapInput struct {
	SessionID uint
	MapID     uint
	Team      string
}

type BanMapOutput struct {
	Session *entities.VetoSession
	Action  *entities.VetoAction
}

func NewBanMapUseCase(
	sessionRepo repositories.VetoSessionRepository,
	actionRepo repositories.VetoActionRepository,
	mapRepo repositories.MapRepository,
	mapPoolRepo repositories.MapPoolRepository,
	logicService *VetoLogicService,
) *BanMapUseCase {
	return &BanMapUseCase{
		sessionRepo:  sessionRepo,
		actionRepo:   actionRepo,
		mapRepo:      mapRepo,
		mapPoolRepo:  mapPoolRepo,
		logicService: logicService,
	}
}

func (uc *BanMapUseCase) Execute(input BanMapInput) (*BanMapOutput, error) {
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

	// Проверяем, что карта доступна (не забанена и не выбрана)
	mapAvailable := false
	for _, m := range availableMaps {
		if m.ID == input.MapID {
			mapAvailable = true
			break
		}
	}
	if !mapAvailable {
		return nil, ErrMapAlreadyBanned
	}

	// Определяем текущий шаг
	currentStep := uc.logicService.GetCurrentStep(session.Actions)

	// Проверяем возможность выполнения действия (теперь статус уже in_progress)
	if !uc.logicService.CanPerformAction(session, entities.VetoActionTypeBan, input.Team, session.Actions, len(availableMaps)) {
		return nil, ErrNotYourTurn
	}

	// Создаем действие
	action := &entities.VetoAction{
		VetoSessionID: session.ID,
		MapID:         input.MapID,
		Team:          input.Team,
		ActionType:    entities.VetoActionTypeBan,
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

	// Проверяем, завершена ли сессия
	availableMapsAfterBan := uc.logicService.GetAvailableMaps(mapPool, append(session.Actions, *action))
	if uc.logicService.IsVetoFinished(session, append(session.Actions, *action), availableMapsAfterBan) {
		// Для Bo1 автоматически выбираем последнюю карту
		if session.Type == entities.VetoTypeBo1 && len(availableMapsAfterBan) == 1 {
			session.SelectedMapID = &availableMapsAfterBan[0].ID
			now := time.Now()
			session.FinishedAt = &now
		}
		session.Status = entities.VetoStatusFinished
	}

	// Обновляем сессию
	if err := uc.sessionRepo.Update(session); err != nil {
		return nil, err
	}

	// Получаем обновленную сессию с действиями
	updatedSession, err := uc.sessionRepo.GetByID(session.ID)
	if err != nil {
		return nil, err
	}

	return &BanMapOutput{
		Session: updatedSession,
		Action:  action,
	}, nil
}