<script setup lang="ts">
import { computed } from 'vue';
import type { MapPool } from '../../types';

interface Props {
  pool: MapPool;
  isSelected?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  isSelected: false
});

const emit = defineEmits<{
  select: [pool: MapPool];
}>();

const mapsList = computed(() => {
  return props.pool.maps.map(m => m.name).join(', ');
});

const handleClick = () => {
  emit('select', props.pool);
};
</script>

<template>
  <div
    class="map-pool-card"
    :class="{ selected: isSelected }"
    @click="handleClick"
  >
    <div class="card-header">
      <h3>{{ pool.name }}</h3>
      <span v-if="pool.isSystem" class="system-badge">System</span>
    </div>
    <div class="card-content">
      <p class="maps-list">
        <strong>Maps:</strong> {{ mapsList }}
      </p>
      <p v-if="pool.type === 'custom'" class="custom-note">
        Create your own map pool
      </p>
    </div>
    <div class="card-footer">
      <button class="btn-select">Select</button>
    </div>
  </div>
</template>

<style scoped>
.map-pool-card {
  background: rgba(255, 255, 255, 0.05);
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
  cursor: pointer;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
}

.map-pool-card:hover {
  border-color: rgba(102, 126, 234, 0.5);
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.2);
}

.map-pool-card.selected {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.1);
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.3);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: white;
  font-weight: 600;
}

.system-badge {
  background: rgba(76, 175, 80, 0.2);
  color: #4caf50;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.card-content {
  margin-bottom: 1rem;
}

.maps-list {
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.9rem;
  line-height: 1.6;
  margin: 0 0 0.5rem 0;
}

.maps-list strong {
  color: white;
}

.custom-note {
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.85rem;
  font-style: italic;
  margin: 0.5rem 0 0 0;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
}

.btn-select {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 0.5rem 1.5rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-select:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}
</style>
