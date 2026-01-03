import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MapName } from '../types/veto';

export interface MapPool {
  id: number;
  name: string;
  type: 'all' | 'competitive' | 'custom';
  maps: MapName[];
}

export const useVetoStore = defineStore('veto', () => {
  const currentPool = ref<MapPool | null>(null);
  const currentSessionId = ref<number | null>(null);

  function setPool(pool: MapPool) {
    currentPool.value = pool;
  }

  function setSessionId(id: number | null) {
    currentSessionId.value = id;
  }

  function clearPool() {
    currentPool.value = null;
    currentSessionId.value = null;
  }

  return {
    currentPool,
    currentSessionId,
    setPool,
    setSessionId,
    clearPool
  };
});
