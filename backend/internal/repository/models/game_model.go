package models

import (
	"time"
	"gorm.io/gorm"
)

type GameModel struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null;size:100"`
	Slug      string         `gorm:"uniqueIndex;not null;size:50"`
	IsActive  bool           `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (GameModel) TableName() string {
	return "games"
}
