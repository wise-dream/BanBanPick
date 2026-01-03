package map_pool

import (
	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
)

type GetPoolUseCase struct {
	mapPoolRepo repositories.MapPoolRepository
}

type GetPoolInput struct {
	PoolID uint
	UserID uint
}

type GetPoolOutput struct {
	Pool *entities.MapPool
}

func NewGetPoolUseCase(
	mapPoolRepo repositories.MapPoolRepository,
) *GetPoolUseCase {
	return &GetPoolUseCase{
		mapPoolRepo: mapPoolRepo,
	}
}

func (uc *GetPoolUseCase) Execute(input GetPoolInput) (*GetPoolOutput, error) {
	pool, err := uc.mapPoolRepo.GetByID(input.PoolID)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, ErrMapPoolNotFound
	}

	// Системные пулы (UserID == nil) доступны всем пользователям
	// Пользовательские пулы доступны только их владельцам
	if pool.UserID != nil && *pool.UserID != input.UserID {
		return nil, ErrUnauthorized
	}

	return &GetPoolOutput{
		Pool: pool,
	}, nil
}
