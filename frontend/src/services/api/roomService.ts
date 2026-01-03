import { apiClient, handleApiResponse, handleApiError } from './client';
import type {
  RoomResponse,
  CreateRoomRequest,
  JoinRoomRequest,
  ParticipantResponse,
  ApiError,
  PaginatedResponse,
} from './types';
import { apiCache, generateCacheKey } from './cache';

/**
 * Получение списка комнат (с кешированием)
 */
export async function getRooms(
  limit: number = 20,
  offset: number = 0,
  type?: 'public' | 'private'
): Promise<PaginatedResponse<RoomResponse>> {
  const cacheKey = generateCacheKey('rooms-list', { limit, offset, type });
  const cached = apiCache.get<PaginatedResponse<RoomResponse>>(cacheKey);
  
  if (cached !== null) {
    return cached;
  }

  try {
    let url = `/api/rooms?limit=${limit}&offset=${offset}`;
    if (type) {
      url += `&type=${type}`;
    }
    const response = await apiClient.get<PaginatedResponse<RoomResponse>>(url);
    apiCache.set(cacheKey, response.data, { ttl: 30 * 1000 }); // 30 секунд (комнаты часто меняются)
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение комнаты по ID
 */
export async function getRoom(id: number): Promise<RoomResponse> {
  try {
    const response = await apiClient.get<RoomResponse>(`/api/rooms/${id}`);
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение комнаты по veto session ID
 */
export async function getRoomBySessionId(sessionId: number): Promise<RoomResponse | null> {
  try {
    const response = await apiClient.get<RoomResponse>(`/api/rooms/by-session/${sessionId}`);
    return response.data;
  } catch (error) {
    // Если комната не найдена, возвращаем null
    if ((error as any)?.response?.status === 404) {
      return null;
    }
    throw handleApiError(error);
  }
}

/**
 * Создание комнаты
 */
export async function createRoom(
  data: CreateRoomRequest
): Promise<RoomResponse> {
  try {
    const response = await apiClient.post<RoomResponse>('/api/rooms', data);
    // Инвалидируем кеш списка комнат
    apiCache.invalidate('rooms-list:*');
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Присоединение к комнате
 */
export async function joinRoom(
  id: number,
  password?: string
): Promise<RoomResponse> {
  try {
    // Отправляем password только если он передан и не пустой
    const payload: { password?: string } = {};
    if (password && password.trim() !== '') {
      payload.password = password;
    }
    
    const response = await apiClient.post<RoomResponse>(
      `/api/rooms/${id}/join`,
      payload
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Выход из комнаты
 */
export async function leaveRoom(id: number): Promise<void> {
  try {
    await apiClient.post(`/api/rooms/${id}/leave`);
    // Инвалидируем кеш для этой комнаты и списка комнат
    apiCache.delete(generateCacheKey('room', { id }));
    apiCache.invalidate('rooms-list:*');
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Удаление комнаты
 */
export async function deleteRoom(id: number): Promise<void> {
  try {
    await apiClient.delete(`/api/rooms/${id}`);
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение списка участников комнаты
 */
export async function getParticipants(
  id: number
): Promise<ParticipantResponse[]> {
  try {
    const response = await apiClient.get<{ data: ParticipantResponse[] }>(
      `/api/rooms/${id}/participants`
    );
    return response.data.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Обновление комнаты
 */
export async function updateRoom(
  id: number,
  data: import('./types').UpdateRoomRequest
): Promise<RoomResponse> {
  try {
    const response = await apiClient.put<RoomResponse>(`/api/rooms/${id}`, data);
    // Инвалидируем кеш комнаты и списка комнат
    apiCache.delete(generateCacheKey('room', { id }));
    apiCache.invalidate('rooms-list:*');
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Преобразование RoomResponse в Room (для совместимости)
 */
import type { Room, RoomParticipant } from '../../types';

export function roomResponseToRoom(response: RoomResponse): Room {
  return {
    id: response.id,
    ownerId: response.owner_id,
    name: response.name,
    code: response.code,
    type: response.type as 'public' | 'private',
    status: response.status as 'waiting' | 'active' | 'finished',
    gameId: response.game_id,
    mapPoolId: response.map_pool_id || undefined,
    vetoSessionId: response.veto_session_id || undefined,
    maxParticipants: response.max_participants,
    createdAt: response.created_at,
    updatedAt: response.updated_at,
    participants: response.participants?.map(p => ({
      id: p.id,
      roomId: response.id,
      userId: p.user_id,
      username: p.username, // Никнейм пользователя
      role: p.role as 'owner' | 'member',
      joinedAt: p.joined_at,
    })) || [],
  };
}
