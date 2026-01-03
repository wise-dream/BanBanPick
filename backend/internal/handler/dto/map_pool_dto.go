package dto

import (
	"time"

	"github.com/bbp/backend/internal/domain/entities"
)

// MapPoolResponse DTO для ответа с пулом карт
type MapPoolResponse struct {
	ID        uint          `json:"id"`
	GameID    uint          `json:"game_id"`
	UserID    *uint         `json:"user_id,omitempty"`
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	IsSystem  bool          `json:"is_system"`
	Maps      []MapResponse `json:"maps"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// CreateCustomMapPoolRequest DTO для создания кастомного пула
type CreateCustomMapPoolRequest struct {
	Name   string `json:"name" binding:"required,min=1,max=255"`
	MapIDs []uint `json:"map_ids" binding:"required,min=1"`
}

// MapResponse DTO для карты
type MapResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	ImageURL      string `json:"image_url"`
	IsCompetitive bool   `json:"is_competitive"`
}

// ToMapPoolResponse конвертирует entity MapPool в MapPoolResponse
func ToMapPoolResponse(pool *entities.MapPool) MapPoolResponse {
	maps := make([]MapResponse, len(pool.Maps))
	for i, m := range pool.Maps {
		maps[i] = MapResponse{
			ID:            m.ID,
			Name:          m.Name,
			Slug:          m.Slug,
			ImageURL:      m.ImageURL,
			IsCompetitive: m.IsCompetitive,
		}
	}

	return MapPoolResponse{
		ID:        pool.ID,
		GameID:    pool.GameID,
		UserID:    pool.UserID,
		Name:      pool.Name,
		Type:      string(pool.Type),
		IsSystem:  pool.IsSystem,
		Maps:      maps,
		CreatedAt: pool.CreatedAt.Format(time.RFC3339),
		UpdatedAt: pool.UpdatedAt.Format(time.RFC3339),
	}
}

// ToMapPoolResponseList конвертирует список пулов
func ToMapPoolResponseList(pools []entities.MapPool) []MapPoolResponse {
	response := make([]MapPoolResponse, len(pools))
	for i, pool := range pools {
		response[i] = ToMapPoolResponse(&pool)
	}
	return response
}
