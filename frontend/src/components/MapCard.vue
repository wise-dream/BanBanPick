<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from '../composables/useI18n';
import type { MapName } from '../types/veto';

const { t } = useI18n();

interface Props {
  mapName: MapName;
  isBanned: boolean;
  isPicked: boolean;
  disabled: boolean;
  canPick?: boolean;
  actionType?: 'ban' | 'pick' | null;
}

const props = withDefaults(defineProps<Props>(), {
  canPick: false,
  actionType: null,
});

const emit = defineEmits<{
  ban: [mapName: MapName];
  pick: [mapName: MapName];
}>();

function handleBanClick() {
  // Не эмитим событие, если кнопка disabled
  if (props.disabled) {
    return
  }
  emit('ban', props.mapName)
}

function handlePickClick() {
  // Не эмитим событие, если кнопка disabled
  if (props.disabled || !props.canPick) {
    return
  }
  emit('pick', props.mapName)
}

const mapClass = computed(() => {
  return `map-${props.mapName.toLowerCase().replace(/\s+/g, '-')}`;
});

const statusText = computed(() => {
  if (props.isPicked) return t('veto.selected');
  if (props.isBanned) return t('veto.banned');
  if (props.actionType === 'pick') return t('veto.availableForPick');
  return t('veto.available');
});

const cardClasses = computed(() => {
  return [
    'map-card',
    mapClass.value,
    {
      // Карта disabled визуально только если забанена или пикнута
      // Но кнопки все равно показываем (disabled)
      'map-disabled': props.isBanned || props.isPicked,
      'map-banned': props.isBanned,
      'map-picked': props.isPicked
    }
  ];
});

// Определяем, должна ли кнопка Ban быть видна
// Кнопка бана всегда видна при actionType === 'ban', даже если карта забанена или пикнута
const shouldShowBanButton = computed(() => {
  return props.actionType === 'ban'
});

// Определяем, должна ли кнопка Pick быть видна
const shouldShowPickButton = computed(() => {
  return props.actionType === 'pick' && props.canPick && !props.isBanned && !props.isPicked
});

// Определяем, должна ли кнопка быть disabled
// Кнопка бана disabled если: карта забанена, пикнута, или общий disabled (не очередь/не начато/завершено)
const isButtonDisabled = computed(() => {
  if (props.actionType === 'ban') {
    return props.disabled || props.isBanned || props.isPicked
  }
  return props.disabled
});
</script>

<template>
  <div :class="cardClasses">
    <div class="map-name">{{ mapName }}</div>
    <div class="map-status">{{ statusText }}</div>
    <div class="map-actions">
      <button
        v-if="shouldShowBanButton"
        class="btn btn-accent"
        :class="{ 'btn-disabled': isButtonDisabled }"
        :disabled="isButtonDisabled"
        @click="handleBanClick"
      >
        {{ t('veto.ban') }}
      </button>
      <button
        v-if="shouldShowPickButton"
        class="btn btn-pick"
        :class="{ 'btn-disabled': isButtonDisabled }"
        :disabled="isButtonDisabled"
        @click="handlePickClick"
      >
        Pick
      </button>
    </div>
    <div v-if="isBanned || isPicked" :class="['map-tag', isPicked ? 'picked' : 'banned']">
      {{ isPicked ? 'PICKED' : 'BANNED' }}
    </div>
  </div>
</template>
