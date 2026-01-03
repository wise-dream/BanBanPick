package repositories

import "github.com/bbp/backend/internal/domain/entities"

type MapPoolRepository interface {
	Create(pool *entities.MapPool) error
	GetByID(id uint) (*entities.MapPool, error)
	GetByGameID(gameID uint) ([]entities.MapPool, error) // Deprecated: используйте GetByGameIDAndUserID
	GetByGameIDAndUserID(gameID, userID uint) ([]entities.MapPool, error)
	GetByUserID(userID uint) ([]entities.MapPool, error)
	GetSystemPools(gameID uint) ([]entities.MapPool, error) // Deprecated: системные пулы больше не используются
	Update(pool *entities.MapPool) error
	Delete(id uint) error
	AddMap(poolID, mapID uint) error
	RemoveMap(poolID, mapID uint) error
}
