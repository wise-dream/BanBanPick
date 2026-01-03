package room

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type LeaveRoomUseCase struct {
	roomRepo repositories.RoomRepository
}

type LeaveRoomInput struct {
	RoomID uint
	UserID uint
}

type LeaveRoomOutput struct {
	Room *entities.Room
}

func NewLeaveRoomUseCase(
	roomRepo repositories.RoomRepository,
) *LeaveRoomUseCase {
	return &LeaveRoomUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *LeaveRoomUseCase) Execute(input LeaveRoomInput) (*LeaveRoomOutput, error) {
	// Получаем комнату
	room, err := uc.roomRepo.GetByID(input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	// Проверяем, что пользователь является участником
	participant, err := uc.roomRepo.GetParticipant(input.RoomID, input.UserID)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, ErrUnauthorized
	}

	// Если пользователь - владелец, удаляем комнату
	if room.IsOwner(input.UserID) {
		if err := uc.roomRepo.Delete(input.RoomID); err != nil {
			return nil, err
		}
		return &LeaveRoomOutput{
			Room: nil, // Комната удалена
		}, nil
	}

	// Удаляем участника
	if err := uc.roomRepo.RemoveParticipant(input.RoomID, input.UserID); err != nil {
		return nil, err
	}

	// Загружаем обновленную комнату с участниками
	room, err = uc.roomRepo.GetByID(input.RoomID)
	if err != nil {
		return nil, err
	}

	// Если после выхода участника в комнате осталось 0 или 1 участник (только владелец),
	// удаляем комнату, так как для игры нужно минимум 2 участника
	if room != nil && len(room.Participants) <= 1 {
		if err := uc.roomRepo.Delete(input.RoomID); err != nil {
			return nil, err
		}
		return &LeaveRoomOutput{
			Room: nil, // Комната удалена, так как не осталось достаточно участников
		}, nil
	}

	return &LeaveRoomOutput{
		Room: room,
	}, nil
}
