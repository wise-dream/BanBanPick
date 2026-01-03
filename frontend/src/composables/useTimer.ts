import { ref, onUnmounted } from 'vue';

const BAN_TIME_LIMIT = 20;

export function useTimer() {
  const timeLeft = ref(BAN_TIME_LIMIT);
  let timerId: ReturnType<typeof setInterval> | null = null;

  function startBanTimer(onTimeout?: () => void) {
    clearInterval(timerId!);
    timeLeft.value = BAN_TIME_LIMIT;

    timerId = setInterval(() => {
      timeLeft.value -= 1;
      if (timeLeft.value <= 0) {
        clearInterval(timerId!);
        timeLeft.value = BAN_TIME_LIMIT;
        onTimeout?.();
      }
    }, 1000);
  }

  function stopBanTimer() {
    if (timerId) {
      clearInterval(timerId);
      timerId = null;
    }
  }

  function resetTimer() {
    stopBanTimer();
    timeLeft.value = BAN_TIME_LIMIT;
  }

  onUnmounted(() => {
    stopBanTimer();
  });

  return {
    timeLeft,
    startBanTimer,
    stopBanTimer,
    resetTimer,
    BAN_TIME_LIMIT
  };
}
