<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from '../../composables/useI18n';
import { X, Eye, EyeOff } from 'lucide-vue-next';

interface Props {
  show: boolean;
  roomName: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  submit: [password: string];
}>();

const { t } = useI18n();

const password = ref('');
const showPassword = ref(false);
const error = ref<string | null>(null);

const handleSubmit = () => {
  error.value = null;

  if (!password.value.trim()) {
    error.value = t('rooms.passwordRequired');
    return;
  }

  emit('submit', password.value.trim());
  password.value = '';
  showPassword.value = false;
};

const handleClose = () => {
  emit('close');
  password.value = '';
  showPassword.value = false;
  error.value = null;
};
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click.self="handleClose">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h2>{{ t('rooms.enterPassword') }}</h2>
            <button class="close-btn" @click="handleClose">
              <X :size="20" />
            </button>
          </div>

          <div class="modal-body">
            <p class="room-name">{{ roomName }}</p>
            <p class="password-description">{{ t('rooms.passwordRequiredForPrivate') }}</p>

            <form @submit.prevent="handleSubmit" class="password-form">
              <div class="form-group">
                <label for="password" class="form-label">{{ t('rooms.password') }}</label>
                <div class="password-input-wrapper">
                  <input
                    id="password"
                    v-model="password"
                    :type="showPassword ? 'text' : 'password'"
                    class="form-input"
                    :class="{ error: error }"
                    :placeholder="t('rooms.passwordPlaceholder')"
                    autofocus
                    @keyup.enter="handleSubmit"
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
                <span v-if="error" class="error-message">{{ error }}</span>
              </div>

              <div class="form-actions">
                <button type="button" class="btn btn-secondary" @click="handleClose">
                  {{ t('common.cancel') }}
                </button>
                <button type="submit" class="btn btn-primary">
                  {{ t('rooms.join') }}
                </button>
              </div>
            </form>
          </div>
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
  max-width: 400px;
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

.modal-body {
  padding: 1.5rem;
}

.room-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: white;
  margin: 0 0 0.5rem 0;
  text-align: center;
}

.password-description {
  font-size: 0.9rem;
  color: rgba(255, 255, 255, 0.7);
  margin: 0 0 1.5rem 0;
  text-align: center;
}

.password-form {
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

.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input-wrapper .form-input {
  padding-right: 3rem;
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

.error-message {
  font-size: 0.85rem;
  color: #ef4444;
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
