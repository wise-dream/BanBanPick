<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from '../composables/useI18n';
import { useAuthStore } from '../store/auth';
import { useRoomCreationStore } from '../store/roomCreation';
import StepIndicator from '../components/common/StepIndicator.vue';
import { Check } from 'lucide-vue-next';
import * as roomApi from '../services/api/roomService';
import type { ApiError } from '../services/api/types';
import { validateCreateRoomForm } from '../utils/validation';
import InputText from 'primevue/inputtext';
import Password from 'primevue/password';

const router = useRouter();
const { t } = useI18n();
const authStore = useAuthStore();
const roomCreationStore = useRoomCreationStore();

const roomName = ref('');
const roomPassword = ref('');
// const showCodeCopied = ref(false); // Не используется
const errors = ref<Record<string, string>>({});
const isLoading = ref(false);
const createdRoom = ref<{ id: number; code: string } | null>(null);

// Проверка авторизации и наличия данных
onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: '/ban/valorant' } });
    return;
  }
  
  if (!roomCreationStore.selectedPool || !roomCreationStore.bestOf) {
    // Если данные не выбраны, возвращаемся к началу
    router.push('/ban/valorant');
    return;
  }
});

const validateForm = (): boolean => {
  const validation = validateCreateRoomForm({
    name: roomName.value,
    password: roomPassword.value,
  });

  errors.value = validation.errors;
  return validation.isValid;
};

const handleCreateRoom = async () => {
  if (!validateForm()) {
    return;
  }

  isLoading.value = true;
  errors.value = {};

  try {
    // Преобразуем bestOf (1, 3, 5) в veto_type ('bo1', 'bo3', 'bo5')
    const vetoType = roomCreationStore.bestOf 
      ? `bo${roomCreationStore.bestOf}` as 'bo1' | 'bo3' | 'bo5'
      : undefined;
    
    const response = await roomApi.createRoom({
      name: roomName.value.trim(),
      type: 'private', // Все комнаты из этого процесса приватные
      game_id: 1, // Valorant
      map_pool_id: roomCreationStore.selectedPool?.id,
      veto_type: vetoType,
      max_participants: 2, // 2 участника (по одному от каждой команды)
    });

    createdRoom.value = {
      id: response.id,
      code: response.code
    };

    // Очищаем store после создания
    roomCreationStore.clear();
    
    // Переходим в созданную комнату
    setTimeout(() => {
      router.push(`/room/${createdRoom.value!.id}`);
    }, 1500);
  } catch (err) {
    const apiError = err as ApiError;
    if (apiError.code === 'HTTP_409') {
      errors.value.name = t('errors.roomNameExists');
    } else {
      createError.value = apiError.message || t('rooms.createError');
    }
    console.error('Error creating room:', err);
  } finally {
    isLoading.value = false;
  }
};

// Функция copyRoomCode удалена, так как не используется

const handleBack = () => {
  router.push('/create-room/best-of');
};

const createError = ref<string | null>(null);
</script>

