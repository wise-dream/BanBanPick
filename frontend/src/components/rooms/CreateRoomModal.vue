<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from '../../composables/useI18n';
import { X, Eye, EyeOff } from 'lucide-vue-next';
import { getAllPools } from '../../services/mapPoolService';

interface Props {
  show: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  created: [room: { name: string; type: 'public' | 'private'; mapPoolId?: number; bestOf?: 'bo1' | 'bo3' | 'bo5'; password?: string }];
}>();

const { t } = useI18n();

const roomName = ref('');
const roomType = ref<'public' | 'private'>('public');
const selectedPoolId = ref<number | null>(null);
const selectedBestOf = ref<'bo1' | 'bo3' | 'bo5'>('bo1');
const roomPassword = ref('');
const showPassword = ref(false);
const errors = ref<Record<string, string>>({});
const isLoadingPools = ref(false);
const availablePools = ref<Array<{ id: number; name: string }>>([]);

// Загружаем пулы при монтировании
onMounted(async () => {
  await loadPools();
});

const loadPools = async () => {
  isLoadingPools.value = true;
  try {
    const pools = await getAllPools();
    availablePools.value = pools.map(pool => ({ id: pool.id, name: pool.name }));
  } catch (error) {
    console.error('Failed to load pools:', error);
    availablePools.value = [];
  } finally {
    isLoadingPools.value = false;
  }
};

const validateForm = (): boolean => {
  errors.value = {};

  if (!roomName.value.trim()) {
    errors.value.name = t('rooms.nameRequired');
    return false;
  }

  if (roomName.value.length > 50) {
    errors.value.name = t('rooms.nameMaxLength');
    return false;
  }

  // Пароль опционален, но если указан, должен быть минимум 4 символа
  if (roomType.value === 'private' && roomPassword.value && roomPassword.value.length < 4) {
    errors.value.password = t('rooms.passwordMinLength');
    return false;
  }

  return true;
};

const handleSubmit = () => {
  if (!validateForm()) {
    return;
  }

  emit('created', {
    name: roomName.value.trim(),
    type: roomType.value,
    mapPoolId: selectedPoolId.value || undefined,
    bestOf: selectedBestOf.value,
    password: roomType.value === 'private' && roomPassword.value ? roomPassword.value : undefined
  });

  // Reset form
  roomName.value = '';
  roomType.value = 'public';
  selectedPoolId.value = null;
  roomPassword.value = '';
  showPassword.value = false;
  errors.value = {};
};

const handleClose = () => {
  emit('close');
  // Reset form on close
  roomName.value = '';
  roomType.value = 'public';
  selectedPoolId.value = null;
  selectedBestOf.value = 'bo1';
  roomPassword.value = '';
  showPassword.value = false;
  errors.value = {};
};
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click.self="handleClose">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h2>{{ t('rooms.createRoom') }}</h2>
            <button class="close-btn" @click="handleClose">
              <X :size="20" />
            </button>
          </div>

          <form @submit.prevent="handleSubmit" class="modal-form">
            <div class="form-group">
              <label for="roomName" class="form-label">{{ t('rooms.roomName') }}</label>
              <input
                id="roomName"
                v-model="roomName"
                type="text"
                class="form-input"
                :class="{ error: errors.name }"
                :placeholder="t('rooms.roomNamePlaceholder')"
                maxlength="50"
              />
              <span v-if="errors.name" class="error-message">{{ errors.name }}</span>
            </div>

            <div class="form-group">
              <label for="roomType" class="form-label">{{ t('rooms.roomType') }}</label>
              <select
                id="roomType"
                v-model="roomType"
                class="form-input"
              >
                <option value="public">{{ t('rooms.type.public') }}</option>
                <option value="private">{{ t('rooms.type.private') }}</option>
              </select>
            </div>

            <div v-if="roomType === 'private'" class="form-group">
              <label for="roomPassword" class="form-label">
                {{ t('rooms.password') }} ({{ t('common.optional') }})
              </label>
              <div class="password-input-wrapper">
                <input
                  id="roomPassword"
                  v-model="roomPassword"
                  :type="showPassword ? 'text' : 'password'"
                  class="form-input"
                  :class="{ error: errors.password }"
                  :placeholder="t('rooms.passwordPlaceholder')"
                />
                <button
                  type="button"
                  class="password-toggle"
                  @click="showPassword = !showPassword"
                  :title="showPassword ? t('rooms.hidePassword') : t('rooms.showPassword')"
                >
                  <Eye v-if="!showPassword" :size="18" />
                  <EyeOff v-else :size="18" />
                </button>
              </div>
              <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
              <span class="password-hint">{{ t('rooms.passwordHint') }}</span>
            </div>

            <div class="form-group">
              <label for="mapPool" class="form-label">{{ t('rooms.mapPool') }} ({{ t('common.optional') }})</label>
              <select
                id="mapPool"
                v-model="selectedPoolId"
                class="form-input"
                :disabled="isLoadingPools"
              >
                <option :value="null">{{ isLoadingPools ? t('common.loading') : t('rooms.noMapPool') }}</option>
                <option
                  v-for="pool in availablePools"
                  :key="pool.id"
                  :value="pool.id"
                >
                  {{ pool.name }}
                </option>
              </select>
            </div>

            <div class="form-group">
              <label for="bestOf" class="form-label">{{ t('rooms.bestOf') }}</label>
              <select
                id="bestOf"
                v-model="selectedBestOf"
                class="form-input"
              >
                <option value="bo1">Best of 1</option>
                <option value="bo3">Best of 3</option>
                <option value="bo5">Best of 5</option>
              </select>
            </div>

            <div class="form-group">
              <label class="form-label">{{ t('rooms.maxParticipants') }}</label>
              <div class="info-text">
                {{ t('rooms.maxParticipantsInfo') }}
              </div>
            </div>

            <div class="form-actions">
              <button type="button" class="btn btn-secondary" @click="handleClose">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" class="btn btn-primary">
                {{ t('rooms.create') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  backdrop-filter: blur(10px);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: white;
  font-weight: 600;
}

.close-btn {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.close-btn:hover {
  color: white;
}

.modal-form {
  padding: 1.5rem;
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
}

.info-text {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.9rem;
  padding: 0.5rem;
  background: rgba(102, 126, 234, 0.1);
  border-radius: 6px;
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

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 0.5rem;
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

.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.3s;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.9);
}
</style>
