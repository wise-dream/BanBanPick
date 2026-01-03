<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from '../composables/useI18n';
import { useAuthStore } from '../store/auth';
import { useRoomCreationStore } from '../store/roomCreation';
import StepIndicator from '../components/common/StepIndicator.vue';
import MapPoolCard from '../components/map-pool/MapPoolCard.vue';
import CreateCustomPoolModal from '../components/map-pool/CreateCustomPoolModal.vue';
import { getAllPools } from '../services/mapPoolService';
import type { MapPool } from '../types';

const router = useRouter();
const { t } = useI18n();
const authStore = useAuthStore();
const roomCreationStore = useRoomCreationStore();

const pools = ref<MapPool[]>([]);
const selectedPool = ref<MapPool | null>(null);
const showCreateModal = ref(false);
const isLoading = ref(true);
const error = ref<string | null>(null);

// Проверка авторизации при монтировании
onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: '/ban/valorant' } });
    return;
  }
  loadPools();
});

// loadPools теперь вызывается в onMounted после проверки авторизации

const loadPools = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    const apiPools = await getAllPools();
    
    // Добавляем карточку для создания кастомного пула
    const customPoolCard: MapPool = {
      id: -1, // Специальный ID для карточки создания
      gameId: 1,
      name: t('mapPool.customMapPool'),
      type: 'custom',
      isSystem: false,
      maps: []
    };
    
    pools.value = [...apiPools, customPoolCard];
  } catch (err) {
    error.value = 'Не удалось загрузить пулы карт';
    console.error('Error loading pools:', err);
  } finally {
    isLoading.value = false;
  }
};

const handleSelectPool = async (pool: MapPool) => {
  // Специальная карточка для создания кастомного пула
  if (pool.id === -1) {
    showCreateModal.value = true;
    return;
  }

  // Проверяем авторизацию перед выбором пула
  if (!authStore.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: '/ban/valorant' } });
    return;
  }

  try {
    selectedPool.value = pool;
    // Сохраняем выбранный пул в store и переходим к выбору Best of
    roomCreationStore.setPool(pool);
    await router.push('/create-room/best-of');
  } catch (error) {
    console.error('Error navigating to best-of selection:', error);
    // В случае ошибки остаемся на текущей странице
  }
};

const handlePoolCreated = (newPool: MapPool) => {
  loadPools(); // Перезагрузить список пулов
  selectedPool.value = newPool;
  // Сохраняем выбранный пул в store и переходим к выбору Best of
  roomCreationStore.setPool(newPool);
  router.push('/create-room/best-of');
};
</script>

<template>
  <div class="map-pool-selection-page">
    <div class="page-container">
      <h1 class="page-title">{{ t('mapPool.title') }}</h1>
      
      <StepIndicator :current-step="1" />

      <p class="page-description">
        {{ t('mapPool.selectPool') }}
      </p>

      <div class="pools-grid">
        <MapPoolCard
          v-for="pool in pools"
          :key="pool.id"
          :pool="pool"
          :is-selected="selectedPool?.id === pool.id"
          @select="handleSelectPool"
        />
      </div>

      <p class="info-text">
        {{ t('mapPool.description') }}
      </p>
    </div>

    <CreateCustomPoolModal
      :is-open="showCreateModal"
      @close="showCreateModal = false"
      @created="handlePoolCreated"
    />
  </div>
</template>

<style scoped>
.map-pool-selection-page {
  width: 100%;
  min-height: calc(100vh - 200px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
  padding: 2rem;
}

.page-container {
  width: 100%;
  max-width: 1440px;
}

.page-title {
  text-align: center;
  color: white;
  font-size: 2.5rem;
  margin-bottom: 2rem;
  font-weight: 700;
}

.page-description {
  text-align: center;
  color: rgba(255, 255, 255, 0.9);
  font-size: 1.1rem;
  margin-bottom: 2rem;
}

.pools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin-bottom: 2rem;
}

.info-text {
  text-align: center;
  color: rgb(255, 255, 255);
  font-size: 0.95rem;
  margin-top: 2rem;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.9);
  border-radius: 8px;
}

@media (max-width: 768px) {
  .pools-grid {
    grid-template-columns: 1fr;
  }

  .page-title {
    font-size: 2rem;
  }
}
</style>
