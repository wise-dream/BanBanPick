<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../store/auth';
import { useI18n } from '../composables/useI18n';
import { login } from '../services/api/authService';
import type { ApiError } from '../services/api/types';
import { validateLoginForm } from '../utils/validation';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const { t } = useI18n();

const email = ref('');
const password = ref('');
const errors = ref<Record<string, string>>({});
const isLoading = ref(false);

const validateForm = (): boolean => {
  const validation = validateLoginForm({
    email: email.value,
    password: password.value,
  });

  errors.value = validation.errors;
  return validation.isValid;
};

const handleSubmit = async () => {
  if (!validateForm()) {
    return;
  }

  isLoading.value = true;
  errors.value = {};

  try {
    const response = await login({
      email: email.value.trim(),
      password: password.value,
    });

    // Сохраняем токен и данные пользователя
    authStore.setAuth(
      {
        id: response.user.id,
        email: response.user.email,
        username: response.user.username,
      },
      response.token
    );

    // Редирект на сохраненный путь или на главную
    const redirect = route.query.redirect as string;
    router.push(redirect || '/');
  } catch (error) {
    const apiError = error as ApiError;
    
    // Обработка различных типов ошибок
    if (apiError.code === 'HTTP_401' || apiError.message.includes('invalid')) {
      errors.value.submit = t('auth.invalidCredentials');
    } else if (apiError.code === 'NETWORK_ERROR') {
      errors.value.submit = 'Ошибка сети. Проверьте подключение к интернету.';
    } else {
      errors.value.submit = apiError.message || t('auth.invalidCredentials');
    }
  } finally {
    isLoading.value = false;
  }
};

</script>

<template>
  <div class="login-page">
    <div class="page-container">
      <div class="auth-card">
        <h1 class="auth-title">{{ t('auth.loginTitle') }}</h1>
        <p class="auth-subtitle">{{ t('auth.loginSubtitle') }}</p>

        <form @submit.prevent="handleSubmit" class="auth-form">
          <div class="form-group">
            <label for="email" class="form-label">{{ t('auth.email') }}</label>
            <input
              id="email"
              v-model="email"
              type="email"
              class="form-input"
              :class="{ 'error': errors.email }"
              :placeholder="t('auth.emailPlaceholder')"
              autocomplete="email"
            />
            <span v-if="errors.email" class="error-message">{{ errors.email }}</span>
          </div>

          <div class="form-group">
            <label for="password" class="form-label">{{ t('auth.password') }}</label>
            <input
              id="password"
              v-model="password"
              type="password"
              class="form-input"
              :class="{ 'error': errors.password }"
              :placeholder="t('auth.passwordPlaceholder')"
              autocomplete="current-password"
            />
            <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
          </div>

          <span v-if="errors.submit" class="error-message submit-error">{{ errors.submit }}</span>

          <button
            type="submit"
            class="btn btn-primary btn-submit"
            :disabled="isLoading"
          >
            <span v-if="!isLoading">{{ t('auth.login') }}</span>
            <span v-else>{{ t('auth.loading') }}</span>
          </button>
        </form>

        <div class="auth-footer">
          <p>
            {{ t('auth.noAccount') }}
            <router-link to="/register" class="auth-link">{{ t('auth.register') }}</router-link>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: calc(100vh - 200px);
  padding: 2rem;
  position: relative;
  z-index: 1;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-container {
  width: 100%;
  max-width: 450px;
}

.auth-card {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 2.5rem;
  width: 100%;
}

.auth-title {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  color: white;
  text-align: center;
}

.auth-subtitle {
  font-size: 0.95rem;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 2rem;
  text-align: center;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
}

.form-input {
  width: 100%;
  padding: 0.75rem 1rem;
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: white;
  font-size: 1rem;
  transition: all 0.2s;
}

.form-input::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
  background: rgba(255, 255, 255, 0.15);
}

.form-input.error {
  border-color: #ef4444;
}

.error-message {
  font-size: 0.85rem;
  color: #ef4444;
  margin-top: 0.25rem;
}

.submit-error {
  text-align: center;
  margin-top: -0.5rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-submit {
  width: 100%;
  margin-top: 0.5rem;
}

.auth-footer {
  margin-top: 2rem;
  text-align: center;
}

.auth-footer p {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.9rem;
}

.auth-link {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.auth-link:hover {
  color: #764ba2;
}

@media (max-width: 768px) {
  .auth-card {
    padding: 2rem 1.5rem;
  }

  .auth-title {
    font-size: 1.75rem;
  }
}
</style>
