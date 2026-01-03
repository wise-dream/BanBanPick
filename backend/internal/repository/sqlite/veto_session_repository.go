package sqlite

import (
	"errors"
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/repository/models"
	"gorm.io/gorm"
)

type vetoSessionRepository struct {
	db *gorm.DB
}

func NewVetoSessionRepository(db *gorm.DB) repositories.VetoSessionRepository {
	return &vetoSessionRepository{db: db}
}

func (r *vetoSessionRepository) Create(session *entities.VetoSession) error {
	model := &models.VetoSessionModel{
		UserID:        session.UserID,
		GameID:        session.GameID,
		MapPoolID:     session.MapPoolID,
		Type:          string(session.Type),
		Status:        string(session.Status),
		TeamAName:     session.TeamAName,
		TeamBName:     session.TeamBName,
		CurrentTeam:   session.CurrentTeam,
		SelectedMapID: session.SelectedMapID,
		SelectedSide:  session.SelectedSide,
		TimerSeconds:  session.TimerSeconds,
		ShareToken:    session.ShareToken,
		FinishedAt:    session.FinishedAt,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	session.ID = model.ID
	session.CreatedAt = model.CreatedAt
	session.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *vetoSessionRepository) GetByID(id uint) (*entities.VetoSession, error) {
	var model models.VetoSessionModel
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	session := toVetoSessionEntity(&model)
	
	// Загружаем Actions отдельно
	var actionModels []models.VetoActionModel
	if err := r.db.Where("veto_session_id = ?", id).Order("step_number ASC").Find(&actionModels).Error; err != nil {
		return nil, err
	}
	
	actions := make([]entities.VetoAction, len(actionModels))
	for i, actionModel := range actionModels {
		actions[i] = entities.VetoAction{
			ID:            actionModel.ID,
			VetoSessionID: actionModel.VetoSessionID,
			MapID:         actionModel.MapID,
			Team:          actionModel.Team,
			ActionType:    entities.VetoActionType(actionModel.ActionType),
			StepNumber:    actionModel.StepNumber,
			CreatedAt:     actionModel.CreatedAt,
		}
	}
	session.Actions = actions

	return session, nil
}

func (r *vetoSessionRepository) GetByShareToken(token string) (*entities.VetoSession, error) {
	var model models.VetoSessionModel
	if err := r.db.Where("share_token = ?", token).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	session := toVetoSessionEntity(&model)
	
	// Загружаем Actions отдельно
	var actionModels []models.VetoActionModel
	if err := r.db.Where("veto_session_id = ?", model.ID).Order("step_number ASC").Find(&actionModels).Error; err != nil {
		return nil, err
	}
	
	actions := make([]entities.VetoAction, len(actionModels))
	for i, actionModel := range actionModels {
		actions[i] = entities.VetoAction{
			ID:            actionModel.ID,
			VetoSessionID: actionModel.VetoSessionID,
			MapID:         actionModel.MapID,
			Team:          actionModel.Team,
			ActionType:    entities.VetoActionType(actionModel.ActionType),
			StepNumber:    actionModel.StepNumber,
			CreatedAt:     actionModel.CreatedAt,
		}
	}
	session.Actions = actions

	return session, nil
}

func (r *vetoSessionRepository) GetByUserID(userID uint) ([]entities.VetoSession, error) {
	var modelList []models.VetoSessionModel
	if err := r.db.Where("user_id = ?", userID).Find(&modelList).Error; err != nil {
		return nil, err
	}

	sessions := make([]entities.VetoSession, len(modelList))
	for i, model := range modelList {
		session := toVetoSessionEntity(&model)
		
		// Загружаем Actions для каждой сессии
		var actionModels []models.VetoActionModel
		if err := r.db.Where("veto_session_id = ?", model.ID).Order("step_number ASC").Find(&actionModels).Error; err != nil {
			return nil, err
		}
		
		actions := make([]entities.VetoAction, len(actionModels))
		for j, actionModel := range actionModels {
			actions[j] = entities.VetoAction{
				ID:            actionModel.ID,
				VetoSessionID: actionModel.VetoSessionID,
				MapID:         actionModel.MapID,
				Team:          actionModel.Team,
				ActionType:    entities.VetoActionType(actionModel.ActionType),
				StepNumber:    actionModel.StepNumber,
				CreatedAt:     actionModel.CreatedAt,
			}
		}
		session.Actions = actions
		sessions[i] = *session
	}

	return sessions, nil
}

func (r *vetoSessionRepository) Update(session *entities.VetoSession) error {
	model := &models.VetoSessionModel{
		ID:            session.ID,
		UserID:        session.UserID,
		GameID:        session.GameID,
		MapPoolID:     session.MapPoolID,
		Type:          string(session.Type),
		Status:        string(session.Status),
		TeamAName:     session.TeamAName,
		TeamBName:     session.TeamBName,
		CurrentTeam:   session.CurrentTeam,
		SelectedMapID: session.SelectedMapID,
		SelectedSide:  session.SelectedSide,
		TimerSeconds:  session.TimerSeconds,
		ShareToken:    session.ShareToken,
		FinishedAt:    session.FinishedAt,
	}

	return r.db.Model(&models.VetoSessionModel{}).Where("id = ?", session.ID).Updates(model).Error
}

func (r *vetoSessionRepository) Delete(id uint) error {
	return r.db.Delete(&models.VetoSessionModel{}, id).Error
}

func toVetoSessionEntity(model *models.VetoSessionModel) *entities.VetoSession {
	return &entities.VetoSession{
		ID:            model.ID,
		UserID:        model.UserID,
		GameID:        model.GameID,
		MapPoolID:     model.MapPoolID,
		Type:          entities.VetoType(model.Type),
		Status:        entities.VetoStatus(model.Status),
		TeamAName:     model.TeamAName,
		TeamBName:     model.TeamBName,
		CurrentTeam:   model.CurrentTeam,
		SelectedMapID: model.SelectedMapID,
		SelectedSide:  model.SelectedSide,
		TimerSeconds:  model.TimerSeconds,
		ShareToken:    model.ShareToken,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		FinishedAt:    model.FinishedAt,
		Actions:       []entities.VetoAction{},
	}
}
