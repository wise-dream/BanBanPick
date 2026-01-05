<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from '../composables/useI18n';
import { useAuthStore } from '../store/auth';
import { useRoomCreationStore, type BestOfType } from '../store/roomCreation';
import StepIndicator from '../components/common/StepIndicator.vue';
import { Star } from 'lucide-vue-next';

const router = useRouter();
const { t } = useI18n();
const authStore = useAuthStore();
const roomCreationStore = useRoomCreationStore();

const selectedBestOf = ref<BestOfType | null>(roomCreationStore.bestOf);

// Проверка авторизации и наличия выбранного пула
onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: '/ban/valorant' } });
    return;
  }
  
  if (!roomCreationStore.selectedPool) {
    // Если пул не выбран, возвращаемся к выбору пула
    router.push('/ban/valorant');
    return;
  }
});

const bestOfOptions: Array<{ value: BestOfType; label: string; stars: number }> = [
  { value: 1, label: 'Best of 1', stars: 1 },
  { value: 3, label: 'Best of 3', stars: 3 },
  { value: 5, label: 'Best of 5', stars: 0 } // 0 означает сплошной круг
];

const handleSelectBestOf = (value: BestOfType) => {
  selectedBestOf.value = value;
  roomCreationStore.setBestOf(value);
  
  // Переходим к финальному шагу создания комнаты
  router.push('/create-room/final');
};

const handleBack = () => {
  router.push('/ban/valorant');
};
</script>

<template>
  <div class="best-of-selection-page">
    <div class="page-container">
      <h1 class="page-title">{{ t('bestOf.title') }}</h1>
      
      <StepIndicator :current-step="2" :total-steps="3" />

      <p class="page-description">
        {{ t('bestOf.description') }}
      </p>

      <div class="best-of-grid">
        <button
          v-for="option in bestOfOptions"
          :key="option.value"
          class="best-of-card"
          :class="{ selected: selectedBestOf === option.value }"
          @click="handleSelectBestOf(option.value)"
        >
          <div class="icon-container">
            <div v-if="option.stars > 0" class="stars-container">
              <Star
                v-for="i in option.stars"
                :key="i"
                :size="32"
                class="star-icon"
                :fill="selectedBestOf === option.value ? 'currentColor' : 'none'"
              />
            </div>
            <div v-else class="circle-icon"></div>
          </div>
          <div class="best-of-number">{{ option.value }}</div>
          <div class="best-of-label">{{ option.label }}</div>
        </button>
      </div>

      <p class="info-text">
        {{ t('bestOf.hint') }}
      </p>

      <div class="actions">
        <button class="btn btn-secondary" @click="handleBack">
          {{ t('common.back') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.best-of-selection-page {
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
  max-width: 1200px;
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
  margin-bottom: 3rem;
}

.best-of-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 2rem;
  margin-bottom: 2rem;
  max-width: 900px;
  margin-left: auto;
  margin-right: auto;
}

.best-of-card {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 2rem 1.5rem;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.best-of-card:hover {
  border-color: rgba(102, 126, 234, 0.5);
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.2);
}

.best-of-card.selected {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.1);
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.3);
}

.icon-container {
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stars-container {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  justify-content: center;
}

.star-icon {
  color: #667eea;
  transition: all 0.3s;
}

.best-of-card.selected .star-icon {
  color: white;
}

.circle-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: #667eea;
  transition: all 0.3s;
}

.best-of-card.selected .circle-icon {
  background: white;
  transform: scale(1.1);
}

.best-of-number {
  font-size: 3rem;
  font-weight: 700;
  color: white;
  line-height: 1;
}

.best-of-label {
  font-size: 1rem;
  color: rgba(255, 255, 255, 0.8);
  text-align: center;
}

.best-of-card.selected .best-of-label {
  color: white;
  font-weight: 600;
}

.info-text {
  text-align: center;
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.95rem;
  margin: 2rem 0;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.9);
  border-radius: 8px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.actions {
  display: flex;
  justify-content: center;
  margin-top: 2rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
}

@media (max-width: 768px) {
  .best-of-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .best-of-number {
    font-size: 2.5rem;
  }
}
</style>
