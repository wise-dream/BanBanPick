package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Manager manages WebSocket connections
type Manager struct {
	// Registered clients grouped by room
	Rooms map[uint]map[*Client]bool

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast message to all clients in a room
	Broadcast chan *RoomMessage

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Message handler
	MessageHandler func(*Client, *Message)
}

// RoomMessage represents a message to be broadcast to a room
type RoomMessage struct {
	RoomID  uint
	Message Message
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return &Manager{
		Rooms:      make(map[uint]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *RoomMessage),
	}
}

// SetMessageHandler sets the message handler
func (m *Manager) SetMessageHandler(handler func(*Client, *Message)) {
	m.MessageHandler = handler
}

// Run starts the manager
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.mu.Lock()
			if m.Rooms[client.RoomID] == nil {
				m.Rooms[client.RoomID] = make(map[*Client]bool)
			}
			m.Rooms[client.RoomID][client] = true
			m.mu.Unlock()

			log.Printf("Client %d joined room %d", client.UserID, client.RoomID)

			// Notify other clients in the room
			m.broadcastToRoom(client.RoomID, Message{
				Type: "room:join",
				Data: map[string]interface{}{
					"user_id": client.UserID,
					"room_id": client.RoomID,
				},
			}, client)

		case client := <-m.Unregister:
			m.mu.Lock()
			if room, ok := m.Rooms[client.RoomID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
					if len(room) == 0 {
						delete(m.Rooms, client.RoomID)
					}
				}
			}
			m.mu.Unlock()

			log.Printf("Client %d left room %d", client.UserID, client.RoomID)

			// Notify other clients in the room
			m.broadcastToRoom(client.RoomID, Message{
				Type: "room:leave",
				Data: map[string]interface{}{
					"user_id": client.UserID,
					"room_id": client.RoomID,
				},
			}, nil)

		case roomMsg := <-m.Broadcast:
			m.broadcastToRoom(roomMsg.RoomID, roomMsg.Message, nil)
		}
	}
}

// HandleMessage handles incoming messages from clients
func (m *Manager) HandleMessage(client *Client, msg *Message) {
	if m.MessageHandler != nil {
		m.MessageHandler(client, msg)
	}
}

// BroadcastToRoom broadcasts a message to all clients in a room
func (m *Manager) BroadcastToRoom(roomID uint, msg Message) {
	m.Broadcast <- &RoomMessage{
		RoomID:  roomID,
		Message: msg,
	}
}

// broadcastToRoom broadcasts a message to all clients in a room (internal)
func (m *Manager) broadcastToRoom(roomID uint, msg Message, exclude *Client) {
	m.mu.RLock()
	room, ok := m.Rooms[roomID]
	if !ok {
		m.mu.RUnlock()
		return
	}

	// Create a copy of the room to avoid locking issues
	clients := make([]*Client, 0, len(room))
	for client := range room {
		if exclude == nil || client != exclude {
			clients = append(clients, client)
		}
	}
	m.mu.RUnlock()

	// Send message to all clients
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			m.mu.Lock()
			delete(m.Rooms[roomID], client)
			m.mu.Unlock()
		}
	}
}

// GetRoomClients returns all clients in a room
func (m *Manager) GetRoomClients(roomID uint) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.Rooms[roomID]
	if !ok {
		return nil
	}

	clients := make([]*Client, 0, len(room))
	for client := range room {
		clients = append(clients, client)
	}

	return clients
}

// GetRoomClientCount returns the number of clients in a room
func (m *Manager) GetRoomClientCount(roomID uint) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.Rooms[roomID]
	if !ok {
		return 0
	}

	return len(room)
}
