<script setup lang="ts">
import { computed } from 'vue';
import type { MapName } from '../types/veto';

interface Props {
  show: boolean;
  mapName: MapName | null;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
}>();

const bgImage = computed(() => {
  if (!props.mapName) return '';
  const urlName = props.mapName.toLowerCase().replace(/\s+/g, '-');
  return `url("/images/${urlName}.png")`;
});
</script>

<template>
  <div v-if="show" class="final-overlay" @click.self="emit('close')">
    <div class="final-card" :style="{ '--final-bg': bgImage }">
      <div class="final-map-name">{{ mapName }}</div>
      <div class="final-subtitle">Эта карта будет сыграна</div>
      <button class="btn btn-accent" @click="emit('close')">OK</button>
    </div>
  </div>
</template>
