<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useAuthStore } from '../store/auth';

const router = useRouter();
const authStore = useAuthStore();

const goToMapPoolSelection = () => {
  // Проверяем авторизацию перед переходом
  if (!authStore.isAuthenticated) {
    // Сохраняем путь для редиректа после логина
    router.push({ path: '/login', query: { redirect: '/ban/valorant' } });
    return;
  }
  
  router.push('/ban/valorant');
};
</script>

<template>
  <div class="home-page">
    <div class="hero-section">
      <h1 class="hero-title">Valorant Map Ban Tool</h1>
      <p class="hero-description">
        Create and manage map veto sessions for your Valorant matches
      </p>
      <button class="btn-primary" @click="goToMapPoolSelection">
        Start Veto Process
      </button>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  width: 100%;
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
}

.hero-section {
  text-align: center;
  color: white;
  max-width: 800px;
  padding: 2rem;
}

.hero-title {
  font-size: 3.5rem;
  margin-bottom: 1rem;
  font-weight: 700;
}

.hero-description {
  font-size: 1.25rem;
  margin-bottom: 2rem;
  opacity: 0.9;
}

.btn-primary {
  background: white;
  color: #667eea;
  border: none;
  padding: 1rem 2.5rem;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

@media (max-width: 768px) {
  .hero-title {
    font-size: 2.5rem;
  }

  .hero-description {
    font-size: 1rem;
  }
}
</style>
