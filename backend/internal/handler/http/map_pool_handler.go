package http

import (
	"net/http"
	"strconv"

	"github.com/bbp/backend/internal/handler/dto"
	"github.com/bbp/backend/internal/middleware"
	"github.com/bbp/backend/internal/usecase/map_pool"
	"github.com/gin-gonic/gin"
)

type MapPoolHandler struct {
	getPoolsUseCase        *map_pool.GetPoolsUseCase
	getPoolUseCase         *map_pool.GetPoolUseCase
	createCustomPoolUseCase *map_pool.CreateCustomPoolUseCase
	deletePoolUseCase      *map_pool.DeletePoolUseCase
}

func NewMapPoolHandler(
	getPoolsUseCase *map_pool.GetPoolsUseCase,
	getPoolUseCase *map_pool.GetPoolUseCase,
	createCustomPoolUseCase *map_pool.CreateCustomPoolUseCase,
	deletePoolUseCase *map_pool.DeletePoolUseCase,
) *MapPoolHandler {
	return &MapPoolHandler{
		getPoolsUseCase:         getPoolsUseCase,
		getPoolUseCase:          getPoolUseCase,
		createCustomPoolUseCase: createCustomPoolUseCase,
		deletePoolUseCase:       deletePoolUseCase,
	}
}

// GetPools обрабатывает GET /api/map-pools/games/:gameId
// Требует авторизации - возвращает только пулы текущего пользователя
func (h *MapPoolHandler) GetPools(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	gameIDStr := c.Param("gameId")
	gameID, err := strconv.ParseUint(gameIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	result, err := h.getPoolsUseCase.Execute(map_pool.GetPoolsInput{
		GameID: uint(gameID),
		UserID: user.ID,
	})
	if err != nil {
		switch err {
		case map_pool.ErrGameNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToMapPoolResponseList(result.Pools))
}

// GetPool обрабатывает GET /api/map-pools/:id
// Требует авторизации - возвращает пул только если он принадлежит текущему пользователю
func (h *MapPoolHandler) GetPool(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool id"})
		return
	}

	result, err := h.getPoolUseCase.Execute(map_pool.GetPoolInput{
		PoolID: uint(id),
		UserID: user.ID,
	})
	if err != nil {
		switch err {
		case map_pool.ErrMapPoolNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "map pool not found"})
		case map_pool.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToMapPoolResponse(result.Pool))
}

// CreateCustomPool обрабатывает POST /api/map-pools
func (h *MapPoolHandler) CreateCustomPool(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateCustomMapPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем game_id из query параметра или используем Valorant (id=1) по умолчанию
	gameIDStr := c.Query("game_id")
	gameID := uint(1) // По умолчанию Valorant
	if gameIDStr != "" {
		parsedID, err := strconv.ParseUint(gameIDStr, 10, 32)
		if err == nil {
			gameID = uint(parsedID)
		}
	}

	result, err := h.createCustomPoolUseCase.Execute(map_pool.CreateCustomPoolInput{
		UserID:  user.ID,
		GameID:  gameID,
		Name:    req.Name,
		MapIDs:  req.MapIDs,
	})

	if err != nil {
		switch err {
		case map_pool.ErrGameNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		case map_pool.ErrMapNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "map not found"})
		case map_pool.ErrInvalidMapPool:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid map pool"})
		case map_pool.ErrPoolHasNoMaps:
			c.JSON(http.StatusBadRequest, gin.H{"error": "map pool must have at least one map"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ToMapPoolResponse(result.Pool))
}

// DeletePool обрабатывает DELETE /api/map-pools/:id
func (h *MapPoolHandler) DeletePool(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool id"})
		return
	}

	err = h.deletePoolUseCase.Execute(map_pool.DeletePoolInput{
		PoolID: uint(id),
		UserID: user.ID,
	})

	if err != nil {
		switch err {
		case map_pool.ErrMapPoolNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "map pool not found"})
		case map_pool.ErrCannotDeleteSystem:
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete system map pool"})
		case map_pool.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
