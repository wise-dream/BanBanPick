package room

import (
	"time"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/pkg/password"
)

type JoinRoomUseCase struct {
	roomRepo repositories.RoomRepository
}

type JoinRoomInput struct {
	RoomID   *uint  // ID комнаты
	UserID   uint
	Password string // Пароль для приватных комнат
}

type JoinRoomOutput struct {
	Room *entities.Room
}

func NewJoinRoomUseCase(
	roomRepo repositories.RoomRepository,
) *JoinRoomUseCase {
	return &JoinRoomUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *JoinRoomUseCase) Execute(input JoinRoomInput) (*JoinRoomOutput, error) {
	// Получаем комнату по ID
	if input.RoomID == nil {
		return nil, ErrInvalidRoom
	}

	room, err := uc.roomRepo.GetByID(*input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	// Проверяем, что пользователь не уже в комнате
	existingParticipant, err := uc.roomRepo.GetParticipant(room.ID, input.UserID)
	if err != nil {
		return nil, err
	}
	if existingParticipant != nil {
		return nil, ErrAlreadyInRoom
	}

	// Проверяем, что пользователь не в другой комнате
	userRoom, err := uc.roomRepo.GetUserRoom(input.UserID)
	if err != nil {
		return nil, err
	}
	if userRoom != nil && userRoom.ID != room.ID {
		return nil, ErrAlreadyInRoom
	}

	// Проверяем пароль для приватных комнат
	if room.Type == entities.RoomTypePrivate {
		if room.Password == nil || *room.Password == "" {
			// Приватная комната без пароля - разрешаем присоединение
			// (можно использовать для комнат, защищенных только кодом)
		} else {
			// Проверяем пароль
			if input.Password == "" {
				return nil, ErrInvalidCode
			}
			if !password.CheckPassword(*room.Password, input.Password) {
				return nil, ErrInvalidCode
			}
		}
	}

	// Проверяем, можно ли присоединиться
	if !room.CanJoin() {
		return nil, ErrRoomFull
	}

	// Добавляем участника
	participant := &entities.RoomParticipant{
		RoomID:   room.ID,
		UserID:   input.UserID,
		Role:     entities.ParticipantRoleMember,
		JoinedAt: time.Now(),
	}

	if err := uc.roomRepo.AddParticipant(participant); err != nil {
		return nil, err
	}

	// Загружаем обновленную комнату с участниками
	room, err = uc.roomRepo.GetByID(room.ID)
	if err != nil {
		return nil, err
	}

	return &JoinRoomOutput{
		Room: room,
	}, nil
}
