package websocket

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bbp/backend/internal/domain/entities"
	"github.com/bbp/backend/internal/domain/repositories"
	"github.com/bbp/backend/internal/handler/dto"
	"github.com/bbp/backend/internal/usecase/veto"
	ws "github.com/bbp/backend/pkg/websocket"
	jwtPkg "github.com/bbp/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: В продакшене проверять origin
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type RoomWebSocketHandler struct {
	manager           *ws.Manager
	roomRepo          repositories.RoomRepository
	vetoSessionRepo   repositories.VetoSessionRepository
	vetoActionRepo    repositories.VetoActionRepository
	mapRepo           repositories.MapRepository
	mapPoolRepo       repositories.MapPoolRepository
	jwtService        *jwtPkg.JWTService
	banMapUseCase     *veto.BanMapUseCase
	pickMapUseCase    *veto.PickMapUseCase
	resetSessionUseCase *veto.ResetSessionUseCase
	startSessionUseCase *veto.StartSessionUseCase
}

func NewRoomWebSocketHandler(
	manager *ws.Manager,
	roomRepo repositories.RoomRepository,
	vetoSessionRepo repositories.VetoSessionRepository,
	vetoActionRepo repositories.VetoActionRepository,
	mapRepo repositories.MapRepository,
	mapPoolRepo repositories.MapPoolRepository,
	jwtService *jwtPkg.JWTService,
	banMapUseCase *veto.BanMapUseCase,
	pickMapUseCase *veto.PickMapUseCase,
	resetSessionUseCase *veto.ResetSessionUseCase,
	startSessionUseCase *veto.StartSessionUseCase,
) *RoomWebSocketHandler {
	handler := &RoomWebSocketHandler{
		manager:           manager,
		roomRepo:          roomRepo,
		vetoSessionRepo:   vetoSessionRepo,
		vetoActionRepo:    vetoActionRepo,
		mapRepo:           mapRepo,
		mapPoolRepo:       mapPoolRepo,
		jwtService:        jwtService,
		banMapUseCase:     banMapUseCase,
		pickMapUseCase:    pickMapUseCase,
		resetSessionUseCase: resetSessionUseCase,
		startSessionUseCase: startSessionUseCase,
	}

	// Set message handler
	manager.SetMessageHandler(handler.HandleMessage)

	return handler
}

// HandleWebSocket handles WebSocket connections for rooms
func (h *RoomWebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter (WebSocket doesn't support headers during upgrade)
	token := c.Query("token")
	if token == "" {
		// Try to get from Authorization header as fallback
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	// Validate token
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	// Create user from claims
	user := &entities.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}

	// Get room ID from URL
	roomIDStr := c.Param("roomId")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	// Verify user is a participant in the room
	room, err := h.roomRepo.GetByID(uint(roomID))
	if err != nil || room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// Check if user is a participant
	isParticipant := false
	for _, p := range room.Participants {
		if p.UserID == user.ID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a participant"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// Create client
	clientID := uint(len(h.manager.GetRoomClients(uint(roomID))) + 1)
	client := ws.NewClient(clientID, user.ID, uint(roomID), conn, h.manager)

	// Register client
	h.manager.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	// Send initial room state
	h.sendRoomState(client, room)
}

// sendRoomState sends the current room state to the client
func (h *RoomWebSocketHandler) sendRoomState(client *ws.Client, room *entities.Room) {
	// Load veto session if exists
	var vetoSession *entities.VetoSession
	if room.VetoSessionID != nil {
		session, err := h.vetoSessionRepo.GetByID(*room.VetoSessionID)
		if err == nil && session != nil {
			vetoSession = session
		}
	}

	state := map[string]interface{}{
		"room_id":      room.ID,
		"veto_session": vetoSession,
	}

	client.SendMessage(ws.Message{
		Type: "room:state",
		Data: state,
	})
}

// broadcastRoomState broadcasts room state to all clients in the room
func (h *RoomWebSocketHandler) broadcastRoomState(roomID uint) {
	room, err := h.roomRepo.GetByID(roomID)
	if err != nil || room == nil {
		return
	}

	// Load veto session if exists
	var vetoSession *entities.VetoSession
	if room.VetoSessionID != nil {
		session, err := h.vetoSessionRepo.GetByID(*room.VetoSessionID)
		if err == nil && session != nil {
			vetoSession = session
		}
	}

	state := map[string]interface{}{
		"room_id":      room.ID,
		"veto_session": vetoSession,
	}

	h.manager.BroadcastToRoom(roomID, ws.Message{
		Type: "room:state",
		Data: state,
	})
}

// HandleMessage handles incoming messages from clients
func (h *RoomWebSocketHandler) HandleMessage(client *ws.Client, msg *ws.Message) {
	log.Printf("Received WebSocket message from user %d in room %d: type=%s", client.UserID, client.RoomID, msg.Type)
	
	// Валидация сообщения
	if msg.Type == "" {
		log.Printf("Error: empty message type from user %d", client.UserID)
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "message type is required"},
		})
		return
	}
	
	switch msg.Type {
	case "veto:ban":
		h.handleVetoBan(client, msg)
	case "veto:pick":
		h.handleVetoPick(client, msg)
	case "veto:swap":
		h.handleVetoSwap(client, msg)
	case "veto:start":
		h.handleVetoStart(client, msg)
	case "veto:reset":
		h.handleVetoReset(client, msg)
	case "ping":
		client.SendMessage(ws.Message{
			Type: "pong",
			Data: map[string]string{"message": "pong"},
		})
	default:
		log.Printf("Unknown message type from user %d: %s", client.UserID, msg.Type)
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "unknown message type: " + msg.Type},
		})
	}
}

