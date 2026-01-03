package room

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/pkg/password"
)

type CreateRoomUseCase struct {
	roomRepo    repositories.RoomRepository
	gameRepo    repositories.GameRepository
	mapPoolRepo repositories.MapPoolRepository
}

type CreateRoomInput struct {
	OwnerID         uint
	Name            string
	Type            entities.RoomType
	GameID          uint
	MapPoolID       *uint
	MaxParticipants int
	Password        *string // Пароль для приватных комнат (опционально)
}

type CreateRoomOutput struct {
	Room *entities.Room
}

func NewCreateRoomUseCase(
	roomRepo repositories.RoomRepository,
	gameRepo repositories.GameRepository,
	mapPoolRepo repositories.MapPoolRepository,
) *CreateRoomUseCase {
	return &CreateRoomUseCase{
		roomRepo:    roomRepo,
		gameRepo:    gameRepo,
		mapPoolRepo: mapPoolRepo,
	}
}

func (uc *CreateRoomUseCase) Execute(input CreateRoomInput) (*CreateRoomOutput, error) {
	// Проверяем, что игра существует
	game, err := uc.gameRepo.GetByID(input.GameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ErrGameNotFound
	}

	// Проверяем, что пул карт существует (если указан)
	if input.MapPoolID != nil {
		mapPool, err := uc.mapPoolRepo.GetByID(*input.MapPoolID)
		if err != nil {
			return nil, err
		}
		if mapPool == nil {
			return nil, ErrMapPoolNotFound
		}
		if mapPool.GameID != input.GameID {
			return nil, ErrInvalidRoom
		}
	}

	// Генерируем уникальный код комнаты
	code, err := generateRoomCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate room code: %w", err)
	}

	// Устанавливаем значения по умолчанию
	maxParticipants := input.MaxParticipants
	if maxParticipants == 0 {
		maxParticipants = 10 // По умолчанию
	}
	if maxParticipants < 2 {
		maxParticipants = 2
	}
	if maxParticipants > 20 {
		maxParticipants = 20
	}

	// Хешируем пароль, если он указан для приватной комнаты
	var hashedPassword *string
	if input.Type == entities.RoomTypePrivate && input.Password != nil && *input.Password != "" {
		hashed, err := password.HashPassword(*input.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		hashedPassword = &hashed
	}

	// Создаем комнату
	room := &entities.Room{
		OwnerID:         input.OwnerID,
		Name:            input.Name,
		Code:            code,
		Password:        hashedPassword,
		Type:            input.Type,
		Status:          entities.RoomStatusWaiting,
		GameID:          input.GameID,
		MapPoolID:       input.MapPoolID,
		MaxParticipants: maxParticipants,
	}

	// Валидация
	if err := room.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем в БД
	if err := uc.roomRepo.Create(room); err != nil {
		return nil, err
	}

	// Добавляем владельца как участника
	participant := &entities.RoomParticipant{
		RoomID:   room.ID,
		UserID:   input.OwnerID,
		Role:     entities.ParticipantRoleOwner,
		JoinedAt: room.CreatedAt,
	}

	if err := uc.roomRepo.AddParticipant(participant); err != nil {
		// Если не удалось добавить участника, удаляем комнату
		_ = uc.roomRepo.Delete(room.ID)
		return nil, fmt.Errorf("failed to add owner as participant: %w", err)
	}

	// Загружаем участников
	participants, err := uc.roomRepo.GetParticipants(room.ID)
	if err != nil {
		return nil, err
	}
	room.Participants = participants

	return &CreateRoomOutput{
		Room: room,
	}, nil
}

// generateRoomCode генерирует уникальный код комнаты (6-8 символов)
func generateRoomCode() (string, error) {
	bytes := make([]byte, 4) // 4 байта = 8 hex символов
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := hex.EncodeToString(bytes)[:8] // Берем первые 8 символов
	return code, nil
}
