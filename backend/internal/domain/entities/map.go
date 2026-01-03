package entities

import (
	"errors"
	"time"
)

type Map struct {
	ID            uint      `json:"id"`
	GameID        uint      `json:"game_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	ImageURL      string    `json:"image_url"`
	IsActive      bool      `json:"is_active"`
	IsCompetitive bool      `json:"is_competitive"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate проверяет валидность данных карты
func (m *Map) Validate() error {
	if m.GameID == 0 {
		return errors.New("game_id is required")
	}
	if m.Name == "" {
		return errors.New("name is required")
	}
	if m.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}
