<script setup lang="ts">
import MapCard from './MapCard.vue';
import type { MapName } from '../types/veto';

interface Props {
  allMaps: MapName[];
  pickedMap: MapName | null;
  finished: boolean;
  started?: boolean;
  canBan?: boolean;
  isMapBanned: (mapName: MapName) => boolean;
  isMapPicked: (mapName: MapName) => boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  ban: [mapName: MapName];
}>();

function handleBan(mapName: MapName) {
  emit('ban', mapName);
}
</script>

<template>
  <div class="maps-grid">
    <MapCard
      v-for="map in allMaps"
      :key="map"
      :map-name="map"
      :is-banned="isMapBanned(map)"
      :is-picked="isMapPicked(map)"
      :disabled="!started || finished || canBan === false"
      @ban="handleBan"
    />
  </div>
</template>
