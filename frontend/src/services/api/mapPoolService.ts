import { apiClient, handleApiResponse, handleApiError } from './client';
import type {
  MapPoolResponse,
  CreateCustomMapPoolRequest,
  ApiError,
} from './types';
import type { MapPool } from '../../types';
import { apiCache, generateCacheKey, withCache } from './cache';

/**
 * Получение всех пулов для игры (с кешированием)
 */
export async function getPools(gameId: number = 1): Promise<MapPoolResponse[]> {
  const cacheKey = generateCacheKey('map-pools', { gameId });
  const cached = apiCache.get<MapPoolResponse[]>(cacheKey);
  
  if (cached !== null) {
    return cached;
  }

  try {
    const response = await apiClient.get<MapPoolResponse[]>(
      `/api/map-pools/games/${gameId}`
    );
    apiCache.set(cacheKey, response.data, { ttl: 10 * 60 * 1000 }); // 10 минут
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение пула по ID (с кешированием)
 */
export async function getPool(id: number): Promise<MapPoolResponse> {
  const cacheKey = generateCacheKey('map-pool', { id });
  const cached = apiCache.get<MapPoolResponse>(cacheKey);
  
  if (cached !== null) {
    return cached;
  }

  try {
    const response = await apiClient.get<MapPoolResponse>(
      `/api/map-pools/${id}`
    );
    apiCache.set(cacheKey, response.data, { ttl: 10 * 60 * 1000 }); // 10 минут
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Создание кастомного пула
 */
export async function createCustomPool(
  data: CreateCustomMapPoolRequest,
  gameId: number = 1
): Promise<MapPoolResponse> {
  try {
    const response = await apiClient.post<MapPoolResponse>(
      `/api/map-pools?game_id=${gameId}`,
      data
    );
    // Инвалидируем кеш пулов для этой игры
    apiCache.invalidate(`map-pools:*gameId=${gameId}*`);
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Удаление кастомного пула
 */
export async function deletePool(id: number): Promise<void> {
  try {
    await apiClient.delete(`/api/map-pools/${id}`);
    // Инвалидируем кеш для этого пула и списка пулов
    apiCache.delete(generateCacheKey('map-pool', { id }));
    apiCache.invalidate('map-pools:*');
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Преобразование MapPoolResponse в MapPool (для совместимости)
 */
export function mapPoolResponseToMapPool(response: MapPoolResponse): MapPool {
  return {
    id: response.id,
    gameId: response.game_id,
    name: response.name,
    type: response.type as 'all' | 'competitive' | 'custom',
    isSystem: response.is_system,
    maps: response.maps.map(map => ({
      id: map.id,
      name: map.name,
      slug: map.slug,
      imageUrl: map.image_url,
      isCompetitive: map.is_competitive,
    })),
  };
}
