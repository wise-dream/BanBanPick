<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../../store/auth';
import { useI18n } from '../../composables/useI18n';
import Navigation from './Navigation.vue';
import LanguageSelector from '../common/LanguageSelector.vue';

const { t } = useI18n();

const router = useRouter();
const authStore = useAuthStore();
const isMobileMenuOpen = ref(false);

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value;
};

const navigateToHome = () => {
  router.push('/');
  isMobileMenuOpen.value = false;
};
</script>

<template>
  <header class="app-header">
    <div class="header-container">
      <div class="logo" @click="navigateToHome">
        <h1>{{ t('header.appName') }}</h1>
      </div>

      <Navigation
        :is-mobile-menu-open="isMobileMenuOpen"
        :is-authenticated="authStore.isAuthenticated"
        @close-mobile-menu="isMobileMenuOpen = false"
      />

      <div class="header-actions">
        <LanguageSelector />
        <div v-if="authStore.isAuthenticated" class="user-menu">
          <button class="btn btn-secondary" @click="router.push('/profile')">
            {{ authStore.user?.username }}
          </button>
          <button class="btn btn-secondary" @click="authStore.logout(); router.push('/')">
            {{ t('common.logout') }}
          </button>
        </div>
        <div v-else class="auth-buttons">
          <button class="btn btn-primary" @click="router.push('/login')">
            {{ t('common.login') }}
          </button>
          <button class="btn btn-secondary" @click="router.push('/register')">
            {{ t('common.register') }}
          </button>
        </div>
        <button
          class="mobile-menu-toggle"
          @click="toggleMobileMenu"
          :aria-label="t('header.toggleMenu')"
        >
          <span></span>
          <span></span>
          <span></span>
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(10px);
  padding: 1rem calc((100vw - 1440px) / 2);
  position: sticky;
  top: 0;
  z-index: 1000;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header-container {
  width: 100%;
  max-width: 1440px;  
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}

.logo {
  cursor: pointer;
  user-select: none;
}

.logo h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: bold;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.auth-buttons,
.user-menu {
  display: flex;
  gap: 0.5rem;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
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
  height: 36px;
  background: rgba(0, 0, 0, 0.9);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
}

.mobile-menu-toggle {
  display: none;
  flex-direction: column;
  gap: 4px;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
}

.mobile-menu-toggle span {
  width: 25px;
  height: 2px;
  background: white;
  transition: all 0.3s;
}

@media (max-width: 1440px) {
  .app-header {
    padding: 1rem 2rem;
  }
}

@media (max-width: 768px) {
  .app-header {
    padding: 1rem 1rem;
  }

  .mobile-menu-toggle {
    display: flex;
  }

  .auth-buttons,
  .user-menu {
    display: none;
  }
}
</style>
