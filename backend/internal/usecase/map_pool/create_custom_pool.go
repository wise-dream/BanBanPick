package map_pool

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type CreateCustomPoolUseCase struct {
	mapPoolRepo repositories.MapPoolRepository
	mapRepo     repositories.MapRepository
	gameRepo    repositories.GameRepository
}

type CreateCustomPoolInput struct {
	UserID  uint
	GameID  uint
	Name    string
	MapIDs []uint
}

type CreateCustomPoolOutput struct {
	Pool *entities.MapPool
}

func NewCreateCustomPoolUseCase(
	mapPoolRepo repositories.MapPoolRepository,
	mapRepo repositories.MapRepository,
	gameRepo repositories.GameRepository,
) *CreateCustomPoolUseCase {
	return &CreateCustomPoolUseCase{
		mapPoolRepo: mapPoolRepo,
		mapRepo:     mapRepo,
		gameRepo:    gameRepo,
	}
}

func (uc *CreateCustomPoolUseCase) Execute(input CreateCustomPoolInput) (*CreateCustomPoolOutput, error) {
	// Проверяем, что игра существует
	game, err := uc.gameRepo.GetByID(input.GameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ErrGameNotFound
	}

	// Проверяем, что все карты существуют и принадлежат игре
	if len(input.MapIDs) == 0 {
		return nil, ErrPoolHasNoMaps
	}

	maps := make([]entities.Map, 0, len(input.MapIDs))
	for _, mapID := range input.MapIDs {
		mapEntity, err := uc.mapRepo.GetByID(mapID)
		if err != nil {
			return nil, err
		}
		if mapEntity == nil {
			return nil, ErrMapNotFound
		}
		if mapEntity.GameID != input.GameID {
			return nil, ErrInvalidMapPool
		}
		maps = append(maps, *mapEntity)
	}

	// Создаем пул
	userID := &input.UserID
	pool := &entities.MapPool{
		GameID:   input.GameID,
		UserID:   userID,
		Name:     input.Name,
		Type:     entities.MapPoolTypeCustom,
		IsSystem: false,
		Maps:     maps,
	}

	// Валидация
	if err := pool.Validate(); err != nil {
		return nil, err
	}

	// Сохраняем в БД
	if err := uc.mapPoolRepo.Create(pool); err != nil {
		return nil, err
	}

	return &CreateCustomPoolOutput{
		Pool: pool,
	}, nil
}
