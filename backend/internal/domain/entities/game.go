package entities

import (
	"errors"
	"time"
)

type Game struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate проверяет валидность данных игры
func (g *Game) Validate() error {
	if g.Name == "" {
		return errors.New("name is required")
	}
	if g.Slug == "" {
		return errors.New("slug is required")
	}
	return nil
}
