package sqlite

import (
	"errors"
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/repository/models"
	"gorm.io/gorm"
)

type mapRepository struct {
	db *gorm.DB
}

func NewMapRepository(db *gorm.DB) repositories.MapRepository {
	return &mapRepository{db: db}
}

func (r *mapRepository) Create(m *entities.Map) error {
	model := &models.MapModel{
		GameID:        m.GameID,
		Name:          m.Name,
		Slug:          m.Slug,
		ImageURL:      m.ImageURL,
		IsActive:      m.IsActive,
		IsCompetitive: m.IsCompetitive,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	m.ID = model.ID
	m.CreatedAt = model.CreatedAt
	m.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *mapRepository) GetByID(id uint) (*entities.Map, error) {
	var model models.MapModel
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toMapEntity(&model), nil
}

func (r *mapRepository) GetByGameID(gameID uint) ([]entities.Map, error) {
	var modelList []models.MapModel
	if err := r.db.Where("game_id = ?", gameID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	maps := make([]entities.Map, len(modelList))
	for i, model := range modelList {
		maps[i] = *toMapEntity(&model)
	}

	return maps, nil
}

func (r *mapRepository) GetBySlug(slug string) (*entities.Map, error) {
	var model models.MapModel
	if err := r.db.Where("slug = ?", slug).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toMapEntity(&model), nil
}

func (r *mapRepository) Update(m *entities.Map) error {
	model := &models.MapModel{
		ID:            m.ID,
		GameID:        m.GameID,
		Name:          m.Name,
		Slug:          m.Slug,
		ImageURL:      m.ImageURL,
		IsActive:      m.IsActive,
		IsCompetitive: m.IsCompetitive,
	}

	return r.db.Model(&models.MapModel{}).Where("id = ?", m.ID).Updates(model).Error
}

func (r *mapRepository) Delete(id uint) error {
	return r.db.Delete(&models.MapModel{}, id).Error
}

func toMapEntity(model *models.MapModel) *entities.Map {
	return &entities.Map{
		ID:            model.ID,
		GameID:        model.GameID,
		Name:          model.Name,
		Slug:          model.Slug,
		ImageURL:      model.ImageURL,
		IsActive:      model.IsActive,
		IsCompetitive: model.IsCompetitive,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}
