package veto

import (
	"math/rand"
	"time"

	"github.com/bbp/backend/internal/domain/entities"
)

type NextActionType string

const (
	NextActionTypeBan  NextActionType = "ban"
	NextActionTypePick NextActionType = "pick"
	NextActionTypeBoth NextActionType = "both" // Для Bo5 шаг 4
)

type VetoLogicService struct{}

func NewVetoLogicService() *VetoLogicService {
	return &VetoLogicService{}
}

// GetCurrentStep возвращает текущий шаг процесса (количество выполненных действий + 1)
func (s *VetoLogicService) GetCurrentStep(actions []entities.VetoAction) int {
	return len(actions) + 1
}

// GetCurrentTeam определяет текущую команду на основе шага и типа Bo
func (s *VetoLogicService) GetCurrentTeam(sessionType entities.VetoType, step int) string {
	// Команда A начинает первой
	if step%2 == 1 {
		return "A"
	}
	return "B"
}

// GetNextActionType определяет тип следующего действия
func (s *VetoLogicService) GetNextActionType(
	session *entities.VetoSession,
	actions []entities.VetoAction,
	availableMapsCount int,
) NextActionType {
	step := s.GetCurrentStep(actions)

	switch session.Type {
	case entities.VetoTypeBo1:
		return NextActionTypeBan // Всегда бан

	case entities.VetoTypeBo3:
		// Bo3: ban, ban, pick, ban, ban, pick, ban, ban
		if step == 3 || step == 6 {
			return NextActionTypePick
		}
		return NextActionTypeBan

	case entities.VetoTypeBo5:
		// Bo5: ban, ban, ban, (ban|pick), pick, pick, ban, ban, pick, ban, ban, pick
		if step == 4 {
			return NextActionTypeBoth // Команда B может выбрать ban или pick
		}
		if step == 5 {
			// Определяем на основе предыдущего действия
			if len(actions) > 0 && actions[len(actions)-1].ActionType == entities.VetoActionTypeBan {
				return NextActionTypePick // Если B забанила, A пикает
			}
			return NextActionTypeBan // Если B пикнула, A банит
		}
		if step == 6 || step == 9 || step == 12 {
			return NextActionTypePick
		}
		return NextActionTypeBan

	default:
		return NextActionTypeBan
	}
}

// IsVetoFinished проверяет, завершен ли процесс
func (s *VetoLogicService) IsVetoFinished(
	session *entities.VetoSession,
	actions []entities.VetoAction,
	availableMaps []entities.Map,
) bool {
	switch session.Type {
	case entities.VetoTypeBo1:
		// Bo1 завершается, когда остается 1 карта
		return len(availableMaps) == 1

	case entities.VetoTypeBo3:
		// Bo3: нужно 2 pick (Map 1 и Map 2), десидер выбирается из оставшихся карт
		// НЕ проверяем len(availableMaps) == 1, потому что может остаться больше карт
		// Десидер выбирается автоматически после выбора сторон для обоих пиков
		pickCount := 0
		for _, action := range actions {
			if action.ActionType == entities.VetoActionTypePick {
				pickCount++
			}
		}
		// Сессия завершена когда есть 2 пика (десидер будет выбран автоматически)
		// Проверка выбора сторон для пиков должна быть в select_side, а не здесь
		return pickCount == 2

	case entities.VetoTypeBo5:
		// Bo5: нужно 4 pick (Map 1-4), десидер выбирается из оставшихся карт
		// НЕ проверяем len(availableMaps) == 1, потому что может остаться больше карт
		// Десидер выбирается автоматически после выбора сторон для всех пиков
		pickCount := 0
		for _, action := range actions {
			if action.ActionType == entities.VetoActionTypePick {
				pickCount++
			}
		}
		// Сессия завершена когда есть 4 пика (десидер будет выбран автоматически)
		// Проверка выбора сторон для пиков должна быть в select_side, а не здесь
		return pickCount == 4

	default:
		return false
	}
}

// CanPerformAction проверяет возможность выполнения действия
func (s *VetoLogicService) CanPerformAction(
	session *entities.VetoSession,
	actionType entities.VetoActionType,
	team string,
	actions []entities.VetoAction,
	availableMapsCount int,
) bool {
	// Разрешаем действия если статус in_progress или not_started (для первого действия)
	if session.Status != entities.VetoStatusInProgress && session.Status != entities.VetoStatusNotStarted {
		return false
	}

	step := s.GetCurrentStep(actions)
	currentTeam := s.GetCurrentTeam(session.Type, step)

	// Проверяем, правильная ли команда
	if currentTeam != team {
		return false
	}

	nextActionType := s.GetNextActionType(session, actions, availableMapsCount)

	switch actionType {
	case entities.VetoActionTypeBan:
		return nextActionType == NextActionTypeBan || nextActionType == NextActionTypeBoth
	case entities.VetoActionTypePick:
		return nextActionType == NextActionTypePick || nextActionType == NextActionTypeBoth
	default:
		return false
	}
}

