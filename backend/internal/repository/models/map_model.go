package models

import (
	"time"
	"gorm.io/gorm"
)

type MapModel struct {
	ID            uint           `gorm:"primaryKey"`
	GameID        uint           `gorm:"not null;index"`
	Name          string         `gorm:"not null;size:100"`
	Slug          string         `gorm:"not null;size:100"`
	ImageURL      string         `gorm:"size:255"`
	IsActive      bool           `gorm:"default:true"`
	IsCompetitive bool           `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (MapModel) TableName() string {
	return "maps"
}
