import { ref } from 'vue';
import type { ApiError } from '../services/api/types';
import { getErrorMessage } from '../services/api/errorHandler';

export interface ToastMessage {
  id: number;
  message: string;
  type: 'error' | 'warning' | 'info' | 'success';
}

const toasts = ref<ToastMessage[]>([]);
let toastIdCounter = 0;

export function useErrorToast() {
  const showError = (error: ApiError | Error | unknown, type: ToastMessage['type'] = 'error') => {
    const message = getErrorMessage(error);
    const id = toastIdCounter++;
    
    toasts.value.push({
      id,
      message,
      type,
    });

    // Auto remove after 5 seconds
    setTimeout(() => {
      removeToast(id);
    }, 5000);
  };

  const showSuccess = (message: string) => {
    showError({ code: '', message } as ApiError, 'success');
  };

  const showWarning = (message: string) => {
    showError({ code: '', message } as ApiError, 'warning');
  };

  const showInfo = (message: string) => {
    showError({ code: '', message } as ApiError, 'info');
  };

  const removeToast = (id: number) => {
    const index = toasts.value.findIndex(t => t.id === id);
    if (index > -1) {
      toasts.value.splice(index, 1);
    }
  };

  return {
    toasts,
    showError,
    showSuccess,
    showWarning,
    showInfo,
    removeToast,
  };
}
