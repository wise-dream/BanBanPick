package sqlite

import (
	"errors"
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/repository/models"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *entities.User) error {
	model := &models.UserModel{
		Email:    user.Email,
		Username: user.Username,
		Password: user.Password,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *userRepository) GetByID(id uint) (*entities.User, error) {
	var model models.UserModel
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toUserEntity(&model), nil
}

func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
	var model models.UserModel
	if err := r.db.Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toUserEntity(&model), nil
}

func (r *userRepository) GetByUsername(username string) (*entities.User, error) {
	var model models.UserModel
	if err := r.db.Where("username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toUserEntity(&model), nil
}

func (r *userRepository) Update(user *entities.User) error {
	model := &models.UserModel{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Password: user.Password,
	}

	return r.db.Model(&models.UserModel{}).Where("id = ?", user.ID).Updates(model).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.UserModel{}, id).Error
}

func toUserEntity(model *models.UserModel) *entities.User {
	return &entities.User{
		ID:        model.ID,
		Email:     model.Email,
		Username:  model.Username,
		Password:  model.Password,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
