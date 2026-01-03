// Утилиты для работы с токенами и авторизацией

/**
 * Сохраняет токен в localStorage
 */
export function saveToken(token: string): void {
  localStorage.setItem('auth_token', token);
}

/**
 * Получает токен из localStorage
 */
export function getToken(): string | null {
  return localStorage.getItem('auth_token');
}

/**
 * Удаляет токен из localStorage
 */
export function removeToken(): void {
  localStorage.removeItem('auth_token');
}

/**
 * Проверяет, истек ли JWT токен
 * JWT токен имеет формат: header.payload.signature
 * payload содержит exp (expiration time) в формате Unix timestamp
 */
export function isTokenExpired(token: string): boolean {
  try {
    const parts = token.split('.');
    if (parts.length !== 3 || !parts[1]) {
      return true; // Невалидный формат токена
    }

    const payload = JSON.parse(atob(parts[1]));
    const exp = payload.exp;

    if (!exp) {
      return true; // Нет поля exp
    }

    // exp в секундах, Date.now() в миллисекундах
    const expirationTime = exp * 1000;
    const currentTime = Date.now();

    return currentTime >= expirationTime;
  } catch (error) {
    console.error('Error checking token expiration:', error);
    return true; // В случае ошибки считаем токен истекшим
  }
}

/**
 * Получает данные пользователя из токена (без проверки подписи)
 * Используется только для получения user ID
 */
export function getUserIdFromToken(token: string): number | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3 || !parts[1]) {
      return null;
    }

    const payload = JSON.parse(atob(parts[1]));
    return payload.user_id || payload.sub || null;
  } catch (error) {
    console.error('Error parsing token:', error);
    return null;
  }
}

/**
 * Проверяет валидность токена (формат и срок действия)
 */
export function isTokenValid(token: string | null): boolean {
  if (!token) {
    return false;
  }

  return !isTokenExpired(token);
}

/**
 * Проверяет токен и выполняет автоматический logout при истечении
 * Используется для периодической проверки токена
 */
export function checkTokenAndLogoutIfExpired(): boolean {
  const token = getToken();
  
  if (!token) {
    return false;
  }

  if (isTokenExpired(token)) {
    // Токен истек, удаляем его
    removeToken();
    
    // Вызываем logout через store (если доступен)
    // Это будет обработано в компонентах, которые используют store
    return false;
  }

  return true;
}

/**
 * Получает время истечения токена в миллисекундах
 */
export function getTokenExpirationTime(token: string): number | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3 || !parts[1]) {
      return null;
    }

    const payload = JSON.parse(atob(parts[1]));
    const exp = payload.exp;

    if (!exp) {
      return null;
    }

    return exp * 1000; // Конвертируем в миллисекунды
  } catch (error) {
    console.error('Error getting token expiration time:', error);
    return null;
  }
}

/**
 * Получает оставшееся время до истечения токена в миллисекундах
 */
export function getTokenTimeUntilExpiration(token: string): number | null {
  const expirationTime = getTokenExpirationTime(token);
  if (!expirationTime) {
    return null;
  }

  const currentTime = Date.now();
  const timeUntilExpiration = expirationTime - currentTime;

  return timeUntilExpiration > 0 ? timeUntilExpiration : 0;
}
