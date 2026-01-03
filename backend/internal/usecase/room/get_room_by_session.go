package room

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetRoomBySessionUseCase struct {
	roomRepo repositories.RoomRepository
}

type GetRoomBySessionOutput struct {
	Room *entities.Room
}

func NewGetRoomBySessionUseCase(
	roomRepo repositories.RoomRepository,
) *GetRoomBySessionUseCase {
	return &GetRoomBySessionUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *GetRoomBySessionUseCase) Execute(sessionID uint) (*GetRoomBySessionOutput, error) {
	room, err := uc.roomRepo.GetByVetoSessionID(sessionID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	return &GetRoomBySessionOutput{
		Room: room,
	}, nil
}
