package dto

import (
	"time"

	"github.com/bbp/backend/internal/domain/entities"
)

// CreateVetoSessionRequest DTO для создания сессии вето
type CreateVetoSessionRequest struct {
	GameID       uint   `json:"game_id" binding:"required"`
	MapPoolID    uint   `json:"map_pool_id" binding:"required"`
	Type         string `json:"type" binding:"required,oneof=bo1 bo3 bo5"`
	TeamAName    string `json:"team_a_name" binding:"required,min=1,max=100"`
	TeamBName    string `json:"team_b_name" binding:"required,min=1,max=100"`
	TimerSeconds int    `json:"timer_seconds" binding:"min=0,max=300"`
}

// VetoSessionResponse DTO для ответа с сессией
type VetoSessionResponse struct {
	ID            uint                      `json:"id"`
	UserID        *uint                     `json:"user_id,omitempty"`
	GameID        uint                      `json:"game_id"`
	MapPoolID     uint                      `json:"map_pool_id"`
	Type          string                    `json:"type"`
	Status        string                    `json:"status"`
	TeamAName     string                    `json:"team_a_name"`
	TeamBName     string                    `json:"team_b_name"`
	CurrentTeam   string                    `json:"current_team"`
	SelectedMapID *uint                     `json:"selected_map_id,omitempty"`
	SelectedSide  *string                   `json:"selected_side,omitempty"`
	TimerSeconds  int                       `json:"timer_seconds"`
	ShareToken    string                    `json:"share_token"`
	CreatedAt     string                    `json:"created_at"`
	UpdatedAt     string                    `json:"updated_at"`
	FinishedAt    *string                   `json:"finished_at,omitempty"`
	MapPool       *MapPoolResponse          `json:"map_pool,omitempty"`
	Actions       []VetoActionResponse      `json:"actions,omitempty"`
}

// NextActionResponse DTO для следующего действия
type NextActionResponse struct {
	ActionType  string `json:"action_type"`  // "ban", "pick", "both"
	CurrentStep int    `json:"current_step"`
	CurrentTeam string `json:"current_team"` // "A" или "B"
	CanBan      bool   `json:"can_ban"`
	CanPick     bool   `json:"can_pick"`
	Message     string `json:"message,omitempty"`
}

// BanMapRequest DTO для бана карты
type BanMapRequest struct {
	MapID uint `json:"map_id" binding:"required"`
	Team  string `json:"team" binding:"required,oneof=A B"`
}

// PickMapRequest DTO для выбора карты
type PickMapRequest struct {
	MapID uint   `json:"map_id" binding:"required"`
	Team  string `json:"team" binding:"required,oneof=A B"`
}

// SelectSideRequest DTO для выбора стороны
type SelectSideRequest struct {
	Side string `json:"side" binding:"required,oneof=attack defence"`
}

// VetoActionResponse DTO для действия
type VetoActionResponse struct {
	ID            uint   `json:"id"`
	VetoSessionID uint   `json:"veto_session_id"`
	MapID         uint   `json:"map_id"`
	Team          string `json:"team"`
	ActionType    string `json:"action_type"`
	StepNumber    int    `json:"step_number"`
	CreatedAt     string `json:"created_at"`
}

// ToVetoSessionResponse конвертирует entity VetoSession в VetoSessionResponse
func ToVetoSessionResponse(session *entities.VetoSession) VetoSessionResponse {
	response := VetoSessionResponse{
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
		CreatedAt:     session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     session.UpdatedAt.Format(time.RFC3339),
	}

	if session.FinishedAt != nil {
		finishedAt := session.FinishedAt.Format(time.RFC3339)
		response.FinishedAt = &finishedAt
	}

	if session.Actions != nil {
		response.Actions = ToVetoActionResponseList(session.Actions)
	}

	return response
}

// ToVetoSessionResponseList конвертирует список сессий
func ToVetoSessionResponseList(sessions []entities.VetoSession) []VetoSessionResponse {
	response := make([]VetoSessionResponse, len(sessions))
	for i, session := range sessions {
		response[i] = ToVetoSessionResponse(&session)
	}
	return response
}

// ToVetoActionResponse конвертирует entity VetoAction в VetoActionResponse
func ToVetoActionResponse(action *entities.VetoAction) VetoActionResponse {
	return VetoActionResponse{
		ID:            action.ID,
		VetoSessionID: action.VetoSessionID,
		MapID:         action.MapID,
		Team:          action.Team,
		ActionType:    string(action.ActionType),
		StepNumber:    action.StepNumber,
		CreatedAt:     action.CreatedAt.Format(time.RFC3339),
	}
}

// ToVetoActionResponseList конвертирует список действий
func ToVetoActionResponseList(actions []entities.VetoAction) []VetoActionResponse {
	response := make([]VetoActionResponse, len(actions))
	for i, action := range actions {
		response[i] = ToVetoActionResponse(&action)
	}
	return response
}