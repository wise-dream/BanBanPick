package veto

import (
	"math/rand"
	"time"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type SelectSideUseCase struct {
	sessionRepo  repositories.VetoSessionRepository
	actionRepo   repositories.VetoActionRepository
	mapPoolRepo  repositories.MapPoolRepository
	logicService *VetoLogicService
}

type SelectSideInput struct {
	SessionID uint
	Side      string // "attack" или "defence"
	Team      string // Команда, выбирающая сторону ("A" или "B")
}

type SelectSideOutput struct {
	Session *entities.VetoSession
	Action  *entities.VetoAction
}

func NewSelectSideUseCase(
	sessionRepo repositories.VetoSessionRepository,
	actionRepo repositories.VetoActionRepository,
	mapPoolRepo repositories.MapPoolRepository,
	logicService *VetoLogicService,
) *SelectSideUseCase {
	return &SelectSideUseCase{
		sessionRepo:  sessionRepo,
		actionRepo:   actionRepo,
		mapPoolRepo:  mapPoolRepo,
		logicService: logicService,
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

	// Проверяем, что сессия в процессе или завершена
	if session.Status == entities.VetoStatusCancelled {
		return nil, ErrSessionFinished
	}

	// Валидируем сторону
	if input.Side != "attack" && input.Side != "defence" {
		return nil, ErrInvalidAction
	}

	// Валидируем команду
	if input.Team != "A" && input.Team != "B" {
		return nil, ErrInvalidAction
	}

	// Проверяем, что последнее действие - это pick
	if len(session.Actions) == 0 {
		return nil, ErrInvalidAction
	}

	lastAction := session.Actions[len(session.Actions)-1]
	if lastAction.ActionType != entities.VetoActionTypePick {
		return nil, ErrInvalidAction
	}

	// Проверяем, что для этого пика еще не выбрана сторона
	if lastAction.SelectedSide != nil {
		return nil, ErrInvalidAction // Сторона уже выбрана
	}

	// Определяем какая команда должна выбирать сторону после этого пика
	shouldSelectTeam := uc.logicService.GetSideSelectionTeam(session.Type, lastAction.StepNumber)
	if shouldSelectTeam != input.Team {
		return nil, ErrNotYourTurn
	}

	// Обновляем действие - добавляем выбор стороны
	lastAction.SelectedSide = &input.Side
	if err := uc.actionRepo.Update(&lastAction); err != nil {
		return nil, err
	}

	// Получаем обновленную сессию с действиями (чтобы получить обновленное действие с selected_side)
	updatedSession, err := uc.sessionRepo.GetByID(session.ID)
	if err != nil {
		return nil, err
	}

	// Проверяем, завершена ли сессия после выбора стороны
	// Если выбор стороны был последним шагом перед завершением, нужно проверить и установить finished
	if updatedSession.Status != entities.VetoStatusFinished {
		// Получаем пул карт для проверки завершения
		mapPool, err := uc.mapPoolRepo.GetByID(updatedSession.MapPoolID)
		if err == nil && mapPool != nil {
			availableMaps := uc.logicService.GetAvailableMaps(mapPool, updatedSession.Actions)
			
			// Проверяем, завершена ли сессия после выбора стороны
			if uc.logicService.IsVetoFinished(updatedSession, updatedSession.Actions, availableMaps) {
				// Для BO3/BO5: выбираем десидер из оставшихся карт (может быть больше одной)
				// Для BO1: должна остаться только одна карта
				if updatedSession.Type == entities.VetoTypeBo1 {
					// BO1: должна остаться только одна карта
					if len(availableMaps) == 1 {
						updatedSession.SelectedMapID = &availableMaps[0].ID
					}
				} else if updatedSession.Type == entities.VetoTypeBo3 || updatedSession.Type == entities.VetoTypeBo5 {
					// BO3/BO5: выбираем десидер случайно из оставшихся карт
					if len(availableMaps) > 0 {
						// Рандомим индекс карты для десидера
						rand.Seed(time.Now().UnixNano())
						randomIndex := rand.Intn(len(availableMaps))
						deciderMap := availableMaps[randomIndex]
						updatedSession.SelectedMapID = &deciderMap.ID
						
						// ВАЖНО: Рандомим сторону для десидера (третья карта в BO3, пятая в BO5)
						// Это происходит автоматически после выбора последней стороны командой
						randomSide := uc.logicService.RandomizeDeciderSide()
						updatedSession.SelectedSide = &randomSide
					}
				}
				
				// Устанавливаем время завершения
				if updatedSession.SelectedMapID != nil {
					now := time.Now()
					updatedSession.FinishedAt = &now
				}
				updatedSession.Status = entities.VetoStatusFinished
				
				// Обновляем сессию в БД
				if err := uc.sessionRepo.Update(updatedSession); err != nil {
					return nil, err
				}
				
				// Перезагружаем сессию чтобы получить актуальное состояние
				updatedSession, err = uc.sessionRepo.GetByID(session.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Находим обновленное действие в загруженной сессии
	var updatedAction *entities.VetoAction
	if len(updatedSession.Actions) > 0 {
		// Ищем действие по ID (последнее действие должно быть с обновленным selected_side)
		for i := len(updatedSession.Actions) - 1; i >= 0; i-- {
			if updatedSession.Actions[i].ID == lastAction.ID {
				updatedAction = &updatedSession.Actions[i]
				break
			}
		}
	}
	
	// Если не нашли, используем последнее действие (должно быть обновленным)
	if updatedAction == nil && len(updatedSession.Actions) > 0 {
		updatedAction = &updatedSession.Actions[len(updatedSession.Actions)-1]
	}

	return &SelectSideOutput{
		Session: updatedSession,
		Action:  updatedAction,
	}, nil
}