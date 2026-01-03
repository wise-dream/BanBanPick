import { apiClient, handleApiError } from './client';
import type {
  ProfileResponse,
  UpdateProfileRequest,
  VetoSessionResponse,
  RoomResponse,
  ApiError,
} from './types';

/**
 * Получение профиля пользователя
 */
export async function getProfile(): Promise<ProfileResponse> {
  try {
    const response = await apiClient.get<ProfileResponse>('/api/users/profile');
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Обновление профиля пользователя
 */
export async function updateProfile(
  data: UpdateProfileRequest
): Promise<ProfileResponse> {
  try {
    const response = await apiClient.put<ProfileResponse>(
      '/api/users/profile',
      data
    );
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение сессий пользователя
 */
export async function getSessions(): Promise<VetoSessionResponse[]> {
  try {
    const response = await apiClient.get<{ sessions: VetoSessionResponse[] }>(
      '/api/users/sessions'
    );
    return response.data.sessions;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение комнат пользователя
 */
export async function getRooms(): Promise<RoomResponse[]> {
  try {
    const response = await apiClient.get<{ rooms: RoomResponse[] }>(
      '/api/users/rooms'
    );
    return response.data.rooms;
  } catch (error) {
    throw handleApiError(error);
  }
}
