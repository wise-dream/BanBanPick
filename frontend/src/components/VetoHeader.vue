<script setup lang="ts">
import type { Team } from '../types/veto';

interface Props {
  currentTeam: Team;
  teamAName: string;
  teamBName: string;
  started: boolean;
  finished: boolean;
}

interface Emits {
  (e: 'start'): void;
  (e: 'swap'): void;
  (e: 'reset'): void;
  (e: 'side'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

</script>

<template>
  <header>
    <div>
      <h1>Map Banning</h1>
      <div class="tagline"></div>
    </div>

    <div class="top-bar">
      <div class="teams">
        <div class="team" :class="{ active: currentTeam === 'A' }">
          <span class="badge">A</span>
          {{ teamAName }}
        </div>
        <div class="team" :class="{ active: currentTeam === 'B' }">
          <span class="badge">B</span>
          {{ teamBName }}
        </div>
      </div>

      <div class="controls">
        <button 
          v-if="!started" 
          class="btn btn-accent" 
          @click="emit('start')"
        >
          Начать
        </button>
        <button 
          v-if="started && !finished" 
          class="btn" 
          @click="emit('swap')"
        >
          Сменить ход
        </button>
        <button 
          v-if="started" 
          class="btn" 
          @click="emit('reset')"
        >
          Сброс
        </button>
        <button 
          v-if="started && finished" 
          class="btn" 
          @click="emit('side')"
        >
          Выбор стороны
        </button>
      </div>
    </div>
  </header>
</template>
