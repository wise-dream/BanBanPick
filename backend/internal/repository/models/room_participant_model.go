package models

import (
	"time"
	"gorm.io/gorm"
)

type RoomParticipantModel struct {
	ID        uint           `gorm:"primaryKey"`
	RoomID    uint           `gorm:"not null;index"`
	UserID    uint           `gorm:"not null;index"`
	Role      string         `gorm:"not null;size:20"`
	JoinedAt  time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RoomParticipantModel) TableName() string {
	return "room_participants"
}
