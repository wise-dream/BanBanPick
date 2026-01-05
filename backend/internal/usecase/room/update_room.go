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
	MapPoolID     *uint
	VetoType      *entities.VetoType
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

	// Если пытаемся изменить veto_type или map_pool_id, проверяем что сессия не начата
	if (input.VetoType != nil || input.MapPoolID != nil) && room.VetoSessionID != nil {
		// Можно изменить настройки только если сессии еще нет или она не начата
		// Валидация статуса сессии будет в handler, если нужно
		// Пока разрешаем изменение - при создании новой сессии будет использован новый тип/пул
	}

	// Валидируем veto_type, если указан
	if input.VetoType != nil {
		if *input.VetoType != entities.VetoTypeBo1 && 
		   *input.VetoType != entities.VetoTypeBo3 && 
		   *input.VetoType != entities.VetoTypeBo5 {
			return nil, ErrInvalidRoom
		}
	}

	// Обновляем поля
	if input.MapPoolID != nil {
		room.MapPoolID = input.MapPoolID
	}
	if input.VetoType != nil {
		room.VetoType = input.VetoType
	}
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
