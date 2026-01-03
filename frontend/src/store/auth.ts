import { defineStore } from 'pinia';
import { ref, computed, onMounted } from 'vue';
import { getCurrentUser } from '../services/api/authService';
import { getToken, isTokenValid, checkTokenAndLogoutIfExpired, getTokenTimeUntilExpiration } from '../services/api/auth';
import type { ApiError } from '../services/api/types';

export interface User {
  id: number;
  email: string;
  username: string;
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(null);
  const isAuthenticated = computed(() => !!token.value && !!user.value && isTokenValid(token.value));

  function setAuth(userData: User, authToken: string) {
    user.value = userData;
    token.value = authToken;
    localStorage.setItem('auth_token', authToken);
    localStorage.setItem('user', JSON.stringify(userData));
  }

  function logout() {
    stopTokenCheck();
    user.value = null;
    token.value = null;
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user');
  }

  /**
   * Инициализация авторизации при загрузке приложения
   * Проверяет токен и загружает данные пользователя
   */
  async function initAuth() {
    const storedToken = getToken();
    
    if (!storedToken || !isTokenValid(storedToken)) {
      // Токен отсутствует или истек
      if (storedToken) {
        // Удаляем истекший токен
        logout();
      }
      return;
    }

    // Токен валиден, загружаем данные пользователя
    token.value = storedToken;
    
    // Пытаемся загрузить данные пользователя из localStorage
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser);
      } catch (error) {
        console.error('Error parsing stored user:', error);
      }
    }

    // Проверяем данные пользователя через API
    try {
      const currentUser = await getCurrentUser();
      user.value = {
        id: currentUser.id,
        email: currentUser.email,
        username: currentUser.username,
      };
      // Обновляем данные в localStorage
      localStorage.setItem('user', JSON.stringify(user.value));
      
      // Запускаем периодическую проверку токена
      startTokenCheck();
    } catch (error) {
      // Если не удалось загрузить пользователя, делаем logout
      console.error('Error loading current user:', error);
      const apiError = error as ApiError;
      if (apiError.code === 'HTTP_401') {
        logout();
      }
    }
  }

  let tokenCheckInterval: number | null = null;

  /**
   * Запускает периодическую проверку токена
   */
  function startTokenCheck() {
    // Останавливаем предыдущий интервал, если есть
    if (tokenCheckInterval) {
      clearInterval(tokenCheckInterval);
    }

    const storedToken = getToken();
    if (!storedToken) {
      return;
    }

    // Проверяем токен каждую минуту
    tokenCheckInterval = window.setInterval(() => {
      if (!checkTokenAndLogoutIfExpired()) {
        // Токен истек, делаем logout
        logout();
        if (tokenCheckInterval) {
          clearInterval(tokenCheckInterval);
          tokenCheckInterval = null;
        }
      }
    }, 60000); // 60 секунд
  }

  /**
   * Останавливает периодическую проверку токена
   */
  function stopTokenCheck() {
    if (tokenCheckInterval) {
      clearInterval(tokenCheckInterval);
      tokenCheckInterval = null;
    }
  }

  return {
    user,
    token,
    isAuthenticated,
    setAuth,
    logout,
    initAuth,
    startTokenCheck,
    stopTokenCheck
  };
});
