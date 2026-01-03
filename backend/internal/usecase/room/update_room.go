package room

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type UpdateRoomUseCase struct {
	roomRepo repositories.RoomRepository
}

type UpdateRoomInput struct {
	RoomID        uint
	UserID        uint // Для проверки прав
	VetoSessionID *uint
	Status        *entities.RoomStatus
}

type UpdateRoomOutput struct {
	Room *entities.Room
}

func NewUpdateRoomUseCase(
	roomRepo repositories.RoomRepository,
) *UpdateRoomUseCase {
	return &UpdateRoomUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *UpdateRoomUseCase) Execute(input UpdateRoomInput) (*UpdateRoomOutput, error) {
	// Получаем комнату
	room, err := uc.roomRepo.GetByID(input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	// Проверяем, что пользователь является владельцем
	if room.OwnerID != input.UserID {
		return nil, ErrUnauthorized
	}

	// Обновляем поля
	if input.VetoSessionID != nil {
		room.VetoSessionID = input.VetoSessionID
	}
	if input.Status != nil {
		room.Status = *input.Status
	}

	// Сохраняем изменения
	if err := uc.roomRepo.Update(room); err != nil {
		return nil, err
	}

	return &UpdateRoomOutput{
		Room: room,
	}, nil
}