// handleVetoBan handles veto ban action
func (h *RoomWebSocketHandler) handleVetoBan(client *ws.Client, msg *ws.Message) {
	// Parse message data
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "invalid message data"},
		})
		return
	}

	sessionID, ok := data["session_id"].(float64)
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "session_id required"},
		})
		return
	}

	mapID, ok := data["map_id"].(float64)
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "map_id required"},
		})
		return
	}

	team, ok := data["team"].(string)
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "team required"},
		})
		return
	}

	// Проверяем, что пользователь участвует в комнате
	room, err := h.roomRepo.GetByID(client.RoomID)
	if err != nil || room == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "room not found"},
		})
		return
	}

	// Проверяем, что в комнате есть veto сессия
	if room.VetoSessionID == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "no veto session in room"},
		})
		return
	}

	// Проверяем, что session_id совпадает с сессией комнаты
	if uint(sessionID) != *room.VetoSessionID {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "session_id mismatch"},
		})
		return
	}

	// Вызываем use case для бана карты
	output, err := h.banMapUseCase.Execute(veto.BanMapInput{
		SessionID: uint(sessionID),
		MapID:     uint(mapID),
		Team:      team,
	})

	if err != nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
		return
	}

	// Broadcast обновленное состояние сессии всем участникам комнаты
	// Конвертируем entity в DTO для правильной структуры с map_pool и actions
	sessionDTO := dto.ToVetoSessionResponse(output.Session)
	
	// Загружаем map_pool для включения в ответ
	if output.Session != nil {
		mapPool, err := h.mapPoolRepo.GetByID(output.Session.MapPoolID)
		if err == nil && mapPool != nil {
			mapPoolResp := dto.ToMapPoolResponse(mapPool)
			sessionDTO.MapPool = &mapPoolResp
		}
	}
	
	h.manager.BroadcastToRoom(client.RoomID, ws.Message{
		Type: "veto:ban",
		Data: map[string]interface{}{
			"session": sessionDTO,
			"action":  output.Action,
			"user_id": client.UserID,
		},
	})

	// Если сессия завершена, отправляем обновленное состояние комнаты
	if output.Session.Status == entities.VetoStatusFinished {
		updatedRoom, _ := h.roomRepo.GetByID(client.RoomID)
		if updatedRoom != nil {
			h.broadcastRoomState(client.RoomID)
		}
	}
}

// handleVetoPick handles veto pick action
func (h *RoomWebSocketHandler) handleVetoPick(client *ws.Client, msg *ws.Message) {
	// Similar to handleVetoBan
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "invalid message data"},
		})
		return
	}

	sessionID, ok := data["session_id"].(float64)
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "session_id required"},
		})
		return
	}

	mapID, ok := data["map_id"].(float64)
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "map_id required"},
		})
		return
	}

	team, ok := data["team"].(string)
	if !ok {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "team required"},
		})
		return
	}

	// Проверяем, что пользователь участвует в комнате
	room, err := h.roomRepo.GetByID(client.RoomID)
	if err != nil || room == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "room not found"},
		})
		return
	}

	// Проверяем, что в комнате есть veto сессия
	if room.VetoSessionID == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "no veto session in room"},
		})
		return
	}

	// Проверяем, что session_id совпадает с сессией комнаты
	if uint(sessionID) != *room.VetoSessionID {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "session_id mismatch"},
		})
		return
	}

	// Вызываем use case для выбора карты
	output, err := h.pickMapUseCase.Execute(veto.PickMapInput{
		SessionID: uint(sessionID),
		MapID:     uint(mapID),
		Team:      team,
	})

	if err != nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
		return
	}

	// Broadcast обновленное состояние сессии всем участникам комнаты
	// Конвертируем entity в DTO для правильной структуры с map_pool и actions
	sessionDTO := dto.ToVetoSessionResponse(output.Session)
	
	// Загружаем map_pool для включения в ответ
	if output.Session != nil {
		mapPool, err := h.mapPoolRepo.GetByID(output.Session.MapPoolID)
		if err == nil && mapPool != nil {
			mapPoolResp := dto.ToMapPoolResponse(mapPool)
			sessionDTO.MapPool = &mapPoolResp
		}
	}
	
	h.manager.BroadcastToRoom(client.RoomID, ws.Message{
		Type: "veto:pick",
		Data: map[string]interface{}{
			"session": sessionDTO,
			"action":  output.Action,
			"user_id": client.UserID,
		},
	})

	// Если сессия завершена, отправляем обновленное состояние комнаты
	if output.Session.Status == entities.VetoStatusFinished {
		updatedRoom, _ := h.roomRepo.GetByID(client.RoomID)
		if updatedRoom != nil {
			h.broadcastRoomState(client.RoomID)
		}
	}
}

