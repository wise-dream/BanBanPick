package repositories

import "github.com/bbp/backend/internal/domain/entities"

type VetoSessionRepository interface {
	Create(session *entities.VetoSession) error
	GetByID(id uint) (*entities.VetoSession, error)
	GetByShareToken(token string) (*entities.VetoSession, error)
	GetByUserID(userID uint) ([]entities.VetoSession, error)
	Update(session *entities.VetoSession) error
	Delete(id uint) error
}
