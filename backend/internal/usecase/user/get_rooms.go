package user

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetRoomsUseCase struct {
	roomRepo repositories.RoomRepository
}

type GetRoomsOutput struct {
	Rooms []entities.Room
}

func NewGetRoomsUseCase(roomRepo repositories.RoomRepository) *GetRoomsUseCase {
	return &GetRoomsUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *GetRoomsUseCase) Execute(userID uint) (*GetRoomsOutput, error) {
	// Получаем комнаты, где пользователь является владельцем
	rooms, err := uc.roomRepo.GetByOwnerID(userID)
	if err != nil {
		return nil, err
	}

	return &GetRoomsOutput{
		Rooms: rooms,
	}, nil
}