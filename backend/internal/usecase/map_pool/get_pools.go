package map_pool

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetPoolsUseCase struct {
	mapPoolRepo repositories.MapPoolRepository
	gameRepo    repositories.GameRepository
}

type GetPoolsInput struct {
	GameID uint
	UserID uint
}

type GetPoolsOutput struct {
	Pools []entities.MapPool
}

func NewGetPoolsUseCase(
	mapPoolRepo repositories.MapPoolRepository,
	gameRepo repositories.GameRepository,
) *GetPoolsUseCase {
	return &GetPoolsUseCase{
		mapPoolRepo: mapPoolRepo,
		gameRepo:    gameRepo,
	}
}

func (uc *GetPoolsUseCase) Execute(input GetPoolsInput) (*GetPoolsOutput, error) {
	// Проверяем, что игра существует
	game, err := uc.gameRepo.GetByID(input.GameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, ErrGameNotFound
	}

	// Получаем системные пулы (user_id IS NULL) + пулы текущего пользователя для игры
	pools, err := uc.mapPoolRepo.GetByGameIDAndUserID(input.GameID, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetPoolsOutput{
		Pools: pools,
	}, nil
}
