<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { X, AlertCircle } from 'lucide-vue-next';
import { useI18n } from '../../composables/useI18n';

const { t } = useI18n();

interface Props {
  message: string;
  type?: 'error' | 'warning' | 'info' | 'success';
  duration?: number;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'error',
  duration: 5000,
});

const emit = defineEmits<{
  close: [];
}>();

const isVisible = ref(true);
let timeoutId: number | null = null;

onMounted(() => {
  if (props.duration > 0) {
    timeoutId = window.setTimeout(() => {
      close();
    }, props.duration);
  }
});

onUnmounted(() => {
  if (timeoutId) {
    clearTimeout(timeoutId);
  }
});

const close = () => {
  isVisible.value = false;
  setTimeout(() => {
    emit('close');
  }, 300); // Wait for animation
};
</script>

<template>
  <Transition name="toast">
    <div v-if="isVisible" class="error-toast" :class="`toast-${type}`">
      <div class="toast-content">
        <AlertCircle :size="20" class="toast-icon" />
        <span class="toast-message">{{ message }}</span>
        <button class="toast-close" @click="close" :aria-label="t('errorToast.close')">
          <X :size="16" />
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.error-toast {
  position: fixed;
  top: 2rem;
  right: 2rem;
  z-index: 10000;
  min-width: 300px;
  max-width: 500px;
  padding: 1rem 1.5rem;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
}

.toast-error {
  background: rgba(239, 68, 68, 0.9);
  border: 1px solid rgba(239, 68, 68, 1);
  color: white;
}

.toast-warning {
  background: rgba(255, 193, 7, 0.9);
  border: 1px solid rgba(255, 193, 7, 1);
  color: #000;
}

.toast-info {
  background: rgba(59, 130, 246, 0.9);
  border: 1px solid rgba(59, 130, 246, 1);
  color: white;
}

.toast-success {
  background: rgba(34, 197, 94, 0.9);
  border: 1px solid rgba(34, 197, 94, 1);
  color: white;
}

.toast-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.toast-icon {
  flex-shrink: 0;
}

.toast-message {
  flex: 1;
  font-size: 0.95rem;
  font-weight: 500;
}

.toast-close {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
  flex-shrink: 0;
}

.toast-close:hover {
  background: rgba(0, 0, 0, 0.1);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

@media (max-width: 768px) {
  .error-toast {
    top: 1rem;
    right: 1rem;
    left: 1rem;
    min-width: auto;
    max-width: none;
  }
}
</style>
