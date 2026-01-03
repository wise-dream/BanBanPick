package dto

import (
	"time"

	"github.com/bbp/backend/internal/domain/entities"
)

// UpdateProfileRequest DTO для обновления профиля
type UpdateProfileRequest struct {
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	Username *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
}

// UserSessionsResponse DTO для ответа с сессиями пользователя
type UserSessionsResponse struct {
	Sessions []VetoSessionResponse `json:"sessions"`
}

// UserRoomsResponse DTO для ответа с комнатами пользователя
type UserRoomsResponse struct {
	Rooms []RoomResponse `json:"rooms"`
}

// RoomResponse DTO для комнаты
type RoomResponse struct {
	ID              uint                 `json:"id"`
	OwnerID         uint                 `json:"owner_id"`
	Name            string               `json:"name"`
	Code            string               `json:"code"`
	Type            string               `json:"type"`
	Status          string               `json:"status"`
	GameID          uint                 `json:"game_id"`
	MapPoolID       *uint                `json:"map_pool_id,omitempty"`
	VetoSessionID   *uint                `json:"veto_session_id,omitempty"`
	MaxParticipants int                  `json:"max_participants"`
	CreatedAt       string               `json:"created_at"`
	UpdatedAt       string               `json:"updated_at"`
	Participants    []RoomParticipantResponse `json:"participants,omitempty"`
}

// RoomParticipantResponse DTO для участника комнаты
type RoomParticipantResponse struct {
	ID       uint   `json:"id"`
	RoomID   uint   `json:"room_id"`
	UserID   uint   `json:"user_id"`
	Username *string `json:"username,omitempty"` // Никнейм пользователя
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}


// ToRoomResponse конвертирует entity Room в RoomResponse
func ToRoomResponse(room *entities.Room) RoomResponse {
	response := RoomResponse{
		ID:              room.ID,
		OwnerID:         room.OwnerID,
		Name:            room.Name,
		Code:            room.Code,
		Type:            string(room.Type),
		Status:          string(room.Status),
		GameID:          room.GameID,
		MapPoolID:       room.MapPoolID,
		VetoSessionID:   room.VetoSessionID,
		MaxParticipants: room.MaxParticipants,
		CreatedAt:       room.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       room.UpdatedAt.Format(time.RFC3339),
	}
	if room.Participants != nil {
		response.Participants = ToRoomParticipantResponseList(room.Participants)
	}
	return response
}

// ToRoomResponseList конвертирует список комнат
func ToRoomResponseList(rooms []entities.Room) []RoomResponse {
	response := make([]RoomResponse, len(rooms))
	for i, room := range rooms {
		response[i] = ToRoomResponse(&room)
	}
	return response
}

// ToRoomParticipantResponse конвертирует entity RoomParticipant в RoomParticipantResponse
func ToRoomParticipantResponse(participant *entities.RoomParticipant) RoomParticipantResponse {
	return RoomParticipantResponse{
		ID:       participant.ID,
		RoomID:   participant.RoomID,
		UserID:   participant.UserID,
		Username: participant.Username,
		Role:     string(participant.Role),
		JoinedAt: participant.JoinedAt.Format(time.RFC3339),
	}
}

// ToRoomParticipantResponseList конвертирует список участников
func ToRoomParticipantResponseList(participants []entities.RoomParticipant) []RoomParticipantResponse {
	response := make([]RoomParticipantResponse, len(participants))
	for i, participant := range participants {
		response[i] = ToRoomParticipantResponse(&participant)
	}
	return response
}