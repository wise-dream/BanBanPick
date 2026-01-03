import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import { useAuthStore } from '../store/auth';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../pages/HomePage.vue')
  },
  {
    path: '/ban/valorant',
    name: 'MapPoolSelection',
    component: () => import('../pages/MapPoolSelectionPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/create-room/best-of',
    name: 'BestOfSelection',
    component: () => import('../pages/BestOfSelectionPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/create-room/final',
    name: 'CreateRoomFinal',
    component: () => import('../pages/CreateRoomFinalPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/veto/valorant/:poolId',
    name: 'VetoProcess',
    component: () => import('../pages/VetoProcessPage.vue')
  },
  {
    path: '/rooms',
    name: 'RoomsList',
    component: () => import('../pages/RoomsListPage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/room/:roomId',
    name: 'Room',
    component: () => import('../pages/RoomPage.vue'),
    props: true,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('../pages/ProfilePage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../pages/LoginPage.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../pages/RegisterPage.vue')
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../pages/NotFoundPage.vue')
  }
];

export const router = createRouter({
  history: createWebHistory(),
  routes
});

// Навигационный guard для проверки авторизации
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore();
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({
      path: '/login',
      query: { redirect: to.fullPath }
    });
  } else {
    next();
  }
});

// Глобальный обработчик ошибок навигации
router.onError((error) => {
  console.error('Router navigation error:', error);
  // Можно добавить логику для обработки ошибок навигации
  // Например, редирект на страницу ошибки или показ уведомления
});
