import { ref, getCurrentInstance, onUnmounted } from 'vue';
import { RoomWebSocket } from '../services/websocket/roomWebSocket';
import type { WebSocketMessage } from '../services/websocket/roomWebSocket';

export function useRoomWebSocket(roomId: number) {
  const ws = ref<RoomWebSocket | null>(null);
  const isConnected = ref(false);
  const messages = ref<WebSocketMessage[]>([]);
  const error = ref<string | null>(null);

  // Проверяем, что мы в контексте компонента перед регистрацией хука
  const instance = getCurrentInstance();
  if (instance) {
    onUnmounted(() => {
      disconnect();
    });
  }

  const connect = () => {
    console.log('🔌 [WebSocket] Attempting to connect to room:', roomId);
    
    if (ws.value) {
      ws.value.disconnect();
    }

    ws.value = new RoomWebSocket(roomId, {
      onOpen: () => {
        isConnected.value = true;
        error.value = null;
        console.log(`✅ [WebSocket] Connected to room ${roomId}`);
      },
      onClose: () => {
        isConnected.value = false;
        console.log(`🔌 [WebSocket] Disconnected from room ${roomId}`);
      },
      onError: (err) => {
        error.value = 'Ошибка WebSocket соединения';
        console.error('❌ [WebSocket] Error:', err);
      },
      onMessage: (message) => {
        console.log('📨 [WebSocket] Raw message received:', {
          roomId,
          timestamp: new Date().toISOString(),
          messageType: message?.type,
          hasData: !!message?.data,
          fullMessage: message
        });
        
        // Валидация сообщения
        if (!message || typeof message !== 'object' || !message.type) {
          console.error('❌ [WebSocket] Invalid message received:', message);
          return;
        }
        
        messages.value.push(message);
        console.log(`📬 [WebSocket] Message added to queue. Total messages: ${messages.value.length}`);
        
        // Keep only last 100 messages
        if (messages.value.length > 100) {
          messages.value.shift();
        }
      },
    });

    ws.value.connect();
  };

  const disconnect = () => {
    console.log('🔌 [WebSocket] Disconnect called:', {
      roomId,
      hasWebSocket: !!ws.value,
      isConnected: isConnected.value,
      timestamp: new Date().toISOString()
    });
    
    if (ws.value) {
      ws.value.disconnect();
      ws.value = null;
    }
    isConnected.value = false;
    
    console.log('✅ [WebSocket] Disconnected:', {
      roomId,
      isConnected: isConnected.value
    });
  };

  const send = (message: WebSocketMessage) => {
    if (!ws.value) {
      console.error('❌ [WebSocket] Cannot send message: WebSocket not initialized', {
        roomId,
        messageType: message?.type,
        message
      });
      return;
    }

    if (!isConnected.value) {
      console.warn('⚠️ [WebSocket] Sending message while disconnected:', {
        roomId,
        messageType: message?.type,
        isConnected: isConnected.value,
        message
      });
    }

    console.log('📤 [WebSocket] Sending message:', {
      roomId,
      timestamp: new Date().toISOString(),
      messageType: message?.type,
      messageData: message?.data,
      fullMessage: message,
      isConnected: isConnected.value
    });

    try {
      ws.value.send(message);
      console.log('✅ [WebSocket] Message sent successfully:', {
        roomId,
        messageType: message?.type
      });
    } catch (err) {
      console.error('❌ [WebSocket] Error sending message:', {
        roomId,
        messageType: message?.type,
        error: err,
        message
      });
      throw err;
    }
  };

  const sendVetoBan = (sessionId: number, mapId: number, team: 'A' | 'B') => {
    console.log('🚫 [WebSocket] sendVetoBan called:', {
      roomId,
      sessionId,
      mapId,
      team,
      timestamp: new Date().toISOString(),
      isConnected: isConnected.value
    });
    
    send({
      type: 'veto:ban',
      data: {
        session_id: sessionId,
        map_id: mapId,
        team: team,
      },
    });
  };

  const sendVetoPick = (sessionId: number, mapId: number, team: 'A' | 'B') => {
    console.log('🎯 [WebSocket] sendVetoPick called:', {
      roomId,
      sessionId,
      mapId,
      team,
      timestamp: new Date().toISOString(),
      isConnected: isConnected.value
    });
    
    send({
      type: 'veto:pick',
      data: {
        session_id: sessionId,
        map_id: mapId,
        team: team,
      },
    });
  };

  const sendVetoSwap = () => {
    console.log('🔄 [WebSocket] sendVetoSwap called:', {
      roomId,
      timestamp: new Date().toISOString(),
      isConnected: isConnected.value
    });
    
    send({
      type: 'veto:swap',
      data: {},
    });
  };

  const sendVetoStart = () => {
    console.log('▶️ [WebSocket] sendVetoStart called:', {
      roomId,
      timestamp: new Date().toISOString(),
      isConnected: isConnected.value
    });
    
    send({
      type: 'veto:start',
      data: {},
    });
  };

  const sendVetoReset = () => {
    console.log('🔄 [WebSocket] sendVetoReset called:', {
      roomId,
      timestamp: new Date().toISOString(),
      isConnected: isConnected.value
    });
    
    send({
      type: 'veto:reset',
      data: {},
    });
  };

  return {
    isConnected,
    messages,
    error,
    connect,
    disconnect,
    send,
    sendVetoBan,
    sendVetoPick,
    sendVetoSwap,
    sendVetoStart,
    sendVetoReset,
  };
}
