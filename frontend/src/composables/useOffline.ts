import { ref, computed, onMounted, onUnmounted } from 'vue';

export function useOffline() {
  const isOffline = ref(!navigator.onLine);
  const wasOffline = ref(false);

  const updateOnlineStatus = () => {
    const wasOfflineBefore = isOffline.value;
    isOffline.value = !navigator.onLine;
    
    if (wasOfflineBefore && !isOffline.value) {
      wasOffline.value = true;
      // Сбрасываем флаг через некоторое время
      setTimeout(() => {
        wasOffline.value = false;
      }, 3000);
    }
  };

  onMounted(() => {
    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);
    
    // Проверяем статус при монтировании
    updateOnlineStatus();
  });

  onUnmounted(() => {
    window.removeEventListener('online', updateOnlineStatus);
    window.removeEventListener('offline', updateOnlineStatus);
  });

  return {
    isOffline,
    wasOffline,
    isOnline: computed(() => !isOffline.value),
  };
}
