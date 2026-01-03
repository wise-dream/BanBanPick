<script setup lang="ts">
import { computed } from 'vue';
import type { LogEntry, MapName } from '../types/veto';

interface Props {
  pickedMap: MapName | null;
  logEntries: LogEntry[];
}

const props = defineProps<Props>();

const summaryText = computed(() => {
  if (!props.pickedMap) {
    return 'Пока карта не выбрана. Баньте, пока не останется одна.';
  }
  return `Играется карта: ${props.pickedMap}. Сторону выбираете при заходе в лобби.`;
});
</script>

<template>
  <section class="panel">
    <div class="panel-header">
      <div class="panel-title">Лог и итог</div>
    </div>

    <div class="summary">
      <div class="summary-title">Выбранная карта</div>
      <div>{{ summaryText }}</div>
    </div>

    <div class="panel-header" style="margin-top:8px;">
      <div class="panel-title" style="font-size:13px;">Ход вето</div>
    </div>
    <div class="log">
      <div
        v-for="(entry, index) in logEntries"
        :key="index"
        class="log-entry"
        v-html="entry.message"
      ></div>
    </div>
  </section>
</template>

<style scoped>
.log {
  scroll-behavior: smooth;
}
</style>
