import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MapPool } from '../types';

export type BestOfType = 1 | 2 | 3 | 5;

export interface RoomCreationState {
  selectedPool: MapPool | null;
  bestOf: BestOfType | null;
}

export const useRoomCreationStore = defineStore('roomCreation', () => {
  const selectedPool = ref<MapPool | null>(null);
  const bestOf = ref<BestOfType | null>(null);

  function setPool(pool: MapPool) {
    selectedPool.value = pool;
  }

  function setBestOf(value: BestOfType) {
    bestOf.value = value;
  }

  function clear() {
    selectedPool.value = null;
    bestOf.value = null;
  }

  function getState(): RoomCreationState {
    return {
      selectedPool: selectedPool.value,
      bestOf: bestOf.value
    };
  }

  return {
    selectedPool,
    bestOf,
    setPool,
    setBestOf,
    clear,
    getState
  };
});
