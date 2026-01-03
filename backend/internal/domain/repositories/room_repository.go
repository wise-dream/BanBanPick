package repositories

import "github.com/bbp/backend/internal/domain/entities"

type RoomFilter struct {
	Type   *string // Опционально: "public" или "private"
	Status *string // Опционально: "waiting", "active", "finished"
}

type RoomRepository interface {
	Create(room *entities.Room) error
	GetByID(id uint) (*entities.Room, error)
	GetByCode(code string) (*entities.Room, error)
	GetByOwnerID(ownerID uint) ([]entities.Room, error)
	GetPublicRooms(limit, offset int) ([]entities.Room, error)
	GetRooms(filter *RoomFilter, limit, offset int) ([]entities.Room, error)
	Update(room *entities.Room) error
	Delete(id uint) error
	AddParticipant(participant *entities.RoomParticipant) error
	RemoveParticipant(roomID, userID uint) error
	GetParticipants(roomID uint) ([]entities.RoomParticipant, error)
	GetParticipant(roomID, userID uint) (*entities.RoomParticipant, error)
	// Получение комнаты, в которой участвует пользователь
	GetUserRoom(userID uint) (*entities.Room, error)
	// Получение комнаты по veto_session_id
	GetByVetoSessionID(sessionID uint) (*entities.Room, error)
	// Подсчет количества комнат с фильтром
	Count(filter *RoomFilter) (int64, error)
}
