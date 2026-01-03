package entities

import (
	"errors"
	"time"
)

type VetoStatus string

const (
	VetoStatusNotStarted VetoStatus = "not_started"
	VetoStatusInProgress VetoStatus = "in_progress"
	VetoStatusFinished   VetoStatus = "finished"
	VetoStatusCancelled  VetoStatus = "cancelled"
)

type VetoType string

const (
	VetoTypeBo1 VetoType = "bo1"
	VetoTypeBo3 VetoType = "bo3"
	VetoTypeBo5 VetoType = "bo5"
)

type VetoSession struct {
	ID            uint        `json:"id"`
	UserID        *uint       `json:"user_id,omitempty"`
	GameID        uint        `json:"game_id"`
	MapPoolID     uint        `json:"map_pool_id"`
	Type          VetoType    `json:"type"`
	Status        VetoStatus  `json:"status"`
	TeamAName     string      `json:"team_a_name"`
	TeamBName     string      `json:"team_b_name"`
	CurrentTeam   string      `json:"current_team"`
	SelectedMapID *uint       `json:"selected_map_id,omitempty"`
	SelectedSide  *string     `json:"selected_side,omitempty"`
	TimerSeconds    int        `json:"timer_seconds"`
	ShareToken    string      `json:"share_token"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	FinishedAt    *time.Time  `json:"finished_at,omitempty"`
	Actions       []VetoAction `json:"actions,omitempty"`
}

// Validate проверяет валидность данных сессии вето
func (vs *VetoSession) Validate() error {
	if vs.GameID == 0 {
		return errors.New("game_id is required")
	}
	if vs.MapPoolID == 0 {
		return errors.New("map_pool_id is required")
	}
	if vs.Type == "" {
		return errors.New("type is required")
	}
	if vs.Type != VetoTypeBo1 && vs.Type != VetoTypeBo3 && vs.Type != VetoTypeBo5 {
		return errors.New("invalid veto type")
	}
	if vs.TeamAName == "" {
		return errors.New("team_a_name is required")
	}
	if vs.TeamBName == "" {
		return errors.New("team_b_name is required")
	}
	if vs.CurrentTeam != "A" && vs.CurrentTeam != "B" {
		return errors.New("current_team must be 'A' or 'B'")
	}
	if vs.TimerSeconds < 0 || vs.TimerSeconds > 300 {
		return errors.New("timer_seconds must be between 0 and 300")
	}
	return nil
}

// CanBan проверяет, можно ли забанить карту
func (vs *VetoSession) CanBan() bool {
	return vs.Status == VetoStatusInProgress && !vs.IsFinished()
}

// CanPick проверяет, можно ли выбрать карту (для Bo3 и Bo5)
func (vs *VetoSession) CanPick() bool {
	return vs.Status == VetoStatusInProgress && !vs.IsFinished() &&
		(vs.Type == VetoTypeBo3 || vs.Type == VetoTypeBo5)
}

// IsFinished проверяет, завершена ли сессия
func (vs *VetoSession) IsFinished() bool {
	return vs.Status == VetoStatusFinished || vs.Status == VetoStatusCancelled
}
