<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from '../composables/useI18n';
import type { MapName } from '../types/veto';

const { t } = useI18n();

interface FinalMapData {
  order: number;
  mapName: MapName;
  attackTeam: 'A' | 'B' | null;
  defenceTeam: 'A' | 'B' | null;
}

interface Props {
  show: boolean;
  mapName: MapName | null; // Для BO1
  vetoType?: 'bo1' | 'bo3' | 'bo5';
  mapsData?: FinalMapData[]; // Для BO3/BO5
  teamAName?: string;
  teamBName?: string;
}

const props = withDefaults(defineProps<Props>(), {
  vetoType: 'bo1',
  mapsData: () => [],
  teamAName: 'Team A',
  teamBName: 'Team B'
});

const emit = defineEmits<{
  close: [];
}>();

// Для BO1 - фон одной карты
const bgImage = computed(() => {
  if (props.vetoType !== 'bo1' || !props.mapName) return '';
  const urlName = props.mapName.toLowerCase().replace(/\s+/g, '-');
  return `url("/images/${urlName}.png")`;
});

// Для BO3/BO5 - фон в зависимости от типа серии
const seriesBgImage = computed(() => {
  if (props.vetoType === 'bo3') {
    return `url("/images/background_bo3.jpg")`;
  }
  if (props.vetoType === 'bo5') {
    return `url("/images/background_bo5.jpg")`;
  }
  return '';
});

// Получить название команды
function getTeamName(team: 'A' | 'B' | null): string {
  if (team === 'A') return props.teamAName;
  if (team === 'B') return props.teamBName;
  return '—';
}

// Получить URL фона карты
function getMapBgImage(mapName: MapName): string {
  const urlName = mapName.toLowerCase().replace(/\s+/g, '-');
  return `url("/images/${urlName}.png")`;
}
</script>

<template>
  <div v-if="show" class="final-overlay" @click.self="emit('close')">
    <!-- BO1: одна карта -->
    <div v-if="vetoType === 'bo1' && mapName" class="final-card" :style="{ '--final-bg': bgImage }">
      <div class="final-map-name">{{ mapName }}</div>
      <div class="final-subtitle">{{ t('finalOverlay.thisMapWillBePlayed') }}</div>
      <button class="btn btn-accent" @click="emit('close')">{{ t('common.ok') }}</button>
    </div>
    
    <!-- BO3/BO5: несколько карт -->
    <div v-else-if="(vetoType === 'bo3' || vetoType === 'bo5') && mapsData.length > 0" class="final-card" :style="{ '--final-bg': seriesBgImage }">
      <div class="final-map-name">{{ t('finalOverlay.vetoResult') }}</div>
      <div class="final-subtitle"></div>
      
      <div class="final-cards-row">
        <div
          v-for="mapData in mapsData"
          :key="mapData.order"
          class="final-map-card-compact"
          :style="{ backgroundImage: getMapBgImage(mapData.mapName) }"
        >
          <div class="final-map-card-inner">
            <div class="final-map-card-title">{{ t('finalOverlay.map') }} {{ mapData.order }}</div>
            <div class="final-map-card-mapname">{{ mapData.mapName }}</div>
            <div class="final-map-card-sides">
              <span class="label">{{ t('finalOverlay.sides') }}</span><br>
              {{ t('veto.attack') }}: <span class="attack">{{ getTeamName(mapData.attackTeam) }}</span><br>
              {{ t('veto.defence') }}: <span class="defence">{{ getTeamName(mapData.defenceTeam) }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <button class="btn btn-accent" style="align-self: flex-end; margin-top: 16px;" @click="emit('close')">{{ t('common.close') }}</button>
    </div>
  </div>
</template>
