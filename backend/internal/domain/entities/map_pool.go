package entities

import (
	"errors"
	"time"
)

type MapPoolType string

const (
	MapPoolTypeAll         MapPoolType = "all"
	MapPoolTypeCompetitive MapPoolType = "competitive"
	MapPoolTypeCustom      MapPoolType = "custom"
)

type MapPool struct {
	ID        uint         `json:"id"`
	GameID    uint         `json:"game_id"`
	UserID    *uint        `json:"user_id,omitempty"`
	Name      string       `json:"name"`
	Type      MapPoolType  `json:"type"`
	IsSystem  bool         `json:"is_system"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Maps      []Map        `json:"maps,omitempty"`
}

// Validate проверяет валидность данных пула карт
func (mp *MapPool) Validate() error {
	if mp.GameID == 0 {
		return errors.New("game_id is required")
	}
	if mp.Name == "" {
		return errors.New("name is required")
	}
	if mp.Type == "" {
		return errors.New("type is required")
	}
	if mp.Type != MapPoolTypeAll && mp.Type != MapPoolTypeCompetitive && mp.Type != MapPoolTypeCustom {
		return errors.New("invalid map pool type")
	}
	if mp.Type == MapPoolTypeCustom && len(mp.Maps) == 0 {
		return errors.New("custom map pool must have at least one map")
	}
	return nil
}
