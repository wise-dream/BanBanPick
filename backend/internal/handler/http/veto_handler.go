package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/handler/dto"
	"github.com/bbp/backend/internal/middleware"
	"github.com/bbp/backend/internal/usecase/veto"
	"github.com/bbp/backend/internal/domain/entities"
	ws "github.com/bbp/backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

type VetoHandler struct {
	createSessionUseCase  *veto.CreateSessionUseCase
	getSessionUseCase     *veto.GetSessionUseCase
	getNextActionUseCase  *veto.GetNextActionUseCase
	banMapUseCase         *veto.BanMapUseCase
	pickMapUseCase        *veto.PickMapUseCase
	selectSideUseCase     *veto.SelectSideUseCase
	resetSessionUseCase   *veto.ResetSessionUseCase
	startSessionUseCase   *veto.StartSessionUseCase
	mapPoolRepo           repositories.MapPoolRepository
	roomRepo              repositories.RoomRepository
	wsManager             *ws.Manager
}

func NewVetoHandler(
	createSessionUseCase *veto.CreateSessionUseCase,
	getSessionUseCase *veto.GetSessionUseCase,
	getNextActionUseCase *veto.GetNextActionUseCase,
	banMapUseCase *veto.BanMapUseCase,
	pickMapUseCase *veto.PickMapUseCase,
	selectSideUseCase *veto.SelectSideUseCase,
	resetSessionUseCase *veto.ResetSessionUseCase,
	startSessionUseCase *veto.StartSessionUseCase,
	mapPoolRepo repositories.MapPoolRepository,
	roomRepo repositories.RoomRepository,
	wsManager *ws.Manager,
) *VetoHandler {
	return &VetoHandler{
		createSessionUseCase: createSessionUseCase,
		getSessionUseCase:    getSessionUseCase,
		getNextActionUseCase: getNextActionUseCase,
		banMapUseCase:        banMapUseCase,
		pickMapUseCase:       pickMapUseCase,
		selectSideUseCase:    selectSideUseCase,
		resetSessionUseCase:  resetSessionUseCase,
		startSessionUseCase:  startSessionUseCase,
		mapPoolRepo:          mapPoolRepo,
		roomRepo:             roomRepo,
		wsManager:            wsManager,
	}
}

