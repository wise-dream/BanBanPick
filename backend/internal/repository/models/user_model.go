package models

import (
	"time"
	"gorm.io/gorm"
)

type UserModel struct {
	ID        uint           `gorm:"primaryKey"`
	Email     string         `gorm:"uniqueIndex;not null;size:255"`
	Username  string         `gorm:"uniqueIndex;not null;size:100"`
	Password  string         `gorm:"not null;size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
	return "users"
}
