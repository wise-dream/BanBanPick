<script setup lang="ts">
import { RouterView } from 'vue-router';
import Header from './components/layout/Header.vue';
import Footer from './components/layout/Footer.vue';
import BackgroundLayer from './components/BackgroundLayer.vue';
import ErrorToast from './components/common/ErrorToast.vue';
import Loader from './components/common/Loader.vue';
import { useErrorToast } from './composables/useErrorToast';
import { useOffline } from './composables/useOffline';

const { toasts, removeToast } = useErrorToast();
const { isOffline, wasOffline } = useOffline();
</script>

<template>
  <BackgroundLayer />
  <div class="app-wrapper">
    <Header />
    <main class="main-content">
      <RouterView />
    </main>
    <Footer />
  </div>
</template>

<style>
#app {
  width: 100%;
  min-height: 100vh;
  position: relative;
}

.app-wrapper {
  position: relative;
  z-index: 1;
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.toast-container {
  position: fixed;
  top: 0;
  right: 0;
  z-index: 10000;
  pointer-events: none;
}

.offline-indicator,
.online-indicator {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10001;
  padding: 1rem;
  text-align: center;
  font-weight: 500;
  backdrop-filter: blur(10px);
}

.offline-indicator {
  background: rgba(239, 68, 68, 0.9);
  color: white;
}

.online-indicator {
  background: rgba(34, 197, 94, 0.9);
  color: white;
}

.offline-enter-active,
.offline-leave-active {
  transition: all 0.3s ease;
}

.offline-enter-from,
.offline-leave-to {
  opacity: 0;
  transform: translateY(-100%);
}
</style>
