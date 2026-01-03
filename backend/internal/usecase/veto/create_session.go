package veto

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type CreateSessionUseCase struct {
	sessionRepo  repositories.VetoSessionRepository
	mapPoolRepo  repositories.MapPoolRepository
	gameRepo     repositories.GameRepository
	logicService *VetoLogicService
}

type CreateSessionInput struct {
	UserID      *uint
	GameID      uint
	MapPoolID   uint
	Type        entities.VetoType
	TeamAName   string
	TeamBName   string
	TimerSeconds int
}

type CreateSessionOutput struct {
	Session *entities.VetoSession
}

func NewCreateSessionUseCase(
	sessionRepo repositories.VetoSessionRepository,
	mapPoolRepo repositories.MapPoolRepository,
	gameRepo repositories.GameRepository,
	logicService *VetoLogicService,
) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		sessionRepo:  sessionRepo,
		mapPoolRepo:  mapPoolRepo,
		gameRepo:     gameRepo,
		logicService: logicService,
	}
}

func (uc *CreateSessionUseCase) Execute(input CreateSessionInput) (*CreateSessionOutput, error) {
	// Проверяем, что игра существует
	game, err := uc.gameRepo.GetByID(input.GameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ErrGameNotFound
	}

	// Проверяем, что пул карт существует
	mapPool, err := uc.mapPoolRepo.GetByID(input.MapPoolID)
	if err != nil {
		return nil, err
	}
	if mapPool == nil {
		return nil, ErrMapPoolNotFound
	}

	// Проверяем, что пул принадлежит игре
	if mapPool.GameID != input.GameID {
		return nil, ErrInvalidMapPool
	}

	// Генерируем уникальный share token
	shareToken, err := generateShareToken()
	if err != nil {
		return nil, err
	}

	// Создаем сессию
	session := &entities.VetoSession{
		UserID:        input.UserID,
		GameID:        input.GameID,
		MapPoolID:     input.MapPoolID,
		Type:          input.Type,
		Status:        entities.VetoStatusNotStarted,
		TeamAName:     input.TeamAName,
		TeamBName:     input.TeamBName,
		CurrentTeam:   "A", // Команда A начинает первой
		TimerSeconds:  input.TimerSeconds,
		ShareToken:    shareToken,
		Actions:       []entities.VetoAction{},
	}

	// Валидируем сессию
	if err := session.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем сессию
	if err := uc.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return &CreateSessionOutput{
		Session: session,
	}, nil
}

// generateShareToken генерирует уникальный токен для публичного доступа
func generateShareToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}