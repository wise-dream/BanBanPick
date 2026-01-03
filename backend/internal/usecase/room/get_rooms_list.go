package room

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetRoomsListUseCase struct {
	roomRepo repositories.RoomRepository
}

type GetRoomsListInput struct {
	Limit  int
	Offset int
	Type   *string // Опционально: "public" или "private", если nil - возвращаются все
}

type GetRoomsListOutput struct {
	Rooms []entities.Room
	Total int64
}

func NewGetRoomsListUseCase(
	roomRepo repositories.RoomRepository,
) *GetRoomsListUseCase {
	return &GetRoomsListUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *GetRoomsListUseCase) Execute(input GetRoomsListInput) (*GetRoomsListOutput, error) {
	// Устанавливаем значения по умолчанию
	limit := input.Limit
	if limit <= 0 {
		limit = 20 // По умолчанию
	}
	if limit > 100 {
		limit = 100 // Максимум
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	// Создаем фильтр
	filter := &repositories.RoomFilter{}
	if input.Type != nil {
		filter.Type = input.Type
	}

	// Получаем комнаты с фильтром
	rooms, err := uc.roomRepo.GetRooms(filter, limit, offset)
	if err != nil {
		return nil, err
	}

	// Подсчитываем общее количество комнат с фильтром
	total, err := uc.roomRepo.Count(filter)
	if err != nil {
		return nil, err
	}

	return &GetRoomsListOutput{
		Rooms: rooms,
		Total: total,
	}, nil
}
