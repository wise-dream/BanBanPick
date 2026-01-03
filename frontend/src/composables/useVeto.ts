import { ref, computed } from 'vue';
import type { MapName, VetoState, LogEntry } from '../types/veto';

export function useVeto() {
  const availableMaps = ref<MapName[]>([]);

  const state = ref<VetoState>({
    currentTeam: "A",
    bans: [],
    pickedMap: null,
    finished: false,
    started: false
  });

  const teamAName = ref("Team A");
  const teamBName = ref("Team B");
  const logEntries = ref<LogEntry[]>([]);

  const remainingMaps = computed(() => {
    return availableMaps.value.filter(map => !state.value.bans.includes(map));
  });

  function initializeMaps(maps: MapName[]) {
    availableMaps.value = maps;
    resetState(); // Сбросить состояние при инициализации новых карт
  }

  const currentTeamName = computed(() => {
    return state.value.currentTeam === "A" ? teamAName.value : teamBName.value;
  });

  function log(message: string) {
    logEntries.value.push({
      message,
      timestamp: Date.now()
    });
  }

  function startVeto() {
    if (state.value.started || state.value.finished) return;
    state.value.started = true;
    log("Новый Bo1 с банами до последней карты начат.");
  }

  function onBan(mapName: MapName) {
    if (!state.value.started || state.value.finished) return;
    if (state.value.bans.includes(mapName)) return;

    const teamName = currentTeamName.value;
    state.value.bans.push(mapName);
    log(`<strong>${teamName}</strong> банит карту <strong>${mapName}</strong>.`);

    const remaining = remainingMaps.value;

    if (remaining.length === 1) {
      const lastMap = remaining[0];
      if (lastMap) {
        state.value.pickedMap = lastMap;
        state.value.finished = true;
        log(
          `<strong>Автопик:</strong> последняя оставшаяся карта <strong>${lastMap}</strong> выбирается для игры.`
        );
        return lastMap;
      }
    }
    state.value.currentTeam = state.value.currentTeam === "A" ? "B" : "A";
    return null;
  }

  function swapCurrentTeam() {
    if (!state.value.started || state.value.finished) return;
    state.value.currentTeam = state.value.currentTeam === "A" ? "B" : "A";
    log(`Ход передан команде <strong>${currentTeamName.value}</strong>.`);
  }

  function resetState() {
    state.value = {
      currentTeam: "A",
      bans: [],
      pickedMap: null,
      finished: false,
      started: false
    };
    logEntries.value = [];
  }

  return {
    state,
    teamAName,
    teamBName,
    logEntries,
    remainingMaps,
    currentTeamName,
    availableMaps,
    initializeMaps,
    startVeto,
    onBan,
    swapCurrentTeam,
    resetState,
    log
  };
}
