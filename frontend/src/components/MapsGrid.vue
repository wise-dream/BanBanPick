<script setup lang="ts">
import MapCard from './MapCard.vue';
import type { MapName } from '../types/veto';

interface Props {
  allMaps: MapName[];
  pickedMap: MapName | null;
  finished: boolean;
  started?: boolean;
  canBan?: boolean;
  canPick?: boolean;
  actionType?: 'ban' | 'pick' | null;
  isMapBanned: (mapName: MapName) => boolean;
  isMapPicked: (mapName: MapName) => boolean;
}

const props = withDefaults(defineProps<Props>(), {
  canPick: false,
  actionType: null,
});

const emit = defineEmits<{
  ban: [mapName: MapName];
  pick: [mapName: MapName];
}>();

function handleBan(mapName: MapName) {
  emit('ban', mapName);
}

function handlePick(mapName: MapName) {
  emit('pick', mapName);
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
      :disabled="!started || finished || isMapPicked(map) || (actionType === 'ban' && (canBan === false || isMapBanned(map))) || (actionType === 'pick' && (canPick === false || isMapBanned(map)))"
      :can-pick="canPick"
      :action-type="actionType"
      @ban="handleBan"
      @pick="handlePick"
    />
  </div>
</template>
