package models

import (
	"time"
	"gorm.io/gorm"
)

type MapPoolModel struct {
	ID        uint           `gorm:"primaryKey"`
	GameID    uint           `gorm:"not null;index"`
	UserID    *uint          `gorm:"index"`
	Name      string         `gorm:"not null;size:255"`
	Type      string         `gorm:"not null;size:50"`
	IsSystem  bool           `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Maps      []MapModel     `gorm:"many2many:map_pool_maps;"`
}

func (MapPoolModel) TableName() string {
	return "map_pools"
}
