<script setup lang="ts">
interface Props {
  size?: 'small' | 'medium' | 'large';
  fullscreen?: boolean;
  message?: string;
}

const props = withDefaults(defineProps<Props>(), {
  size: 'medium',
  fullscreen: false,
  message: '',
});
</script>

<template>
  <div :class="['loader-container', { 'loader-fullscreen': fullscreen }]">
    <div :class="['loader', `loader-${size}`]">
      <div class="loader-spinner">
        <div class="spinner-ring"></div>
        <div class="spinner-ring"></div>
        <div class="spinner-ring"></div>
        <div class="spinner-ring"></div>
      </div>
      <p v-if="message" class="loader-message">{{ message }}</p>
    </div>
  </div>
</template>

<style scoped>
.loader-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.loader-fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(5px);
  z-index: 9999;
  padding: 0;
}

.loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.loader-small .loader-spinner {
  width: 24px;
  height: 24px;
}

.loader-medium .loader-spinner {
  width: 48px;
  height: 48px;
}

.loader-large .loader-spinner {
  width: 64px;
  height: 64px;
}

.loader-spinner {
  position: relative;
  width: 48px;
  height: 48px;
}

.spinner-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border: 3px solid transparent;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;
}

.spinner-ring:nth-child(1) {
  animation-delay: -0.45s;
}

.spinner-ring:nth-child(2) {
  animation-delay: -0.3s;
  border-top-color: #764ba2;
}

.spinner-ring:nth-child(3) {
  animation-delay: -0.15s;
  border-top-color: #667eea;
}

.spinner-ring:nth-child(4) {
  animation-delay: 0s;
  border-top-color: #764ba2;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.loader-message {
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.9rem;
  margin: 0;
  text-align: center;
}

.loader-fullscreen .loader-message {
  color: white;
  font-size: 1rem;
}
</style>
