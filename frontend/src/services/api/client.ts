import axios from 'axios';
import type { AxiosInstance, AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import { getToken, removeToken, isTokenExpired } from './auth';
import { useAuthStore } from '../../store/auth';
import type { ApiResponse, ApiError } from './types';
import { getErrorMessage, logError, isAuthError, canRetry } from './errorHandler';

// Базовый URL API из переменных окружения
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

/**
 * Создает и настраивает экземпляр axios клиента
 */
function createApiClient(): AxiosInstance {
  const client = axios.create({
    baseURL: API_BASE_URL,
    timeout: 30000, // 30 секунд
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor - добавляет токен к каждому запросу
  client.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      const token = getToken();

      // Проверяем, не истек ли токен
      if (token && isTokenExpired(token)) {
        // Токен истек, удаляем его и делаем logout
        removeToken();
        const authStore = useAuthStore();
        authStore.logout();
        
        // Перенаправляем на страницу входа, если не на ней уже
        if (window.location.pathname !== '/login') {
          window.location.href = '/login';
        }
        
        return Promise.reject(new Error('Token expired'));
      }

      // Добавляем токен в заголовки, если он есть
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }

      return config;
    },
    (error: AxiosError) => {
      return Promise.reject(error);
    }
  );

      // Response interceptor - обрабатывает ошибки
  client.interceptors.response.use(
    (response: AxiosResponse) => {
      return response;
    },
    (error: AxiosError<any>) => {
      // Обработка ошибок
      if (error.response) {
        const status = error.response.status;
        const data = error.response.data;

        // 401 Unauthorized - токен невалиден или истек
        if (status === 401) {
          removeToken();
          const authStore = useAuthStore();
          authStore.logout();

          // Перенаправляем на страницу входа
          if (window.location.pathname !== '/login') {
            window.location.href = '/login';
          }
        }

        // Преобразуем ошибку в стандартный формат
        // Бэкенд может возвращать ошибки в формате { error: "..." } или { message: "..." }
        let errorMessage = 'Произошла ошибка';
        if (data) {
          if (typeof data === 'string') {
            errorMessage = data;
          } else if (data.message) {
            errorMessage = data.message;
          } else if (data.error) {
            errorMessage = data.error;
          } else if (error.message) {
            errorMessage = error.message;
          }
        } else if (error.message) {
          errorMessage = error.message;
        }

        const apiError: ApiError = {
          code: data?.code || `HTTP_${status}`,
          message: errorMessage,
          details: data?.details,
        };

        return Promise.reject(apiError);
      }

      // Ошибка сети или таймаут
      if (error.request) {
        const networkError: ApiError = {
          code: 'NETWORK_ERROR',
          message: 'Ошибка сети. Проверьте подключение к интернету.',
        };
        return Promise.reject(networkError);
      }

      // Другая ошибка
      const unknownError: ApiError = {
        code: 'UNKNOWN_ERROR',
        message: error.message || 'Произошла неизвестная ошибка',
      };
      return Promise.reject(unknownError);
    }
  );

  return client;
}

// Экспортируем настроенный клиент
export const apiClient = createApiClient();

/**
 * Вспомогательная функция для обработки ответов API
 */
export function handleApiResponse<T>(response: AxiosResponse<ApiResponse<T>>): T {
  if (response.data.data !== undefined) {
    return response.data.data;
  }
  // Если ответ не в формате ApiResponse, возвращаем весь response.data
  return response.data as unknown as T;
}

/**
 * Вспомогательная функция для обработки ошибок API
 */
export function handleApiError(error: unknown): ApiError {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<ApiError>;
    if (axiosError.response?.data) {
      const apiError = axiosError.response.data;
      logError(apiError, 'API');
      return apiError;
    }
    const networkError: ApiError = {
      code: 'NETWORK_ERROR',
      message: axiosError.message || 'Ошибка сети',
    };
    logError(networkError, 'Network');
    return networkError;
  }

  if (error instanceof Error) {
    const unknownError: ApiError = {
      code: 'UNKNOWN_ERROR',
      message: error.message,
    };
    logError(unknownError, 'Unknown');
    return unknownError;
  }

  const defaultError: ApiError = {
    code: 'UNKNOWN_ERROR',
    message: 'Произошла неизвестная ошибка',
  };
  logError(defaultError, 'Unknown');
  return defaultError;
}