<template>
  <div class="create-room-final-page">
    <div class="page-container">
      <h1 class="page-title">{{ t('createRoom.title') }}</h1>
      
      <StepIndicator :current-step="3" :total-steps="3" />

      <div v-if="!createdRoom" class="creation-form">
        <div class="form-section">
          <h2 class="section-title">{{ t('createRoom.roomDetails') }}</h2>
          
          <div class="form-group">
            <label for="roomName" class="form-label">{{ t('rooms.roomName') }}</label>
            <InputText
              id="roomName"
              v-model="roomName"
              :class="{ 'p-invalid': errors.name }"
              :placeholder="t('rooms.roomNamePlaceholder')"
              maxlength="50"
              class="w-full"
            />
            <span v-if="errors.name" class="error-message">{{ errors.name }}</span>
          </div>

          <div class="form-group">
            <label for="roomPassword" class="form-label">
              {{ t('rooms.password') }} ({{ t('common.optional') }})
            </label>
            <Password
              id="roomPassword"
              v-model="roomPassword"
              :class="{ 'p-invalid': errors.password }"
              :placeholder="t('rooms.passwordPlaceholder')"
              :toggleMask="true"
              :feedback="false"
              inputClass="w-full"
              class="w-full"
            />
            <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
            <span class="password-hint">{{ t('rooms.passwordHint') }}</span>
          </div>
        </div>

        <div class="form-section">
          <h2 class="section-title">{{ t('createRoom.summary') }}</h2>
          
          <div class="summary-card">
            <div class="summary-item">
              <span class="summary-label">{{ t('mapPool.title') }}:</span>
              <span class="summary-value">{{ roomCreationStore.selectedPool?.name }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">{{ t('bestOf.title') }}:</span>
              <span class="summary-value">{{ t('createRoomFinal.bestOfLabel', { value: roomCreationStore.bestOf }) }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">{{ t('rooms.maxParticipants') }}:</span>
              <span class="summary-value">2 ({{ t('rooms.maxParticipantsInfo') }})</span>
            </div>
          </div>
        </div>

        <div v-if="createError" class="error-message-global">
          {{ createError }}
        </div>

        <div class="form-actions">
          <button class="btn btn-secondary" @click="handleBack" :disabled="isLoading">
            {{ t('common.back') }}
          </button>
          <button
            class="btn btn-primary"
            @click="handleCreateRoom"
            :disabled="isLoading"
          >
            <span v-if="!isLoading">{{ t('createRoom.createRoom') }}</span>
            <span v-else>{{ t('common.loading') }}</span>
          </button>
        </div>
      </div>

      <div v-else class="success-message">
        <div class="success-icon">
          <Check :size="48" />
        </div>
        <h2>{{ t('createRoom.success') }}</h2>
        <p>{{ t('createRoom.redirecting') }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.create-room-final-page {
  width: 100%;
  min-height: calc(100vh - 200px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
  padding: 2rem;
}

.page-container {
  width: 100%;
  max-width: 800px;
}

.page-title {
  text-align: center;
  color: white;
  font-size: 2.5rem;
  margin-bottom: 2rem;
  font-weight: 700;
}

.creation-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-section {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 2rem;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: white;
  margin: 0 0 1.5rem 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.form-group:last-child {
  margin-bottom: 0;
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

.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input-wrapper .form-input {
  padding-right: 3rem;
}

.password-toggle {
  position: absolute;
  right: 0.75rem;
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  transition: color 0.2s;
}

.password-toggle:hover {
  color: white;
}

.password-hint {
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 0.25rem;
}

.error-message {
  font-size: 0.85rem;
  color: #ef4444;
}

.error-message-global {
  background: rgba(239, 68, 68, 0.2);
  border: 1px solid rgba(239, 68, 68, 0.5);
  color: #ef4444;
  padding: 1rem;
  border-radius: 8px;
  text-align: center;
}

.code-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.code-display {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.code-label {
  font-size: 0.9rem;
  color: rgba(255, 255, 255, 0.7);
}

.code-box {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  padding: 1rem;
  flex-wrap: wrap;
}

.code-value {
  font-family: monospace;
  font-size: 1.5rem;
  font-weight: 700;
  color: white;
  letter-spacing: 0.2em;
  flex: 1;
  min-width: 150px;
}

.copy-btn,
.regenerate-btn {
  background: rgba(102, 126, 234, 0.2);
  border: 1px solid rgba(102, 126, 234, 0.5);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.copy-btn:hover,
.regenerate-btn:hover {
  background: rgba(102, 126, 234, 0.4);
  border-color: #667eea;
}

.copied-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: #4caf50;
  font-weight: 600;
}

.code-hint {
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.6);
  margin: 0;
}

.summary-card {
  background: rgba(0, 0, 0, 0.9);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.summary-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.summary-label {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.9rem;
}

.summary-value {
  color: white;
  font-weight: 600;
  font-size: 0.95rem;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1rem;
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

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-secondary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.2);
}

.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.success-message {
  text-align: center;
  padding: 3rem 2rem;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
}

.success-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 1.5rem;
  background: rgba(76, 175, 80, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #4caf50;
}

.success-message h2 {
  color: white;
  font-size: 2rem;
  margin: 0 0 1rem 0;
}

.success-message p {
  color: rgba(255, 255, 255, 0.7);
  font-size: 1.1rem;
}

@media (max-width: 768px) {
  .page-title {
    font-size: 2rem;
  }

  .form-section {
    padding: 1.5rem;
  }

  .code-box {
    flex-direction: column;
    align-items: stretch;
  }

  .code-value {
    text-align: center;
  }

  .form-actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }
}
</style>
