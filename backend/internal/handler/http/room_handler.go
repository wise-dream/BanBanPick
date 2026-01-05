package http

import (
	"net/http"
	"strconv"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/handler/dto"
	"github.com/bbp/backend/internal/middleware"
	"github.com/bbp/backend/internal/usecase/room"
	ws "github.com/bbp/backend/pkg/websocket"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	createRoomUseCase       *room.CreateRoomUseCase
	getRoomUseCase          *room.GetRoomUseCase
	getRoomBySessionUseCase *room.GetRoomBySessionUseCase
	getRoomsListUseCase     *room.GetRoomsListUseCase
	joinRoomUseCase         *room.JoinRoomUseCase
	leaveRoomUseCase        *room.LeaveRoomUseCase
	deleteRoomUseCase       *room.DeleteRoomUseCase
	updateRoomUseCase       *room.UpdateRoomUseCase
	wsManager               *ws.Manager
}

func NewRoomHandler(
	createRoomUseCase *room.CreateRoomUseCase,
	getRoomUseCase *room.GetRoomUseCase,
	getRoomBySessionUseCase *room.GetRoomBySessionUseCase,
	getRoomsListUseCase *room.GetRoomsListUseCase,
	joinRoomUseCase *room.JoinRoomUseCase,
	leaveRoomUseCase *room.LeaveRoomUseCase,
	deleteRoomUseCase *room.DeleteRoomUseCase,
	updateRoomUseCase *room.UpdateRoomUseCase,
	wsManager *ws.Manager,
) *RoomHandler {
	return &RoomHandler{
		createRoomUseCase:       createRoomUseCase,
		getRoomUseCase:          getRoomUseCase,
		getRoomBySessionUseCase: getRoomBySessionUseCase,
		getRoomsListUseCase:     getRoomsListUseCase,
		joinRoomUseCase:         joinRoomUseCase,
		leaveRoomUseCase:        leaveRoomUseCase,
		deleteRoomUseCase:       deleteRoomUseCase,
		updateRoomUseCase:       updateRoomUseCase,
		wsManager:               wsManager,
	}
}

// GetRooms обрабатывает GET /api/rooms
func (h *RoomHandler) GetRooms(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	typeStr := c.Query("type") // Опциональный параметр: "public" или "private"

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	input := room.GetRoomsListInput{
		Limit:  limit,
		Offset: offset,
	}
	
	// Если указан тип, добавляем в фильтр
	if typeStr != "" && (typeStr == "public" || typeStr == "private") {
		input.Type = &typeStr
	}

	result, err := h.getRoomsListUseCase.Execute(input)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  dto.ToRoomResponseList(result.Rooms),
		"total": result.Total,
	})
}

// CreateRoom обрабатывает POST /api/rooms
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxParticipants := 10 // По умолчанию
	if req.MaxParticipants != nil {
		maxParticipants = *req.MaxParticipants
	}

	var vetoType *entities.VetoType
	if req.VetoType != nil {
		vt := entities.VetoType(*req.VetoType)
		vetoType = &vt
	}
	
	result, err := h.createRoomUseCase.Execute(room.CreateRoomInput{
		OwnerID:         user.ID,
		Name:            req.Name,
		Type:            entities.RoomType(req.Type),
		GameID:          req.GameID,
		MapPoolID:       req.MapPoolID,
		VetoType:        vetoType,
		MaxParticipants: maxParticipants,
		Password:        req.Password,
	})

	if err != nil {
		switch err {
		case room.ErrGameNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		case room.ErrMapPoolNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "map pool not found"})
		case room.ErrInvalidRoom:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ToRoomResponse(result.Room))
}

// GetRoom обрабатывает GET /api/rooms/:id
func (h *RoomHandler) GetRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	result, err := h.getRoomUseCase.Execute(uint(id))
	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToRoomResponse(result.Room))
}

// GetRoomBySession обрабатывает GET /api/rooms/by-session/:sessionId
func (h *RoomHandler) GetRoomBySession(c *gin.Context) {
	sessionIdStr := c.Param("sessionId")
	sessionId, err := strconv.ParseUint(sessionIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	result, err := h.getRoomBySessionUseCase.Execute(uint(sessionId))
	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToRoomResponse(result.Room))
}

