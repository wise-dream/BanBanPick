<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { X } from 'lucide-vue-next';
import { getAllMaps, saveCustomPool } from '../../services/mapPoolService';
import type { Map } from '../../types';
import { useI18n } from '../../composables/useI18n';
import InputText from 'primevue/inputtext';

interface Props {
  isOpen: boolean;
}

interface Emits {
  (e: 'close'): void;
  (e: 'created', pool: any): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();
const { t } = useI18n();

const poolName = ref('');
const selectedMapIds = ref<number[]>([]);
const error = ref('');
const isLoading = ref(false);

const allMaps = ref<Map[]>([]);

// Загружаем карты при монтировании
onMounted(async () => {
  allMaps.value = await getAllMaps();
});

const selectedMaps = computed(() => {
  return allMaps.value.filter((map: Map) => selectedMapIds.value.includes(map.id));
});

const toggleMap = (mapId: number) => {
  const index = selectedMapIds.value.indexOf(mapId);
  if (index > -1) {
    selectedMapIds.value.splice(index, 1);
  } else {
    selectedMapIds.value.push(mapId);
  }
  error.value = '';
};

const handleCreate = async () => {
  error.value = '';
  isLoading.value = true;

  if (!poolName.value.trim()) {
    error.value = t('mapPool.nameRequired');
    isLoading.value = false;
    return;
  }

  if (selectedMapIds.value.length === 0) {
    error.value = t('mapPool.atLeastOneMap');
    isLoading.value = false;
    return;
  }

  try {
    const newPool = await saveCustomPool({
      gameId: 1,
      name: poolName.value.trim(),
      type: 'custom',
      maps: selectedMaps.value
    });

    emit('created', newPool);
    handleClose();
  } catch (err: any) {
    const apiError = err as { message?: string; code?: string };
    if (apiError.code === 'HTTP_409') {
      error.value = t('errors.poolNameExists');
    } else if (apiError.message) {
      error.value = apiError.message;
    } else {
      error.value = t('mapPool.saveError');
    }
  } finally {
    isLoading.value = false;
  }
};

const handleClose = () => {
  poolName.value = '';
  selectedMapIds.value = [];
  error.value = '';
  emit('close');
};

watch(() => props.isOpen, (isOpen) => {
  if (!isOpen) {
    handleClose();
  }
});
</script>

<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="handleClose">
    <div class="modal-content">
      <div class="modal-header">
        <h2>{{ t('mapPool.createCustomPool') }}</h2>
        <button class="close-button" @click="handleClose" aria-label="Close">
          <X :size="24" />
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="pool-name" class="form-label">
            {{ t('mapPool.poolName') }}
          </label>
          <InputText
            id="pool-name"
            v-model="poolName"
            :placeholder="t('mapPool.poolNamePlaceholder')"
            maxlength="100"
            class="w-full"
          />
        </div>

        <div class="form-group">
          <label class="form-label">
            {{ t('mapPool.selectMaps') }} ({{ selectedMapIds.length }})
          </label>
          <div class="maps-grid">
            <button
              v-for="map in allMaps"
              :key="map.id"
              class="map-option"
              :class="{ selected: selectedMapIds.includes(map.id) }"
              @click="toggleMap(map.id)"
            >
              <span class="map-name">{{ map.name }}</span>
              <span v-if="map.isCompetitive" class="competitive-badge">Competitive</span>
            </button>
          </div>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button class="btn btn-primary" @click="handleCreate" :disabled="selectedMapIds.length === 0 || isLoading">
          <span v-if="!isLoading">{{ t('mapPool.create') }}</span>
          <span v-else>{{ t('common.loading') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 2rem;
}

.modal-content {
  background: rgba(0, 0, 0, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: white;
  font-weight: 600;
}

.close-button {
  background: transparent;
  border: none;
  color: white;
  cursor: pointer;
  padding: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
}

.close-button:hover {
  background: rgba(255, 255, 255, 0.1);
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-label {
  display: block;
  margin-bottom: 0.5rem;
  color: white;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  color: white;
  font-size: 1rem;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
}

.maps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 0.75rem;
  max-height: 300px;
  overflow-y: auto;
  padding: 0.5rem;
}

.map-option {
  background: rgba(0, 0, 0, 0.9);
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  padding: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  text-align: left;
}

.map-option:hover {
  border-color: rgba(102, 126, 234, 0.5);
  background: rgba(102, 126, 234, 0.1);
}

.map-option.selected {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.2);
}

.map-name {
  color: white;
  font-weight: 500;
  font-size: 0.9rem;
}

.competitive-badge {
  font-size: 0.7rem;
  color: #4caf50;
  font-weight: 600;
}

.error-message {
  color: #ef4444;
  font-size: 0.85rem;
  margin-top: 0.5rem;
  padding: 0.5rem;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 4px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>
