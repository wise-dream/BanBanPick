<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useI18n } from '../composables/useI18n';

const { t } = useI18n();

interface Props {
  show: boolean;
  teamAName: string;
  teamBName: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
}>();

const showFullscreen = ref(false);
const showCard = ref(false);
const sideResult = ref<'ATTACK' | 'DEFENCE'>('ATTACK');
const sideSubtitle = ref('');

const sideIcon = computed(() => {
  return sideResult.value === 'ATTACK' ? '/images/attack.png' : '/images/defence.png';
});

const cardBg = computed(() => {
  return `url("${sideIcon.value}")`;
});

watch(() => props.show, (newVal) => {
  if (newVal) {
    showSideRandom();
  } else {
    showFullscreen.value = false;
    showCard.value = false;
  }
});

function showSideRandom() {
  const sides: ('ATTACK' | 'DEFENCE')[] = ['ATTACK', 'DEFENCE'];
  
  sideResult.value = 'ATTACK';
  sideSubtitle.value = t('veto.selectSide') + '...';
  
  showFullscreen.value = true;
  showCard.value = false;

  let ticks = 0;
  const maxTicks = 30;

  function spinStep() {
    const randomSide = sides[Math.floor(Math.random() * sides.length)] as 'ATTACK' | 'DEFENCE';
    sideResult.value = randomSide;
    ticks++;

    if (ticks < maxTicks) {
      const progress = ticks / maxTicks;
      const minDelay = 40;
      const maxDelay = 250;
      const delay = minDelay + (maxDelay - minDelay) * progress * progress;

      setTimeout(spinStep, delay);
    } else {
      const finalSide = sides[Math.floor(Math.random() * sides.length)] as 'ATTACK' | 'DEFENCE';
      const otherSide = finalSide === 'ATTACK' ? 'DEFENCE' : 'ATTACK';

      showFullscreen.value = false;

      sideResult.value = finalSide;
      sideSubtitle.value = `${props.teamAName} — ${finalSide}, ${props.teamBName} — ${otherSide}`;

      showCard.value = true;
    }
  }

  spinStep();
}
</script>

<template>
  <div v-if="show">
    <!-- Fullscreen spinner -->
    <div v-if="showFullscreen" class="side-overlay-fullscreen">
      <div class="side-full-content">
        <img :src="sideIcon" class="side-full-icon" alt="" />
        <div class="final-map-name">{{ sideResult }}</div>
      </div>
    </div>

    <!-- Result card -->
    <div v-if="showCard" class="final-overlay" @click.self="emit('close')">
      <div class="final-card side-final-card" :style="{ '--final-bg': cardBg }">
        <div class="final-map-name">{{ sideResult }}</div>
        <div class="final-subtitle">{{ sideSubtitle }}</div>
        <button class="btn btn-accent" @click="emit('close')">{{ t('common.ok') }}</button>
      </div>
    </div>
  </div>
</template>
