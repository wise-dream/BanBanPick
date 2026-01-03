package room

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetRoomUseCase struct {
	roomRepo repositories.RoomRepository
}

type GetRoomOutput struct {
	Room *entities.Room
}

func NewGetRoomUseCase(
	roomRepo repositories.RoomRepository,
) *GetRoomUseCase {
	return &GetRoomUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *GetRoomUseCase) Execute(roomID uint) (*GetRoomOutput, error) {
	room, err := uc.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	return &GetRoomOutput{
		Room: room,
	}, nil
}
