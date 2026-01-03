import { describe, it, expect, beforeEach } from 'vitest';
import { useVeto } from '../useVeto';
import type { MapName } from '../../types/veto';

describe('useVeto', () => {
  let veto: ReturnType<typeof useVeto>;
  const testMaps: MapName[] = ['Bind', 'Haven', 'Split', 'Ascent', 'Icebox', 'Breeze', 'Fracture'];

  beforeEach(() => {
    veto = useVeto();
    veto.initializeMaps(testMaps);
  });

  describe('initializeMaps', () => {
    it('should initialize maps correctly', () => {
      expect(veto.availableMaps.value).toEqual(testMaps);
      expect(veto.state.value.bans).toEqual([]);
      expect(veto.state.value.finished).toBe(false);
      expect(veto.state.value.started).toBe(false);
    });

    it('should reset state when initializing new maps', () => {
      veto.startVeto();
      veto.onBan('Bind');
      veto.initializeMaps(['NewMap1', 'NewMap2']);
      
      expect(veto.state.value.bans).toEqual([]);
      expect(veto.state.value.started).toBe(false);
      expect(veto.state.value.finished).toBe(false);
    });
  });

  describe('startVeto', () => {
    it('should start veto process', () => {
      veto.startVeto();
      
      expect(veto.state.value.started).toBe(true);
      expect(veto.logEntries.value.length).toBeGreaterThan(0);
    });

    it('should not start if already started', () => {
      veto.startVeto();
      const logCount = veto.logEntries.value.length;
      veto.startVeto();
      
      expect(veto.logEntries.value.length).toBe(logCount);
    });

    it('should not start if finished', () => {
      veto.startVeto();
      // Завершаем процесс
      testMaps.slice(0, -1).forEach(map => veto.onBan(map));
      const logCount = veto.logEntries.value.length;
      veto.startVeto();
      
      expect(veto.logEntries.value.length).toBe(logCount);
    });
  });

  describe('onBan', () => {
    beforeEach(() => {
      veto.startVeto();
    });

    it('should ban a map', () => {
      veto.onBan('Bind');
      
      expect(veto.state.value.bans).toContain('Bind');
      expect(veto.remainingMaps.value).not.toContain('Bind');
      expect(veto.logEntries.value.length).toBeGreaterThan(1);
    });

    it('should switch teams after ban', () => {
      const initialTeam = veto.state.value.currentTeam;
      veto.onBan('Bind');
      
      expect(veto.state.value.currentTeam).not.toBe(initialTeam);
    });

    it('should not ban if veto not started', () => {
      veto.resetState();
      veto.onBan('Bind');
      
      expect(veto.state.value.bans).not.toContain('Bind');
    });

    it('should not ban if already banned', () => {
      veto.onBan('Bind');
      const bansCount = veto.state.value.bans.length;
      veto.onBan('Bind');
      
      expect(veto.state.value.bans.length).toBe(bansCount);
    });

    it('should finish when only one map remains', () => {
      // Баним все карты кроме одной
      testMaps.slice(0, -1).forEach(map => veto.onBan(map));
      
      expect(veto.state.value.finished).toBe(true);
      expect(veto.state.value.pickedMap).toBe(testMaps[testMaps.length - 1]);
      expect(veto.logEntries.value.some(entry => entry.message.includes('Автопик'))).toBe(true);
    });

    it('should return last map when finished', () => {
      const lastMap = testMaps[testMaps.length - 1];
      testMaps.slice(0, -1).forEach(map => veto.onBan(map));
      const result = veto.onBan('SomeMap'); // Попытка забанить последнюю
      
      expect(veto.state.value.pickedMap).toBe(lastMap);
    });
  });

  describe('swapCurrentTeam', () => {
    beforeEach(() => {
      veto.startVeto();
    });

    it('should swap teams', () => {
      const initialTeam = veto.state.value.currentTeam;
      veto.swapCurrentTeam();
      
      expect(veto.state.value.currentTeam).not.toBe(initialTeam);
      expect(veto.logEntries.value.some(entry => entry.message.includes('Ход передан'))).toBe(true);
    });

    it('should not swap if veto not started', () => {
      veto.resetState();
      const initialTeam = veto.state.value.currentTeam;
      veto.swapCurrentTeam();
      
      expect(veto.state.value.currentTeam).toBe(initialTeam);
    });

    it('should not swap if finished', () => {
      testMaps.slice(0, -1).forEach(map => veto.onBan(map));
      const initialTeam = veto.state.value.currentTeam;
      veto.swapCurrentTeam();
      
      expect(veto.state.value.currentTeam).toBe(initialTeam);
    });
  });

  describe('resetState', () => {
    it('should reset state to initial values', () => {
      veto.startVeto();
      veto.onBan('Bind');
      veto.resetState();
      
      expect(veto.state.value.currentTeam).toBe('A');
      expect(veto.state.value.bans).toEqual([]);
      expect(veto.state.value.pickedMap).toBe(null);
      expect(veto.state.value.finished).toBe(false);
      expect(veto.state.value.started).toBe(false);
      expect(veto.logEntries.value).toEqual([]);
    });
  });

  describe('remainingMaps', () => {
    it('should return all maps when no bans', () => {
      expect(veto.remainingMaps.value).toEqual(testMaps);
    });

    it('should exclude banned maps', () => {
      veto.startVeto();
      veto.onBan('Bind');
      veto.onBan('Haven');
      
      expect(veto.remainingMaps.value).not.toContain('Bind');
      expect(veto.remainingMaps.value).not.toContain('Haven');
      expect(veto.remainingMaps.value.length).toBe(testMaps.length - 2);
    });
  });

  describe('currentTeamName', () => {
    it('should return team A name when current team is A', () => {
      expect(veto.state.value.currentTeam).toBe('A');
      expect(veto.currentTeamName.value).toBe('Team A');
    });

    it('should return team B name when current team is B', () => {
      veto.startVeto();
      veto.swapCurrentTeam();
      expect(veto.currentTeamName.value).toBe('Team B');
    });
  });

  describe('log', () => {
    it('should add log entry', () => {
      const initialCount = veto.logEntries.value.length;
      veto.log('Test message');
      
      expect(veto.logEntries.value.length).toBe(initialCount + 1);
      expect(veto.logEntries.value[veto.logEntries.value.length - 1].message).toBe('Test message');
    });

    it('should include timestamp in log entry', () => {
      const before = Date.now();
      veto.log('Test message');
      const after = Date.now();
      
      const entry = veto.logEntries.value[veto.logEntries.value.length - 1];
      expect(entry.timestamp).toBeGreaterThanOrEqual(before);
      expect(entry.timestamp).toBeLessThanOrEqual(after);
    });
  });
});
