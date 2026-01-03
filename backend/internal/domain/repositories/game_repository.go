package repositories

import "github.com/bbp/backend/internal/domain/entities"

type GameRepository interface {
	Create(game *entities.Game) error
	GetByID(id uint) (*entities.Game, error)
	GetBySlug(slug string) (*entities.Game, error)
	GetAll() ([]entities.Game, error)
	Update(game *entities.Game) error
	Delete(id uint) error
}
