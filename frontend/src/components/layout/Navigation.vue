<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router';

interface Props {
  isMobileMenuOpen: boolean;
  isAuthenticated: boolean;
}

interface Emits {
  (e: 'close-mobile-menu'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();
const router = useRouter();
const route = useRoute();

const navItems = [
  { name: 'Home', path: '/', key: 'home' },
  { name: 'Rooms', path: '/rooms', key: 'rooms' },
  { name: 'Profile', path: '/profile', key: 'profile', requiresAuth: true }
];

const navigate = (path: string) => {
  router.push(path);
  emit('close-mobile-menu');
};

const isActive = (path: string) => {
  return route.path === path;
};
</script>

<template>
  <nav class="navigation" :class="{ 'mobile-open': isMobileMenuOpen }">
    <ul class="nav-list">
      <li
        v-for="item in navItems"
        :key="item.key"
        v-show="!item.requiresAuth || isAuthenticated"
      >
        <a
          :href="item.path"
          @click.prevent="navigate(item.path)"
          :class="{ active: isActive(item.path) }"
          class="nav-link"
        >
          {{ item.name }}
        </a>
      </li>
    </ul>
  </nav>
</template>

<style scoped>
.navigation {
  flex: 1;
  display: flex;
  justify-content: center;
}

.nav-list {
  display: flex;
  list-style: none;
  margin: 0;
  padding: 0;
  gap: 2rem;
}

.nav-link {
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  font-size: 0.95rem;
  transition: color 0.2s;
  position: relative;
  padding: 0.5rem 0;
}

.nav-link:hover {
  color: white;
}

.nav-link.active {
  color: white;
}

.nav-link.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

@media (max-width: 768px) {
  .navigation {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: rgba(0, 0, 0, 0.95);
    backdrop-filter: blur(10px);
    max-height: 0;
    overflow: hidden;
    transition: max-height 0.3s ease;
  }

  .navigation.mobile-open {
    max-height: 500px;
  }

  .nav-list {
    flex-direction: column;
    padding: 1rem 2rem;
    gap: 0;
  }

  .nav-link {
    display: block;
    padding: 1rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
}
</style>
