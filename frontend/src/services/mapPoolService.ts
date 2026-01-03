import type { MapPool, Map as GameMap } from '../types';
import * as mapPoolApi from './api/mapPoolService';

const STORAGE_KEY = 'custom_map_pools';

// Кеш для всех карт (получаются из первого системного пула через API)
let cachedMaps: GameMap[] | null = null;

/**
 * Получить все карты (из API или из кеша)
 */
export async function getAllMaps(): Promise<GameMap[]> {
  if (cachedMaps) {
    return cachedMaps;
  }

  try {
    // Получаем системные пулы и берем карты из первого пула
    const pools = await mapPoolApi.getPools(1); // Valorant game_id = 1
    const systemPool = pools.find(p => p.is_system && p.maps.length > 0);
    
    if (systemPool && systemPool.maps.length > 0) {
      cachedMaps = systemPool.maps.map(m => ({
        id: m.id,
        name: m.name,
        slug: m.slug,
        imageUrl: m.image_url,
        isCompetitive: m.is_competitive,
      }));
      return cachedMaps;
    }
  } catch (error) {
    console.warn('Failed to load maps from API:', error);
  }

  // Fallback - пустой массив
  return [];
}

/**
 * Синхронная версия (использует кеш или возвращает пустой массив)
 */
export function getAllMapsSync(): GameMap[] {
  return cachedMaps || [];
}

// Получить системные пулы (fallback - пустой массив, так как данные теперь из API)
export function getSystemPools(): MapPool[] {
  // Данные системных пулов теперь приходят из API
  // Fallback возвращает пустой массив
  return [];
}

// Получить кастомные пулы из localStorage
export function getCustomPools(): MapPool[] {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) return [];
    
    const pools = JSON.parse(stored) as Array<{
      id: number;
      gameId: number;
      name: string;
      type: string;
      isSystem: boolean;
      mapIds: number[];
    }>;
    
    // Для fallback из localStorage нужны карты, но они теперь только из API
    // Возвращаем пулы без карт (maps будет пустым массивом)

    return pools.map(pool => ({
      id: pool.id,
      gameId: pool.gameId,
      name: pool.name,
      type: pool.type as 'all' | 'competitive' | 'custom',
      isSystem: pool.isSystem,
      maps: [] // Карты теперь только из API, в fallback возвращаем пустой массив
    }));
  } catch (error) {
    console.error('Error loading custom pools:', error);
    return [];
  }
}

// Получить все пулы (системные + кастомные)
// Использует API, но сохраняет fallback на localStorage для обратной совместимости
export async function getAllPools(): Promise<MapPool[]> {
  try {
    // Пытаемся загрузить через API
    const apiPools = await mapPoolApi.getPools(1); // Valorant game_id = 1
    return apiPools.map(mapPoolApi.mapPoolResponseToMapPool);
  } catch (error) {
    // Fallback на localStorage, если API недоступен
    console.warn('Failed to load pools from API, using localStorage fallback:', error);
    return [...getSystemPools(), ...getCustomPools()];
  }
}

// Синхронная версия для обратной совместимости (использует localStorage)
export function getAllPoolsSync(): MapPool[] {
  return [...getSystemPools(), ...getCustomPools()];
}

// Получить пул по ID
// Использует API, но имеет fallback на localStorage для обратной совместимости
export async function getPoolById(id: number): Promise<MapPool | null> {
  try {
    // Пытаемся загрузить через API
    const apiPool = await mapPoolApi.getPool(id);
    return mapPoolApi.mapPoolResponseToMapPool(apiPool);
  } catch (error) {
    // Fallback на localStorage, если API недоступен
    console.warn('Failed to load pool from API, using localStorage fallback:', error);
    const allPools = getAllPoolsSync();
    return allPools.find(pool => pool.id === id) || null;
  }
}

// Сохранить кастомный пул
// Использует API, но имеет fallback на localStorage
export async function saveCustomPool(
  pool: Omit<MapPool, 'id' | 'isSystem'>
): Promise<MapPool> {
  try {
    // Пытаемся сохранить через API
    const apiPool = await mapPoolApi.createCustomPool({
      name: pool.name,
      map_ids: pool.maps.map(m => m.id),
    }, pool.gameId);
    
    return mapPoolApi.mapPoolResponseToMapPool(apiPool);
  } catch (error) {
    // Fallback на localStorage, если API недоступен
    console.warn('Failed to save pool via API, using localStorage fallback:', error);
    const customPools = getCustomPools();
    const newId = customPools.length > 0 
      ? Math.max(...customPools.map(p => p.id)) + 1 
      : 100;

    const newPool: MapPool = {
      ...pool,
      id: newId,
      isSystem: false
    };

    const poolsToSave = [...customPools, {
      id: newPool.id,
      gameId: newPool.gameId,
      name: newPool.name,
      type: newPool.type,
      isSystem: false,
      mapIds: newPool.maps.map(m => m.id)
    }];

    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(poolsToSave));
    } catch (err) {
      console.error('Error saving custom pool to localStorage:', err);
      throw new Error('Failed to save custom pool');
    }

    return newPool;
  }
}

// Удалить кастомный пул
// Использует API, но имеет fallback на localStorage
export async function deleteCustomPool(id: number): Promise<boolean> {
  try {
    // Пытаемся удалить через API
    await mapPoolApi.deletePool(id);
    return true;
  } catch (error) {
    // Fallback на localStorage, если API недоступен
    console.warn('Failed to delete pool via API, using localStorage fallback:', error);
    const customPools = getCustomPools();
    const filtered = customPools.filter(pool => pool.id !== id);

    if (filtered.length === customPools.length) {
      return false; // Пул не найден
    }

    try {
      const poolsToSave = filtered.map(pool => ({
        id: pool.id,
        gameId: pool.gameId,
        name: pool.name,
        type: pool.type,
        isSystem: false,
        mapIds: pool.maps.map(m => m.id)
      }));
      localStorage.setItem(STORAGE_KEY, JSON.stringify(poolsToSave));
      return true;
    } catch (err) {
      console.error('Error deleting custom pool from localStorage:', err);
      return false;
    }
  }
}
