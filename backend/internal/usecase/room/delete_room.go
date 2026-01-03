package room

import (
	"github.com/bbp/backend/internal/domain/repositories"
)

type DeleteRoomUseCase struct {
	roomRepo repositories.RoomRepository
}

type DeleteRoomInput struct {
	RoomID uint
	UserID uint
}

func NewDeleteRoomUseCase(
	roomRepo repositories.RoomRepository,
) *DeleteRoomUseCase {
	return &DeleteRoomUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *DeleteRoomUseCase) Execute(input DeleteRoomInput) error {
	// Получаем комнату
	room, err := uc.roomRepo.GetByID(input.RoomID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	// Проверяем, что пользователь является владельцем
	if !room.IsOwner(input.UserID) {
		return ErrUnauthorized
	}

	// Удаляем комнату
	if err := uc.roomRepo.Delete(input.RoomID); err != nil {
		return err
	}

	return nil
}
