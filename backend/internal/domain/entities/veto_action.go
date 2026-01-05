package entities

import (
	"errors"
	"time"
)

type VetoActionType string

const (
	VetoActionTypeBan  VetoActionType = "ban"
	VetoActionTypePick VetoActionType = "pick"
)

type VetoAction struct {
	ID            uint          `json:"id"`
	VetoSessionID uint          `json:"veto_session_id"`
	MapID         uint          `json:"map_id"`
	Team          string        `json:"team"`
	ActionType    VetoActionType `json:"action_type"`
	StepNumber    int           `json:"step_number"`
	SelectedSide  *string       `json:"selected_side,omitempty"` // "attack" или "defence" - для действий типа pick
	CreatedAt     time.Time     `json:"created_at"`
}

// Validate проверяет валидность данных действия
func (va *VetoAction) Validate() error {
	if va.VetoSessionID == 0 {
		return errors.New("veto_session_id is required")
	}
	if va.MapID == 0 {
		return errors.New("map_id is required")
	}
	if va.Team != "A" && va.Team != "B" {
		return errors.New("team must be 'A' or 'B'")
	}
	if va.ActionType != VetoActionTypeBan && va.ActionType != VetoActionTypePick {
		return errors.New("invalid action type")
	}
	if va.StepNumber < 1 {
		return errors.New("step_number must be at least 1")
	}
	return nil
}
