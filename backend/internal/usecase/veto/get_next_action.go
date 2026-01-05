package veto

import (
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetNextActionUseCase struct {
	sessionRepo  repositories.VetoSessionRepository
	mapPoolRepo  repositories.MapPoolRepository
	logicService *VetoLogicService
}

type GetNextActionOutput struct {
	ActionType        NextActionType `json:"action_type"`
	CurrentStep       int            `json:"current_step"`
	CurrentTeam       string         `json:"current_team"`
	CanBan            bool           `json:"can_ban"`
	CanPick           bool           `json:"can_pick"`
	NeedsSideSelection bool          `json:"needs_side_selection"` // Нужен ли выбор стороны после последнего действия
	SideSelectionTeam string         `json:"side_selection_team,omitempty"` // Какая команда должна выбрать сторону
	Message           string         `json:"message,omitempty"`
}

func NewGetNextActionUseCase(
	sessionRepo repositories.VetoSessionRepository,
	mapPoolRepo repositories.MapPoolRepository,
	logicService *VetoLogicService,
) *GetNextActionUseCase {
	return &GetNextActionUseCase{
		sessionRepo:  sessionRepo,
		mapPoolRepo:  mapPoolRepo,
		logicService: logicService,
	}
}

func (uc *GetNextActionUseCase) Execute(sessionID uint) (*GetNextActionOutput, error) {
	// Получаем сессию
	session, err := uc.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Получаем пул карт
	mapPool, err := uc.mapPoolRepo.GetByID(session.MapPoolID)
	if err != nil {
		return nil, err
	}
	if mapPool == nil {
		return nil, ErrMapPoolNotFound
	}

	// Получаем доступные карты
	availableMaps := uc.logicService.GetAvailableMaps(mapPool, session.Actions)
	
	// Определяем текущий шаг
	currentStep := uc.logicService.GetCurrentStep(session.Actions)
	
	// Определяем текущую команду
	currentTeam := uc.logicService.GetCurrentTeam(session.Type, currentStep)
	
	// Определяем тип следующего действия
	nextActionType := uc.logicService.GetNextActionType(session, session.Actions, len(availableMaps))
	
	// Проверяем, нужен ли выбор стороны после последнего действия
	// ВАЖНО: Эта проверка должна быть ПЕРЕД проверкой завершения сессии
	// Потому что даже если сессия "завершена" (все карты пикнуты/забанены),
	// выбор стороны все еще может быть необходим
	needsSideSelection := uc.logicService.NeedsSideSelection(session, session.Actions)
	sideSelectionTeam := ""
	if needsSideSelection && len(session.Actions) > 0 {
		lastAction := session.Actions[len(session.Actions)-1]
		sideSelectionTeam = uc.logicService.GetSideSelectionTeam(session.Type, lastAction.StepNumber)
	}

	// Если нужен выбор стороны, блокируем следующие действия и возвращаем выбор стороны
	// даже если сессия технически "завершена" (все карты выбраны)
	if needsSideSelection {
		return &GetNextActionOutput{
			ActionType:         NextActionTypeBan, // Не используется, но нужен для структуры
			CurrentStep:        currentStep,
			CurrentTeam:        sideSelectionTeam,
			CanBan:             false,
			CanPick:            false,
			NeedsSideSelection: true,
			SideSelectionTeam:  sideSelectionTeam,
			Message:            "Side selection required",
		}, nil
	}

	// Проверяем, завершена ли сессия (только если выбор стороны не нужен)
	if uc.logicService.IsVetoFinished(session, session.Actions, availableMaps) {
		return &GetNextActionOutput{
			ActionType:         NextActionTypeBan,
			CurrentStep:        currentStep,
			CurrentTeam:        currentTeam,
			CanBan:             false,
			CanPick:            false,
			NeedsSideSelection: false,
			SideSelectionTeam:  "",
			Message:            "Veto process is finished",
		}, nil
	}

	canBan := nextActionType == NextActionTypeBan || nextActionType == NextActionTypeBoth
	canPick := nextActionType == NextActionTypePick || nextActionType == NextActionTypeBoth

	return &GetNextActionOutput{
		ActionType:         nextActionType,
		CurrentStep:        currentStep,
		CurrentTeam:        currentTeam,
		CanBan:             canBan,
		CanPick:            canPick,
		NeedsSideSelection: false,
		SideSelectionTeam:  "",
	}, nil
}