// JoinRoom обрабатывает POST /api/rooms/:id/join
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req dto.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Пароль опционален, игнорируем ошибку если тело пустое
		req.Password = ""
	}

	roomID := uint(id)
	result, err := h.joinRoomUseCase.Execute(room.JoinRoomInput{
		RoomID:   &roomID,
		UserID:   user.ID,
		Password: req.Password,
	})

	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		case room.ErrAlreadyInRoom:
			c.JSON(http.StatusConflict, gin.H{"error": "user is already in a room"})
		case room.ErrRoomFull:
			c.JSON(http.StatusBadRequest, gin.H{"error": "room is full"})
		case room.ErrInvalidCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room password"})
		case room.ErrInvalidRoom:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	if result.Room == nil {
		c.JSON(http.StatusOK, gin.H{"message": "room deleted"})
		return
	}

	// Отправляем WebSocket сообщение всем участникам комнаты об обновлении списка участников
	if h.wsManager != nil {
		h.wsManager.BroadcastToRoom(roomID, ws.Message{
			Type: "room:participants:updated",
			Data: map[string]interface{}{
				"room_id":      roomID,
				"user_id":      user.ID,
				"participants": result.Room.Participants,
			},
		})
	}

	c.JSON(http.StatusOK, dto.ToRoomResponse(result.Room))
}

// LeaveRoom обрабатывает POST /api/rooms/:id/leave
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	result, err := h.leaveRoomUseCase.Execute(room.LeaveRoomInput{
		RoomID: uint(id),
		UserID: user.ID,
	})

	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		case room.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	if result.Room == nil {
		// Комната удалена - отправляем сообщение всем участникам
		if h.wsManager != nil {
			h.wsManager.BroadcastToRoom(uint(id), ws.Message{
				Type: "room:deleted",
				Data: map[string]interface{}{
					"room_id": uint(id),
				},
			})
		}
		c.JSON(http.StatusOK, gin.H{"message": "room deleted"})
		return
	}

	// Отправляем WebSocket сообщение всем участникам комнаты об обновлении списка участников
	if h.wsManager != nil {
		h.wsManager.BroadcastToRoom(uint(id), ws.Message{
			Type: "room:participants:updated",
			Data: map[string]interface{}{
				"room_id":      uint(id),
				"user_id":      user.ID,
				"participants": result.Room.Participants,
			},
		})
	}

	c.JSON(http.StatusOK, dto.ToRoomResponse(result.Room))
}

// DeleteRoom обрабатывает DELETE /api/rooms/:id
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	err = h.deleteRoomUseCase.Execute(room.DeleteRoomInput{
		RoomID: uint(id),
		UserID: user.ID,
	})

	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		case room.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// UpdateRoom обрабатывает PUT /api/rooms/:id
func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	// Получаем пользователя из контекста (требует авторизации)
	user, err := middleware.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем статус из строки в RoomStatus
	var status *entities.RoomStatus
	if req.Status != nil {
		s := entities.RoomStatus(*req.Status)
		status = &s
	}
	
	// Преобразуем veto_type из строки в VetoType
	var vetoType *entities.VetoType
	if req.VetoType != nil {
		vt := entities.VetoType(*req.VetoType)
		vetoType = &vt
	}

	result, err := h.updateRoomUseCase.Execute(room.UpdateRoomInput{
		RoomID:        uint(id),
		UserID:        user.ID,
		MapPoolID:     req.MapPoolID,
		VetoType:      vetoType,
		VetoSessionID: req.VetoSessionID,
		Status:        status,
	})

	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		case room.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	// Отправляем WebSocket сообщение об обновлении комнаты
	if h.wsManager != nil {
		var vetoTypeStr *string
		if result.Room.VetoType != nil {
			s := string(*result.Room.VetoType)
			vetoTypeStr = &s
		}
		
		h.wsManager.BroadcastToRoom(uint(id), ws.Message{
			Type: "room:state",
			Data: map[string]interface{}{
				"room_id":        uint(id),
				"veto_session_id": result.Room.VetoSessionID,
				"map_pool_id":     result.Room.MapPoolID,
				"veto_type":       vetoTypeStr,
				"status":          result.Room.Status,
			},
		})
	}

	c.JSON(http.StatusOK, dto.ToRoomResponse(result.Room))
}

// GetParticipants обрабатывает GET /api/rooms/:id/participants
func (h *RoomHandler) GetParticipants(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	result, err := h.getRoomUseCase.Execute(uint(id))
	if err != nil {
		switch err {
		case room.ErrRoomNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	participants := make([]dto.RoomParticipantResponse, len(result.Room.Participants))
	for i, p := range result.Room.Participants {
		participants[i] = dto.ToRoomParticipantResponse(&p)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": participants,
	})
}
