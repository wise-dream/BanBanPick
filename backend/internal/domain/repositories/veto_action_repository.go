package repositories

import "github.com/bbp/backend/internal/domain/entities"

type VetoActionRepository interface {
	Create(action *entities.VetoAction) error
	GetBySessionID(sessionID uint) ([]entities.VetoAction, error)
	DeleteBySessionID(sessionID uint) error
	Delete(id uint) error
}