package repositories

import "github.com/bbp/backend/internal/domain/entities"

type UserRepository interface {
	Create(user *entities.User) error
	GetByID(id uint) (*entities.User, error)
	GetByEmail(email string) (*entities.User, error)
	GetByUsername(username string) (*entities.User, error)
	Update(user *entities.User) error
	Delete(id uint) error
}