// handleVetoSwap handles veto swap action
func (h *RoomWebSocketHandler) handleVetoSwap(client *ws.Client, msg *ws.Message) {
	// Broadcast swap action
	h.manager.BroadcastToRoom(client.RoomID, ws.Message{
		Type: "veto:swap",
		Data: map[string]interface{}{
			"user_id": client.UserID,
		},
	})
}

// handleVetoStart handles veto start action
func (h *RoomWebSocketHandler) handleVetoStart(client *ws.Client, msg *ws.Message) {
	// Проверяем, что пользователь участвует в комнате
	room, err := h.roomRepo.GetByID(client.RoomID)
	if err != nil || room == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "room not found"},
		})
		return
	}

	// Проверяем, что в комнате есть veto сессия
	if room.VetoSessionID == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "no veto session in room"},
		})
		return
	}

	// Извлекаем данные из исходного сообщения
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Printf("Error: invalid message data in veto:start from user %d", client.UserID)
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "invalid message data"},
		})
		return
	}

	// Используем session_id из комнаты или из сообщения
	sessionID := *room.VetoSessionID
	if msgSessionID, ok := data["session_id"].(float64); ok {
		sessionID = uint(msgSessionID)
	}

	// Вызываем use case для старта сессии
	output, err := h.startSessionUseCase.Execute(veto.StartSessionInput{
		SessionID: sessionID,
	})

	if err != nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
		return
	}

	// Broadcast обновленное состояние сессии всем участникам комнаты
	// Конвертируем entity в DTO для правильной структуры с map_pool и actions
	sessionDTO := dto.ToVetoSessionResponse(output.Session)
	
	// Загружаем map_pool для включения в ответ
	if output.Session != nil {
		mapPool, err := h.mapPoolRepo.GetByID(output.Session.MapPoolID)
		if err == nil && mapPool != nil {
			mapPoolResp := dto.ToMapPoolResponse(mapPool)
			sessionDTO.MapPool = &mapPoolResp
		}
	}
	
	h.manager.BroadcastToRoom(client.RoomID, ws.Message{
		Type: "veto:start",
		Data: map[string]interface{}{
			"session": sessionDTO,
			"user_id": client.UserID,
		},
	})
	
	log.Printf("Broadcasted veto:start to room %d for session %d", client.RoomID, sessionID)
}

// handleVetoReset handles veto reset action
func (h *RoomWebSocketHandler) handleVetoReset(client *ws.Client, msg *ws.Message) {
	// Проверяем, что пользователь участвует в комнате
	room, err := h.roomRepo.GetByID(client.RoomID)
	if err != nil || room == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "room not found"},
		})
		return
	}

	// Проверяем, что в комнате есть veto сессия
	if room.VetoSessionID == nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": "no veto session in room"},
		})
		return
	}

	// Вызываем use case для сброса сессии
	output, err := h.resetSessionUseCase.Execute(veto.ResetSessionInput{
		SessionID: *room.VetoSessionID,
	})

	if err != nil {
		client.SendMessage(ws.Message{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
		return
	}

	// Broadcast обновленное состояние сессии всем участникам комнаты
	// Конвертируем entity в DTO для правильной структуры с map_pool и actions
	sessionDTO := dto.ToVetoSessionResponse(output.Session)
	
	// Загружаем map_pool для включения в ответ
	if output.Session != nil {
		mapPool, err := h.mapPoolRepo.GetByID(output.Session.MapPoolID)
		if err == nil && mapPool != nil {
			mapPoolResp := dto.ToMapPoolResponse(mapPool)
			sessionDTO.MapPool = &mapPoolResp
		}
	}
	
	h.manager.BroadcastToRoom(client.RoomID, ws.Message{
		Type: "veto:reset",
		Data: map[string]interface{}{
			"session": sessionDTO,
			"user_id": client.UserID,
		},
	})
	
	log.Printf("Broadcasted veto:reset to room %d for session %d", client.RoomID, *room.VetoSessionID)
}