// GetAvailableMaps возвращает доступные карты (не забаненные и не выбранные)
func (s *VetoLogicService) GetAvailableMaps(
	mapPool *entities.MapPool,
	actions []entities.VetoAction,
) []entities.Map {
	// Собираем ID забаненных и выбранных карт
	bannedMapIDs := make(map[uint]bool)
	pickedMapIDs := make(map[uint]bool)

	for _, action := range actions {
		if action.ActionType == entities.VetoActionTypeBan {
			bannedMapIDs[action.MapID] = true
		} else if action.ActionType == entities.VetoActionTypePick {
			pickedMapIDs[action.MapID] = true
		}
	}

	// Фильтруем карты
	availableMaps := []entities.Map{}
	for _, m := range mapPool.Maps {
		if !bannedMapIDs[m.ID] && !pickedMapIDs[m.ID] {
			availableMaps = append(availableMaps, m)
		}
	}

	return availableMaps
}

// GetBannedMaps возвращает список забаненных карт
func (s *VetoLogicService) GetBannedMaps(actions []entities.VetoAction) []uint {
	bannedMapIDs := []uint{}
	for _, action := range actions {
		if action.ActionType == entities.VetoActionTypeBan {
			bannedMapIDs = append(bannedMapIDs, action.MapID)
		}
	}
	return bannedMapIDs
}

// GetPickedMaps возвращает список выбранных карт
func (s *VetoLogicService) GetPickedMaps(actions []entities.VetoAction) []uint {
	pickedMapIDs := []uint{}
	for _, action := range actions {
		if action.ActionType == entities.VetoActionTypePick {
			pickedMapIDs = append(pickedMapIDs, action.MapID)
		}
	}
	return pickedMapIDs
}

// GetSideSelectionTeam определяет какая команда должна выбирать сторону после пика на указанном шаге
// Логика:
// - BO3: после пика на шаге 3 (команда A пикает) → команда B выбирает сторону
//        после пика на шаге 6 (команда B пикает) → команда A выбирает сторону
// - BO5: после каждого пика противоположная команда выбирает сторону
func (s *VetoLogicService) GetSideSelectionTeam(sessionType entities.VetoType, pickStep int) string {
	switch sessionType {
	case entities.VetoTypeBo3:
		if pickStep == 3 {
			// После пика A на шаге 3, B выбирает сторону
			return "B"
		} else if pickStep == 6 {
			// После пика B на шаге 6, A выбирает сторону
			return "A"
		}
	case entities.VetoTypeBo5:
		// После каждого пика противоположная команда выбирает сторону
		// Пик на шаге 5 (A) → B выбирает
		// Пик на шаге 6 (B) → A выбирает
		// Пик на шаге 9 (A) → B выбирает
		// Пик на шаге 12 (B) → A выбирает
		pickTeam := s.GetCurrentTeam(sessionType, pickStep)
		if pickTeam == "A" {
			return "B"
		}
		return "A"
	}
	return ""
}

// NeedsSideSelection проверяет нужен ли выбор стороны после последнего действия
func (s *VetoLogicService) NeedsSideSelection(session *entities.VetoSession, actions []entities.VetoAction) bool {
	if len(actions) == 0 {
		return false
	}

	lastAction := actions[len(actions)-1]
	
	// Выбор стороны нужен только после пика
	if lastAction.ActionType != entities.VetoActionTypePick {
		return false
	}

	// Если сторона уже выбрана, больше не нужна
	if lastAction.SelectedSide != nil {
		return false
	}

	// Проверяем, есть ли команда которая должна выбирать сторону
	shouldSelectTeam := s.GetSideSelectionTeam(session.Type, lastAction.StepNumber)
	return shouldSelectTeam != ""
}

// RandomizeDeciderSide рандомит сторону (attack или defence) для десидера
// Используется для третьей карты в BO3 или пятой карты в BO5
// Возвращает строку "attack" или "defence"
func (s *VetoLogicService) RandomizeDeciderSide() string {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 0 {
		return "attack"
	}
	return "defence"
}

// RandomizeDeciderTeam рандомит команду, которая получает выбранную сторону для десидера
// Используется вместе с RandomizeDeciderSide для полного рандома сторон десидера
// Возвращает "A" или "B"
func (s *VetoLogicService) RandomizeDeciderTeam() string {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 0 {
		return "A"
	}
	return "B"
}