package repositories

import "github.com/bbp/backend/internal/domain/entities"

type MapRepository interface {
	Create(m *entities.Map) error
	GetByID(id uint) (*entities.Map, error)
	GetByGameID(gameID uint) ([]entities.Map, error)
	GetBySlug(slug string) (*entities.Map, error)
	Update(m *entities.Map) error
	Delete(id uint) error
}
