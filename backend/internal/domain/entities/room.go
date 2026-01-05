package entities

import (
	"errors"
	"time"
)

type RoomType string

const (
	RoomTypePublic  RoomType = "public"
	RoomTypePrivate RoomType = "private"
)

type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusActive   RoomStatus = "active"
	RoomStatusFinished RoomStatus = "finished"
)

type Room struct {
	ID              uint         `json:"id"`
	OwnerID         uint         `json:"owner_id"`
	Name            string       `json:"name"`
	Code            string       `json:"code"`
	Password        *string      `json:"-"` // Хеш пароля (не возвращается в JSON)
	Type            RoomType     `json:"type"`
	Status          RoomStatus   `json:"status"`
	GameID          uint         `json:"game_id"`
	MapPoolID       *uint        `json:"map_pool_id,omitempty"`
	VetoType        *VetoType    `json:"veto_type,omitempty"` // Тип вето (bo1, bo3, bo5)
	VetoSessionID   *uint        `json:"veto_session_id,omitempty"`
	MaxParticipants int          `json:"max_participants"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Participants    []RoomParticipant `json:"participants,omitempty"`
}

// Validate проверяет валидность данных комнаты
func (r *Room) Validate() error {
	if r.OwnerID == 0 {
		return errors.New("owner_id is required")
	}
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) > 255 {
		return errors.New("name must be no more than 255 characters")
	}
	if r.Code == "" {
		return errors.New("code is required")
	}
	if len(r.Code) < 6 || len(r.Code) > 8 {
		return errors.New("code must be between 6 and 8 characters")
	}
	if r.Type != RoomTypePublic && r.Type != RoomTypePrivate {
		return errors.New("invalid room type")
	}
	if r.GameID == 0 {
		return errors.New("game_id is required")
	}
	if r.MaxParticipants < 2 || r.MaxParticipants > 20 {
		return errors.New("max_participants must be between 2 and 20")
	}
	return nil
}

// CanJoin проверяет, можно ли присоединиться к комнате
func (r *Room) CanJoin() bool {
	return r.Status == RoomStatusWaiting && len(r.Participants) < r.MaxParticipants
}

// IsOwner проверяет, является ли пользователь владельцем
func (r *Room) IsOwner(userID uint) bool {
	return r.OwnerID == userID
}
