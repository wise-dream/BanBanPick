<script setup lang="ts">
import { computed } from 'vue';
import type { MapName } from '../types/veto';

interface Props {
  mapName: MapName;
  isBanned: boolean;
  isPicked: boolean;
  disabled: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  ban: [mapName: MapName];
}>();

function handleBanClick() {
  // Не эмитим событие, если кнопка disabled
  if (props.disabled) {
    return
  }
  emit('ban', props.mapName)
}

const mapClass = computed(() => {
  return `map-${props.mapName.toLowerCase().replace(/\s+/g, '-')}`;
});

const statusText = computed(() => {
  if (props.isPicked) return 'Выбрана для игры';
  if (props.isBanned) return 'Забанена';
  return 'Доступна для бана';
});

const cardClasses = computed(() => {
  return [
    'map-card',
    mapClass.value,
    {
      'map-disabled': props.isBanned || props.isPicked,
      'map-banned': props.isBanned,
      'map-picked': props.isPicked
    }
  ];
});

// Определяем, должна ли кнопка быть видна
const shouldShowButton = computed(() => {
  return !props.isBanned && !props.isPicked
});

// Определяем, должна ли кнопка быть disabled
const isButtonDisabled = computed(() => {
  return props.disabled
});
</script>

<template>
  <div :class="cardClasses">
    <div class="map-name">{{ mapName }}</div>
    <div class="map-status">{{ statusText }}</div>
    <div class="map-actions">
      <button
        v-if="shouldShowButton"
        class="btn btn-accent"
        :class="{ 'btn-disabled': isButtonDisabled }"
        :disabled="isButtonDisabled"
        @click="handleBanClick"
      >
        Ban
      </button>
    </div>
    <div v-if="isBanned || isPicked" :class="['map-tag', isPicked ? 'picked' : 'banned']">
      {{ isPicked ? 'PICKED' : 'BANNED' }}
    </div>
  </div>
</template>
