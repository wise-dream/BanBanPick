package dto

// RoomResponse и RoomParticipantResponse уже определены в user_dto.go
// Здесь добавляем только запросы для Rooms

// CreateRoomRequest DTO для создания комнаты
type CreateRoomRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=255"`
	Type            string  `json:"type" binding:"required,oneof=public private"`
	GameID          uint    `json:"game_id" binding:"required"`
	MapPoolID       *uint   `json:"map_pool_id"`
	MaxParticipants *int    `json:"max_participants" binding:"omitempty,min=2,max=20"`
	Password        *string `json:"password" binding:"omitempty,min=4,max=50"` // Пароль для приватных комнат (опционально)
}

// JoinRoomRequest DTO для присоединения к комнате
type JoinRoomRequest struct {
	Password string `json:"password" binding:"omitempty,min=4,max=50"` // Пароль для приватных комнат
}

// UpdateRoomRequest DTO для обновления комнаты
type UpdateRoomRequest struct {
	VetoSessionID *uint   `json:"veto_session_id"` // ID сессии вето
	Status        *string `json:"status" binding:"omitempty,oneof=waiting active finished"` // Статус комнаты
}
