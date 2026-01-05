package sqlite

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/repository/models"
	"gorm.io/gorm"
)

type vetoActionRepository struct {
	db *gorm.DB
}

func NewVetoActionRepository(db *gorm.DB) repositories.VetoActionRepository {
	return &vetoActionRepository{db: db}
}

func (r *vetoActionRepository) Create(action *entities.VetoAction) error {
	model := &models.VetoActionModel{
		VetoSessionID: action.VetoSessionID,
		MapID:         action.MapID,
		Team:          action.Team,
		ActionType:    string(action.ActionType),
		StepNumber:    action.StepNumber,
		SelectedSide:  action.SelectedSide,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	action.ID = model.ID
	action.CreatedAt = model.CreatedAt
	return nil
}

func (r *vetoActionRepository) GetBySessionID(sessionID uint) ([]entities.VetoAction, error) {
	var modelList []models.VetoActionModel
	if err := r.db.Where("veto_session_id = ?", sessionID).Order("step_number ASC").Find(&modelList).Error; err != nil {
		return nil, err
	}

	actions := make([]entities.VetoAction, len(modelList))
	for i, model := range modelList {
		actions[i] = *toVetoActionEntity(&model)
	}

	return actions, nil
}

func (r *vetoActionRepository) DeleteBySessionID(sessionID uint) error {
	return r.db.Where("veto_session_id = ?", sessionID).Delete(&models.VetoActionModel{}).Error
}

func (r *vetoActionRepository) Update(action *entities.VetoAction) error {
	model := &models.VetoActionModel{
		SelectedSide: action.SelectedSide,
	}
	return r.db.Model(&models.VetoActionModel{}).Where("id = ?", action.ID).Update("selected_side", model.SelectedSide).Error
}

func (r *vetoActionRepository) Delete(id uint) error {
	return r.db.Delete(&models.VetoActionModel{}, id).Error
}

func toVetoActionEntity(model *models.VetoActionModel) *entities.VetoAction {
	return &entities.VetoAction{
		ID:            model.ID,
		VetoSessionID: model.VetoSessionID,
		MapID:         model.MapID,
		Team:          model.Team,
		ActionType:    entities.VetoActionType(model.ActionType),
		StepNumber:    model.StepNumber,
		SelectedSide:  model.SelectedSide,
		CreatedAt:     model.CreatedAt,
	}
}