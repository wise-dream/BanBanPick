package map_pool

import (
	"github.com/bbp/backend/internal/domain/repositories"
)

type DeletePoolUseCase struct {
	mapPoolRepo repositories.MapPoolRepository
}

type DeletePoolInput struct {
	PoolID uint
	UserID uint
}

func NewDeletePoolUseCase(
	mapPoolRepo repositories.MapPoolRepository,
) *DeletePoolUseCase {
	return &DeletePoolUseCase{
		mapPoolRepo: mapPoolRepo,
	}
}

func (uc *DeletePoolUseCase) Execute(input DeletePoolInput) error {
	// Получаем пул
	pool, err := uc.mapPoolRepo.GetByID(input.PoolID)
	if err != nil {
		return err
	}
	if pool == nil {
		return ErrMapPoolNotFound
	}

	// Проверяем, что пользователь является владельцем
	// Пулы теперь всегда привязаны к пользователям, системных пулов нет
	if pool.UserID == nil || *pool.UserID != input.UserID {
		return ErrUnauthorized
	}

	// Удаляем пул
	if err := uc.mapPoolRepo.Delete(input.PoolID); err != nil {
		return err
	}

	return nil
}
