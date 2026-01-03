package sqlite

import (
	"errors"
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/repository/models"
	"gorm.io/gorm"
)

type mapPoolRepository struct {
	db *gorm.DB
}

func NewMapPoolRepository(db *gorm.DB) repositories.MapPoolRepository {
	return &mapPoolRepository{db: db}
}

func (r *mapPoolRepository) Create(pool *entities.MapPool) error {
	model := &models.MapPoolModel{
		GameID:   pool.GameID,
		UserID:   pool.UserID,
		Name:     pool.Name,
		Type:     string(pool.Type),
		IsSystem: pool.IsSystem,
	}

	// Преобразуем карты в модели
	mapModels := make([]models.MapModel, len(pool.Maps))
	for i, m := range pool.Maps {
		mapModels[i] = models.MapModel{
			ID: m.ID,
		}
	}
	model.Maps = mapModels

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	pool.ID = model.ID
	pool.CreatedAt = model.CreatedAt
	pool.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *mapPoolRepository) GetByID(id uint) (*entities.MapPool, error) {
	var model models.MapPoolModel
	if err := r.db.Preload("Maps").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toMapPoolEntity(&model), nil
}

func (r *mapPoolRepository) GetByGameID(gameID uint) ([]entities.MapPool, error) {
	var modelList []models.MapPoolModel
	if err := r.db.Preload("Maps").Where("game_id = ?", gameID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	pools := make([]entities.MapPool, len(modelList))
	for i, model := range modelList {
		pools[i] = *toMapPoolEntity(&model)
	}

	return pools, nil
}

func (r *mapPoolRepository) GetByGameIDAndUserID(gameID, userID uint) ([]entities.MapPool, error) {
	var modelList []models.MapPoolModel
	// Получаем системные пулы (user_id IS NULL) + пулы пользователя
	if err := r.db.Preload("Maps").Where("game_id = ? AND (user_id = ? OR user_id IS NULL)", gameID, userID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	pools := make([]entities.MapPool, len(modelList))
	for i, model := range modelList {
		pools[i] = *toMapPoolEntity(&model)
	}

	return pools, nil
}

func (r *mapPoolRepository) GetByUserID(userID uint) ([]entities.MapPool, error) {
	var modelList []models.MapPoolModel
	if err := r.db.Preload("Maps").Where("user_id = ?", userID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	pools := make([]entities.MapPool, len(modelList))
	for i, model := range modelList {
		pools[i] = *toMapPoolEntity(&model)
	}

	return pools, nil
}

func (r *mapPoolRepository) GetSystemPools(gameID uint) ([]entities.MapPool, error) {
	var modelList []models.MapPoolModel
	if err := r.db.Preload("Maps").Where("game_id = ? AND is_system = ?", gameID, true).Find(&modelList).Error; err != nil {
		return nil, err
	}

	pools := make([]entities.MapPool, len(modelList))
	for i, model := range modelList {
		pools[i] = *toMapPoolEntity(&model)
	}

	return pools, nil
}

func (r *mapPoolRepository) Update(pool *entities.MapPool) error {
	model := &models.MapPoolModel{
		ID:       pool.ID,
		GameID:   pool.GameID,
		UserID:   pool.UserID,
		Name:     pool.Name,
		Type:     string(pool.Type),
		IsSystem: pool.IsSystem,
	}

	return r.db.Model(&models.MapPoolModel{}).Where("id = ?", pool.ID).Updates(model).Error
}

func (r *mapPoolRepository) Delete(id uint) error {
	return r.db.Delete(&models.MapPoolModel{}, id).Error
}

func (r *mapPoolRepository) AddMap(poolID, mapID uint) error {
	var pool models.MapPoolModel
	if err := r.db.First(&pool, poolID).Error; err != nil {
		return err
	}

	var mapModel models.MapModel
	if err := r.db.First(&mapModel, mapID).Error; err != nil {
		return err
	}

	return r.db.Model(&pool).Association("Maps").Append(&mapModel)
}

func (r *mapPoolRepository) RemoveMap(poolID, mapID uint) error {
	var pool models.MapPoolModel
	if err := r.db.First(&pool, poolID).Error; err != nil {
		return err
	}

	var mapModel models.MapModel
	if err := r.db.First(&mapModel, mapID).Error; err != nil {
		return err
	}

	return r.db.Model(&pool).Association("Maps").Delete(&mapModel)
}

func toMapPoolEntity(model *models.MapPoolModel) *entities.MapPool {
	maps := make([]entities.Map, len(model.Maps))
	for i, m := range model.Maps {
		maps[i] = entities.Map{
			ID:            m.ID,
			GameID:        m.GameID,
			Name:          m.Name,
			Slug:          m.Slug,
			ImageURL:      m.ImageURL,
			IsActive:      m.IsActive,
			IsCompetitive: m.IsCompetitive,
			CreatedAt:     m.CreatedAt,
			UpdatedAt:     m.UpdatedAt,
		}
	}

	return &entities.MapPool{
		ID:        model.ID,
		GameID:    model.GameID,
		UserID:    model.UserID,
		Name:      model.Name,
		Type:      entities.MapPoolType(model.Type),
		IsSystem:  model.IsSystem,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Maps:      maps,
	}
}
