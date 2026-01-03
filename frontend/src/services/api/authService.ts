import { apiClient, handleApiError } from './client';
import type {
  RegisterRequest,
  LoginRequest,
  AuthResponse,
  UserResponse,
} from './types';

/**
 * Регистрация нового пользователя
 */
export async function register(
  data: RegisterRequest
): Promise<AuthResponse> {
  try {
    const response = await apiClient.post<AuthResponse>('/api/auth/register', data);
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Вход пользователя
 */
export async function login(data: LoginRequest): Promise<AuthResponse> {
  try {
    const response = await apiClient.post<AuthResponse>('/api/auth/login', data);
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Получение текущего пользователя
 */
export async function getCurrentUser(): Promise<UserResponse> {
  try {
    const response = await apiClient.get<UserResponse>('/api/auth/me');
    return response.data;
  } catch (error) {
    throw handleApiError(error);
  }
}

/**
 * Выход пользователя (на клиенте)
 * На бэкенде нет отдельного endpoint для logout, так как используется stateless JWT
 */
export function logout(): void {
  // Токен удаляется через store
  // Эта функция может быть использована для дополнительной логики при выходе
}
