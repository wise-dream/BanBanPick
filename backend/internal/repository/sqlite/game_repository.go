package sqlite

import (
	"errors"
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/repository/models"
	"gorm.io/gorm"
)

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) repositories.GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(game *entities.Game) error {
	model := &models.GameModel{
		Name:     game.Name,
		Slug:     game.Slug,
		IsActive: game.IsActive,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	game.ID = model.ID
	game.CreatedAt = model.CreatedAt
	game.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *gameRepository) GetByID(id uint) (*entities.Game, error) {
	var model models.GameModel
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toGameEntity(&model), nil
}

func (r *gameRepository) GetBySlug(slug string) (*entities.Game, error) {
	var model models.GameModel
	if err := r.db.Where("slug = ?", slug).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toGameEntity(&model), nil
}

func (r *gameRepository) GetAll() ([]entities.Game, error) {
	var modelList []models.GameModel
	if err := r.db.Find(&modelList).Error; err != nil {
		return nil, err
	}

	games := make([]entities.Game, len(modelList))
	for i, model := range modelList {
		games[i] = *toGameEntity(&model)
	}

	return games, nil
}

func (r *gameRepository) Update(game *entities.Game) error {
	model := &models.GameModel{
		ID:       game.ID,
		Name:     game.Name,
		Slug:     game.Slug,
		IsActive: game.IsActive,
	}

	return r.db.Model(&models.GameModel{}).Where("id = ?", game.ID).Updates(model).Error
}

func (r *gameRepository) Delete(id uint) error {
	return r.db.Delete(&models.GameModel{}, id).Error
}

func toGameEntity(model *models.GameModel) *entities.Game {
	return &entities.Game{
		ID:        model.ID,
		Name:      model.Name,
		Slug:      model.Slug,
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
