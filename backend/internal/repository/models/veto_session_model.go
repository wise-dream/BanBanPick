package models

import (
	"time"
	"gorm.io/gorm"
)

type VetoSessionModel struct {
	ID            uint           `gorm:"primaryKey"`
	UserID        *uint          `gorm:"index"`
	GameID        uint           `gorm:"not null;index"`
	MapPoolID     uint           `gorm:"not null;index"`
	Type          string         `gorm:"not null;size:10"`
	Status        string         `gorm:"not null;size:20"`
	TeamAName     string         `gorm:"not null;size:100"`
	TeamBName     string         `gorm:"not null;size:100"`
	CurrentTeam   string         `gorm:"not null;size:1"`
	SelectedMapID *uint          `gorm:"index"`
	SelectedSide  *string        `gorm:"size:20"`
	TimerSeconds  int            `gorm:"default:0"`
	ShareToken    string         `gorm:"uniqueIndex;not null;size:64"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	FinishedAt    *time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (VetoSessionModel) TableName() string {
	return "veto_sessions"
}