// CreateSession обрабатывает POST /api/veto/sessions
func (h *VetoHandler) CreateSession(c *gin.Context) {
	var req dto.CreateVetoSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем user ID из контекста (опционально, если пользователь авторизован)
	var userID *uint
	user, err := middleware.GetUserFromContext(c)
	if err == nil && user != nil {
		userID = &user.ID
	}

	result, err := h.createSessionUseCase.Execute(veto.CreateSessionInput{
		UserID:       userID,
		GameID:       req.GameID,
		MapPoolID:    req.MapPoolID,
		Type:         entities.VetoType(req.Type),
		TeamAName:    req.TeamAName,
		TeamBName:    req.TeamBName,
		TimerSeconds: req.TimerSeconds,
	})

	if err != nil {
		switch err {
		case veto.ErrGameNotFound, veto.ErrMapPoolNotFound, veto.ErrInvalidMapPool:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ToVetoSessionResponse(result.Session))
}

// GetSession обрабатывает GET /api/veto/sessions/:id
func (h *VetoHandler) GetSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	result, err := h.getSessionUseCase.Execute(uint(id))
	if err != nil {
		switch err {
		case veto.ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// Загружаем map_pool для включения в ответ
	var mapPool *entities.MapPool
	if result.Session != nil {
		pool, err := h.mapPoolRepo.GetByID(result.Session.MapPoolID)
		if err == nil && pool != nil {
			mapPool = pool
		}
	}

	response := dto.ToVetoSessionResponse(result.Session)
	if mapPool != nil {
		mapPoolResp := dto.ToMapPoolResponse(mapPool)
		response.MapPool = &mapPoolResp
	}

	c.JSON(http.StatusOK, response)
}

// GetSessionByShareToken обрабатывает GET /api/veto/sessions/share/:token
func (h *VetoHandler) GetSessionByShareToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	result, err := h.getSessionUseCase.ExecuteByShareToken(token)
	if err != nil {
		switch err {
		case veto.ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// Загружаем map_pool для включения в ответ
	var mapPool *entities.MapPool
	if result.Session != nil {
		pool, err := h.mapPoolRepo.GetByID(result.Session.MapPoolID)
		if err != nil {
			log.Printf("Error loading map pool %d for session by token: %v", result.Session.MapPoolID, err)
		} else if pool != nil {
			mapPool = pool
			log.Printf("Loaded map pool %d for session by token: %d maps", pool.ID, len(pool.Maps))
			if len(pool.Maps) == 0 {
				log.Printf("Warning: map pool %d has no maps", pool.ID)
			}
		} else {
			log.Printf("Map pool %d not found for session by token", result.Session.MapPoolID)
		}
	}

	response := dto.ToVetoSessionResponse(result.Session)
	if mapPool != nil && len(mapPool.Maps) > 0 {
		mapPoolResp := dto.ToMapPoolResponse(mapPool)
		response.MapPool = &mapPoolResp
		log.Printf("Including map pool in response for session by token with %d maps", len(mapPool.Maps))
	} else {
		log.Printf("Not including map pool in response for session by token (pool is nil or has no maps)")
	}

	c.JSON(http.StatusOK, response)
}

// GetNextAction обрабатывает GET /api/veto/sessions/:id/next-action
func (h *VetoHandler) GetNextAction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	result, err := h.getNextActionUseCase.Execute(uint(id))
	if err != nil {
		switch err {
		case veto.ErrSessionNotFound, veto.ErrMapPoolNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.NextActionResponse{
		ActionType:         string(result.ActionType),
		CurrentStep:        result.CurrentStep,
		CurrentTeam:        result.CurrentTeam,
		CanBan:             result.CanBan,
		CanPick:            result.CanPick,
		NeedsSideSelection: result.NeedsSideSelection,
		SideSelectionTeam:  result.SideSelectionTeam,
		Message:            result.Message,
	})
}

// BanMap обрабатывает POST /api/veto/sessions/:id/ban
func (h *VetoHandler) BanMap(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var req dto.BanMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.banMapUseCase.Execute(veto.BanMapInput{
		SessionID: uint(id),
		MapID:     req.MapID,
		Team:      req.Team,
	})

	if err != nil {
		switch err {
		case veto.ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case veto.ErrSessionFinished:
			c.JSON(http.StatusBadRequest, gin.H{"error": "session is already finished"})
		case veto.ErrMapNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "map not found"})
		case veto.ErrMapAlreadyBanned:
			c.JSON(http.StatusBadRequest, gin.H{"error": "map is already banned"})
		case veto.ErrNotYourTurn:
			c.JSON(http.StatusBadRequest, gin.H{"error": "not your turn"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToVetoSessionResponse(result.Session))
}

// PickMap обрабатывает POST /api/veto/sessions/:id/pick
func (h *VetoHandler) PickMap(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var req dto.PickMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.pickMapUseCase.Execute(veto.PickMapInput{
		SessionID: uint(id),
		MapID:     req.MapID,
		Team:      req.Team,
	})

	if err != nil {
		switch err {
		case veto.ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case veto.ErrSessionFinished:
			c.JSON(http.StatusBadRequest, gin.H{"error": "session is already finished"})
		case veto.ErrInvalidAction:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		case veto.ErrMapNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "map not found"})
		case veto.ErrMapAlreadyPicked:
			c.JSON(http.StatusBadRequest, gin.H{"error": "map is already picked"})
		case veto.ErrNotYourTurn:
			c.JSON(http.StatusBadRequest, gin.H{"error": "not your turn"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToVetoSessionResponse(result.Session))
}

// SelectSide обрабатывает POST /api/veto/sessions/:id/select-side
func (h *VetoHandler) SelectSide(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var req dto.SelectSideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.selectSideUseCase.Execute(veto.SelectSideInput{
		SessionID: uint(id),
		Side:      req.Side,
		Team:      req.Team,
	})

	if err != nil {
		switch err {
		case veto.ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case veto.ErrInvalidAction:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		case veto.ErrNotYourTurn:
			c.JSON(http.StatusBadRequest, gin.H{"error": "not your turn to select side"})
		case veto.ErrSessionFinished:
			c.JSON(http.StatusBadRequest, gin.H{"error": "session is finished"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// Находим комнату по session_id для отправки WebSocket сообщения
	room, err := h.roomRepo.GetByVetoSessionID(uint(id))
	if err == nil && room != nil {
		// Конвертируем entity в DTO для правильной структуры с map_pool и actions
		sessionDTO := dto.ToVetoSessionResponse(result.Session)
		
		// Загружаем map_pool для включения в ответ
		if result.Session != nil {
			mapPool, err := h.mapPoolRepo.GetByID(result.Session.MapPoolID)
			if err == nil && mapPool != nil {
				mapPoolResp := dto.ToMapPoolResponse(mapPool)
				sessionDTO.MapPool = &mapPoolResp
			}
		}
		
		// Broadcast обновленное состояние сессии всем участникам комнаты
		h.wsManager.BroadcastToRoom(room.ID, ws.Message{
			Type: "veto:side",
			Data: map[string]interface{}{
				"session": sessionDTO,
				"action":  result.Action,
			},
		})
		
		log.Printf("Broadcasted veto:side to room %d for session %d", room.ID, uint(id))
	}

	// Загружаем map_pool для включения в ответ (как и в других handler'ах)
	sessionDTO := dto.ToVetoSessionResponse(result.Session)
	if result.Session != nil {
		mapPool, err := h.mapPoolRepo.GetByID(result.Session.MapPoolID)
		if err == nil && mapPool != nil {
			mapPoolResp := dto.ToMapPoolResponse(mapPool)
			sessionDTO.MapPool = &mapPoolResp
		}
	}

	c.JSON(http.StatusOK, sessionDTO)
}

// StartSession обрабатывает POST /api/veto/sessions/:id/start
func (h *VetoHandler) StartSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	result, err := h.startSessionUseCase.Execute(veto.StartSessionInput{
		SessionID: uint(id),
	})

	if err != nil {
		switch err {
		case veto.ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case veto.ErrSessionAlreadyStarted:
			c.JSON(http.StatusBadRequest, gin.H{"error": "session is already started"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// Находим комнату по session_id для отправки WebSocket сообщения
	room, err := h.roomRepo.GetByVetoSessionID(uint(id))
	if err == nil && room != nil {
		// Конвертируем entity в DTO для правильной структуры с map_pool и actions
		sessionDTO := dto.ToVetoSessionResponse(result.Session)
		
		// Загружаем map_pool для включения в ответ
		if result.Session != nil {
			mapPool, err := h.mapPoolRepo.GetByID(result.Session.MapPoolID)
			if err == nil && mapPool != nil {
				mapPoolResp := dto.ToMapPoolResponse(mapPool)
				sessionDTO.MapPool = &mapPoolResp
			}
		}
		
		// Broadcast обновленное состояние сессии всем участникам комнаты
		h.wsManager.BroadcastToRoom(room.ID, ws.Message{
			Type: "veto:start",
			Data: map[string]interface{}{
				"session": sessionDTO,
			},
		})
		
		log.Printf("Broadcasted veto:start to room %d for session %d", room.ID, uint(id))
	}

	c.JSON(http.StatusOK, dto.ToVetoSessionResponse(result.Session))
}

// ResetSession обрабатывает POST /api/veto/sessions/:id/reset
func (h *VetoHandler) ResetSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	result, err := h.resetSessionUseCase.Execute(veto.ResetSessionInput{
		SessionID: uint(id),
	})

	if err != nil {
		switch err {
		case veto.ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// Находим комнату по session_id для отправки WebSocket сообщения
	room, err := h.roomRepo.GetByVetoSessionID(uint(id))
	if err == nil && room != nil {
		// Конвертируем entity в DTO для правильной структуры с map_pool и actions
		sessionDTO := dto.ToVetoSessionResponse(result.Session)
		
		// Загружаем map_pool для включения в ответ
		if result.Session != nil {
			mapPool, err := h.mapPoolRepo.GetByID(result.Session.MapPoolID)
			if err == nil && mapPool != nil {
				mapPoolResp := dto.ToMapPoolResponse(mapPool)
				sessionDTO.MapPool = &mapPoolResp
			}
		}
		
		// Broadcast обновленное состояние сессии всем участникам комнаты
		h.wsManager.BroadcastToRoom(room.ID, ws.Message{
			Type: "veto:reset",
			Data: map[string]interface{}{
				"session": sessionDTO,
			},
		})
		
		log.Printf("Broadcasted veto:reset to room %d for session %d", room.ID, uint(id))
	}

	c.JSON(http.StatusOK, dto.ToVetoSessionResponse(result.Session))
}