import { getToken } from '../api/auth';

export interface WebSocketMessage {
  type: string;
  data: any;
}

export interface RoomWebSocketOptions {
  onMessage?: (message: WebSocketMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (error: Event) => void;
}

export class RoomWebSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 2000; // Начальная задержка 2 секунды
  private reconnectTimer: number | null = null;
  private pingInterval: number | null = null;
  private options: RoomWebSocketOptions;
  private isConnecting = false;
  private isConnected = false;
  private shouldReconnect = true; // Флаг для предотвращения переподключения при явном disconnect

  constructor(roomId: number, options: RoomWebSocketOptions = {}) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Используем VITE_WS_URL или извлекаем хост из VITE_API_URL, иначе используем localhost:8080
    let wsHost: string;
    if (import.meta.env.VITE_WS_URL) {
      wsHost = import.meta.env.VITE_WS_URL;
    } else if (import.meta.env.VITE_API_URL) {
      // Извлекаем хост из API URL (например, http://localhost:8080 -> localhost:8080)
      const apiUrl = new URL(import.meta.env.VITE_API_URL);
      wsHost = apiUrl.host;
    } else {
      // Fallback на localhost:8080 для бэкенда
      wsHost = 'localhost:8080';
    }
    this.url = `${wsProtocol}//${wsHost}/ws/room/${roomId}`;
    this.options = options;
  }

  /**
   * Connect to WebSocket server
   */
  connect(): void {
    if (this.isConnecting || this.isConnected) {
      return;
    }

    // Включаем автоматическое переподключение при новом подключении
    this.shouldReconnect = true;
    this.isConnecting = true;
    const token = getToken();

    if (!token) {
      console.error('No token available for WebSocket connection');
      this.isConnecting = false;
      return;
    }

    try {
      // Add token to URL as query parameter
      const urlWithToken = `${this.url}?token=${encodeURIComponent(token)}`;
      this.ws = new WebSocket(urlWithToken);

      this.ws.onopen = () => {
        this.isConnecting = false;
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.reconnectDelay = 2000; // Сбрасываем задержку при успешном подключении
        this.startPing();
        if (this.options.onOpen) {
          this.options.onOpen();
        }
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          console.log('Received WebSocket message:', message.type, message.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error, 'Raw data:', event.data);
          // Отправляем ошибку обратно на сервер для логирования
          if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
              type: 'error',
              data: { message: 'Failed to parse message', raw: event.data }
            }));
          }
        }
      };

      this.ws.onclose = () => {
        this.isConnected = false;
        this.stopPing();
        if (this.options.onClose) {
          this.options.onClose();
        }
        // Переподключаемся только если это не было явное отключение
        if (this.shouldReconnect) {
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
        this.isConnecting = false;
        if (this.options.onError) {
          this.options.onError(error);
        }
      };
    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      this.isConnecting = false;
      this.attemptReconnect();
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    // Отключаем автоматическое переподключение
    this.shouldReconnect = false;
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.stopPing();
    if (this.ws) {
      // Устанавливаем обработчики в null перед закрытием, чтобы избежать вызова onclose
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      this.ws.onopen = null;
      
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close();
      }
      this.ws = null;
    }
    this.isConnected = false;
    this.isConnecting = false;
  }

  /**
   * Send a message to the server
   */
  send(message: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  /**
   * Handle incoming messages
   */
  private handleMessage(message: WebSocketMessage): void {
    // Handle pong responses
    if (message.type === 'pong') {
      return;
    }

    // Call custom message handler
    if (this.options.onMessage) {
      this.options.onMessage(message);
    }
  }

  /**
   * Start ping interval to keep connection alive
   */
  private startPing(): void {
    this.pingInterval = window.setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping', data: {} });
      }
    }, 30000); // Ping every 30 seconds
  }

  /**
   * Stop ping interval
   */
  private stopPing(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  /**
   * Attempt to reconnect to WebSocket server
   */
  private attemptReconnect(): void {
    // Не переподключаемся, если это было явное отключение
    if (!this.shouldReconnect) {
      return;
    }

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    // Exponential backoff: 1s, 2s, 4s, 8s, 16s, max 30s
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);

    this.reconnectTimer = window.setTimeout(() => {
      // Проверяем, что переподключение все еще нужно
      if (!this.shouldReconnect) {
        return;
      }
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      this.connect();
    }, this.reconnectDelay);
  }

  /**
   * Get connection status
   */
  get connected(): boolean {
    return this.isConnected && this.ws?.readyState === WebSocket.OPEN;
  }
}
