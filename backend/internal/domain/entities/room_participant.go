package entities

import (
	"errors"
	"time"
)

type ParticipantRole string

const (
	ParticipantRoleOwner  ParticipantRole = "owner"
	ParticipantRoleMember ParticipantRole = "member"
)

type RoomParticipant struct {
	ID       uint             `json:"id"`
	RoomID   uint             `json:"room_id"`
	UserID   uint             `json:"user_id"`
	Username *string          `json:"username,omitempty"` // Никнейм пользователя (загружается через JOIN)
	Role     ParticipantRole  `json:"role"`
	JoinedAt time.Time        `json:"joined_at"`
}

// Validate проверяет валидность данных участника
func (rp *RoomParticipant) Validate() error {
	if rp.RoomID == 0 {
		return errors.New("room_id is required")
	}
	if rp.UserID == 0 {
		return errors.New("user_id is required")
	}
	if rp.Role != ParticipantRoleOwner && rp.Role != ParticipantRoleMember {
		return errors.New("invalid participant role")
	}
	return nil
}

// IsOwner проверяет, является ли участник владельцем
func (rp *RoomParticipant) IsOwner() bool {
	return rp.Role == ParticipantRoleOwner
}
