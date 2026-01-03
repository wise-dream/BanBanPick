import type { ApiError } from './types';

/**
 * Классификация ошибок API
 */
export const ErrorType = {
  NETWORK: 'NETWORK',
  VALIDATION: 'VALIDATION',
  AUTHENTICATION: 'AUTHENTICATION',
  AUTHORIZATION: 'AUTHORIZATION',
  NOT_FOUND: 'NOT_FOUND',
  CONFLICT: 'CONFLICT',
  SERVER: 'SERVER',
  UNKNOWN: 'UNKNOWN',
} as const;

export type ErrorType = typeof ErrorType[keyof typeof ErrorType];

/**
 * Классифицирует ошибку по типу
 */
export function classifyError(error: ApiError): ErrorType {
  const code = error.code || '';

  if (code === 'NETWORK_ERROR') {
    return ErrorType.NETWORK;
  }

  if (code.startsWith('HTTP_')) {
    const status = parseInt(code.replace('HTTP_', ''));
    if (status >= 400 && status < 500) {
      if (status === 401) {
        return ErrorType.AUTHENTICATION;
      }
      if (status === 403) {
        return ErrorType.AUTHORIZATION;
      }
      if (status === 404) {
        return ErrorType.NOT_FOUND;
      }
      if (status === 409) {
        return ErrorType.CONFLICT;
      }
      if (status === 422 || status === 400) {
        return ErrorType.VALIDATION;
      }
    }
    if (status >= 500) {
      return ErrorType.SERVER;
    }
  }

  return ErrorType.UNKNOWN;
}

/**
 * Преобразует ошибку в понятное сообщение для пользователя
 */
export function getErrorMessage(error: ApiError | Error | unknown): string {
  if (!error) {
    return 'Произошла неизвестная ошибка';
  }

  // ApiError
  if (typeof error === 'object' && 'message' in error) {
    const apiError = error as ApiError;
    const errorType = classifyError(apiError);

    switch (errorType) {
      case ErrorType.NETWORK:
        return 'Ошибка сети. Проверьте подключение к интернету.';
      case ErrorType.AUTHENTICATION:
        return 'Необходима авторизация. Пожалуйста, войдите в систему.';
      case ErrorType.AUTHORIZATION:
        return 'У вас нет прав для выполнения этого действия.';
      case ErrorType.NOT_FOUND:
        return 'Запрашиваемый ресурс не найден.';
      case ErrorType.CONFLICT:
        return apiError.message || 'Конфликт данных. Возможно, ресурс уже существует.';
      case ErrorType.VALIDATION:
        return apiError.message || 'Ошибка валидации данных.';
      case ErrorType.SERVER:
        return 'Ошибка сервера. Попробуйте позже.';
      default:
        return apiError.message || 'Произошла ошибка';
    }
  }

  // Standard Error
  if (error instanceof Error) {
    return error.message;
  }

  return 'Произошла неизвестная ошибка';
}

/**
 * Логирует ошибку
 */
export function logError(error: ApiError | Error | unknown, context?: string): void {
  const message = getErrorMessage(error);
  const errorType = error && typeof error === 'object' && 'code' in error
    ? classifyError(error as ApiError)
    : ErrorType.UNKNOWN;

  console.error(`[${errorType}]`, context ? `${context}: ${message}` : message, error);
}

/**
 * Проверяет, является ли ошибка сетевой
 */
export function isNetworkError(error: ApiError | Error | unknown): boolean {
  if (error && typeof error === 'object' && 'code' in error) {
    return classifyError(error as ApiError) === ErrorType.NETWORK;
  }
  return false;
}

/**
 * Проверяет, является ли ошибка ошибкой авторизации
 */
export function isAuthError(error: ApiError | Error | unknown): boolean {
  if (error && typeof error === 'object' && 'code' in error) {
    const errorType = classifyError(error as ApiError);
    return errorType === ErrorType.AUTHENTICATION || errorType === ErrorType.AUTHORIZATION;
  }
  return false;
}

/**
 * Проверяет, можно ли повторить запрос
 */
export function canRetry(error: ApiError | Error | unknown): boolean {
  if (error && typeof error === 'object' && 'code' in error) {
    const errorType = classifyError(error as ApiError);
    return errorType === ErrorType.NETWORK || errorType === ErrorType.SERVER;
  }
  return false;
}
