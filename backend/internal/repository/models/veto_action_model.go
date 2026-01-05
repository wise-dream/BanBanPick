package models

import (
	"time"
	"gorm.io/gorm"
)

type VetoActionModel struct {
	ID            uint           `gorm:"primaryKey"`
	VetoSessionID uint           `gorm:"not null;index"`
	MapID         uint           `gorm:"not null;index"`
	Team          string         `gorm:"not null;size:1"`
	ActionType    string         `gorm:"not null;size:10"`
	StepNumber    int            `gorm:"not null"`
	SelectedSide  *string        `gorm:"size:20"` // attack или defence - для действий типа pick
	CreatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (VetoActionModel) TableName() string {
	return "veto_actions"
}
