package models

import (
	"time"
	"gorm.io/gorm"
)

type RoomModel struct {
	ID              uint           `gorm:"primaryKey"`
	OwnerID         uint           `gorm:"not null;index"`
	Name            string         `gorm:"not null;size:255"`
	Code            string         `gorm:"uniqueIndex;not null;size:8"`
	Password        *string        `gorm:"size:255"` // Хеш пароля для приватных комнат (опционально)
	Type            string         `gorm:"not null;size:20"`
	Status          string         `gorm:"not null;size:20;default:'waiting'"`
	GameID          uint           `gorm:"not null;index"`
	MapPoolID       *uint          `gorm:"index"`
	VetoSessionID   *uint       `gorm:"index"`
	MaxParticipants int            `gorm:"default:10"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (RoomModel) TableName() string {
	return "rooms"
}
