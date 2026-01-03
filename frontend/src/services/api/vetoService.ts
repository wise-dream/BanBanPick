import { apiClient, handleApiError } from './client';
import type {
  CreateVetoSessionRequest,
  VetoSessionResponse,
  NextActionResponse,
} from './types';

/**
 * Создание новой сессии вето
 */
export async function createSession(
  data: CreateVetoSessionRequest
): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.post<VetoSessionResponse>(
      '/api/veto/sessions',
      data
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение сессии по ID
 */
export async function getSession(id: number): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.get<VetoSessionResponse>(
      `/api/veto/sessions/${id}`
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение сессии по share token
 */
export async function getSessionByToken(
  token: string
): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.get<VetoSessionResponse>(
      `/api/veto/sessions/share/${token}`
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение следующего доступного действия
 */
export async function getNextAction(
  sessionId: number
): Promise<NextActionResponse> {
  try {
    const response = await apiClient.get<NextActionResponse>(
      `/api/veto/sessions/${sessionId}/next-action`
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Бан карты
 */
export async function banMap(
  sessionId: number,
  mapId: number,
  team: 'A' | 'B'
): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.post<VetoSessionResponse>(
      `/api/veto/sessions/${sessionId}/ban`,
      {
        map_id: mapId,
        team: team,
      }
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Выбор карты (pick)
 */
export async function pickMap(
  sessionId: number,
  mapId: number,
  team: 'A' | 'B'
): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.post<VetoSessionResponse>(
      `/api/veto/sessions/${sessionId}/pick`,
      {
        map_id: mapId,
        team: team,
      }
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Выбор стороны (attack/defence)
 */
export async function selectSide(
  sessionId: number,
  side: 'attack' | 'defence'
): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.post<VetoSessionResponse>(
      `/api/veto/sessions/${sessionId}/select-side`,
      {
        side: side,
      }
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Старт сессии (изменяет статус с not_started на in_progress)
 */
export async function startSession(
  sessionId: number
): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.post<VetoSessionResponse>(
      `/api/veto/sessions/${sessionId}/start`
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Сброс сессии
 */
export async function resetSession(
  sessionId: number
): Promise<VetoSessionResponse> {
  try {
    const response = await apiClient.post<VetoSessionResponse>(
      `/api/veto/sessions/${sessionId}/reset`
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}
