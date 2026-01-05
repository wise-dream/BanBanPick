import { ref, computed } from 'vue';
import * as vetoService from '../services/api/vetoService';
import type {
  VetoSessionResponse,
  CreateVetoSessionRequest,
  NextActionResponse,
} from '../services/api/types';
import type { MapName, LogEntry } from '../types/veto';
import { useErrorToast } from './useErrorToast';
import { useI18n } from './useI18n';

export interface VetoSessionState {
  sessionId: number | null;
  session: VetoSessionResponse | null;
  loading: boolean;
  error: string | null;
  currentTeam: 'A' | 'B';
  bans: MapName[];
  pickedMaps: MapName[];
  finished: boolean;
  started: boolean;
  selectedMap: MapName | null;
}

export interface UseVetoSessionOptions {
  currentPool?: () => { maps: Array<{ id: number; name: string }> } | null;
}

export function useVetoSession(options?: UseVetoSessionOptions) {
  const { showError } = useErrorToast();
  const { t } = useI18n();

  const sessionId = ref<number | null>(null);
  const session = ref<VetoSessionResponse | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const logEntries = ref<LogEntry[]>([]);

  // Состояние из session
  const state = computed<VetoSessionState>(() => {
    if (!session.value) {
      return {
        sessionId: null,
        session: null,
        loading: loading.value,
        error: error.value,
        currentTeam: 'A',
        bans: [],
        pickedMaps: [],
        finished: false,
        started: false,
        selectedMap: null,
      };
    }

    // Извлекаем забаненные и выбранные карты из actions
    const bans: MapName[] = [];
    const picks: MapName[] = [];
    const mapsById = new Map<number, string>();

    if (session.value.map_pool?.maps) {
      session.value.map_pool.maps.forEach(map => {
        mapsById.set(map.id, map.name as MapName);
      });
    }

    if (session.value.actions) {
      session.value.actions.forEach(action => {
        // Используем map_id из action для поиска карты
        // action.map может быть undefined, поэтому используем map_id
        const mapId = action.map?.id || action.map_id;
        if (mapId && mapsById.has(mapId)) {
          const mapName = mapsById.get(mapId)!;
          if (action.action_type === 'ban') {
            bans.push(mapName as MapName);
          } else if (action.action_type === 'pick') {
            picks.push(mapName as MapName);
          }
        }
      });
    }

    return {
      sessionId: session.value.id,
      session: session.value,
      loading: loading.value,
      error: error.value,
      currentTeam: (session.value.current_team as 'A' | 'B') || 'A',
      bans,
      pickedMaps: picks,
      finished: session.value.status === 'finished',
      started: session.value.status === 'in_progress' || session.value.status === 'finished',
      selectedMap: (session.value.selected_map_id
        ? (mapsById.get(session.value.selected_map_id) as MapName | undefined) || null
        : null) as MapName | null,
    };
  });

  /**
   * Создание новой сессии вето
   */
  async function createSession(
    poolId: number,
    gameId: number,
    type: 'bo1' | 'bo3' | 'bo5',
    teamAName: string,
    teamBName: string,
    timerSeconds: number = 60
  ): Promise<boolean> {
    loading.value = true;
    error.value = null;

    try {
      const request: CreateVetoSessionRequest = {
        game_id: gameId,
        map_pool_id: poolId,
        type,
        team_a_name: teamAName,
        team_b_name: teamBName,
        timer_seconds: timerSeconds,
      };

      const createdSession = await vetoService.createSession(request);
      session.value = createdSession;
      sessionId.value = createdSession.id;
      
      logEntries.value = [];
      log(t('vetoSession.sessionCreated', { type: type.toUpperCase() }));
      log(t('vetoSession.teamInfo', { teamA: teamAName, teamB: teamBName }));

      return true;
    } catch (err: any) {
      error.value = err.message || 'Не удалось создать сессию';
      showError(error.value);
      return false;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Загрузка сессии по ID или share_token
   */
  async function loadSession(idOrToken: string | number): Promise<boolean> {
    loading.value = true;
    error.value = null;

    try {
      let loadedSession: VetoSessionResponse;

      if (typeof idOrToken === 'number') {
        loadedSession = await vetoService.getSession(idOrToken);
      } else {
        loadedSession = await vetoService.getSessionByToken(idOrToken);
      }

      // Обновляем сессию - это должно триггерить пересчет state computed
      session.value = loadedSession;
      sessionId.value = loadedSession.id;
      
      // Восстанавливаем лог из действий
      rebuildLogFromActions();
      
      // Логируем для отладки
      console.log('Session loaded in useVetoSession:', {
        sessionId: loadedSession.id,
        actionsCount: loadedSession.actions?.length || 0,
        mapPoolMapsCount: loadedSession.map_pool?.maps?.length || 0,
        hasMapPool: !!loadedSession.map_pool,
        hasActions: !!loadedSession.actions
      });

      return true;
    } catch (err: any) {
      error.value = err.message || t('vetoSession.loadError');
      showError(error.value);
      return false;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Обновление сессии напрямую из WebSocket сообщения (без API запроса)
   * Используется для мгновенного обновления состояния без задержек
   */
  function updateSessionFromWebSocket(sessionData: VetoSessionResponse) {
    if (!sessionData || !sessionData.id) {
      console.warn('Invalid session data in WebSocket message')
      return false
    }

    // Обновляем сессию напрямую - это триггерит пересчет state computed
    session.value = sessionData
    sessionId.value = sessionData.id
    
    // Восстанавливаем лог из действий
    rebuildLogFromActions()
    
    console.log('Session updated from WebSocket (direct):', {
      sessionId: sessionData.id,
      actionsCount: sessionData.actions?.length || 0,
      mapPoolMapsCount: sessionData.map_pool?.maps?.length || 0,
      hasMapPool: !!sessionData.map_pool,
      hasActions: !!sessionData.actions
    })

    return true
  }

  /**
   * Забан карты
   */
  async function banMap(mapName: MapName): Promise<boolean> {
    if (!session.value || !sessionId.value) {
      error.value = 'Сессия не загружена';
      return false;
    }

    // Находим ID карты по имени
    // Сначала ищем в session.value.map_pool.maps
    let map = session.value.map_pool?.maps?.find(m => m.name === mapName);
    
    // Fallback: если не найдена, ищем в currentPool
    if (!map && options?.currentPool) {
      const pool = options.currentPool();
      if (pool?.maps) {
        const foundMap = pool.maps.find(m => m.name === mapName);
        if (foundMap) {
          // Проверяем тип карты - может быть Map (из types) или MapResponse (из API)
          if ('slug' in foundMap && 'imageUrl' in foundMap) {
            // Это Map из types (frontend тип)
            const gameMap = foundMap as { id: number; name: string; slug: string; imageUrl: string; isCompetitive: boolean };
            map = {
              id: gameMap.id,
              name: gameMap.name,
              slug: gameMap.slug,
              image_url: gameMap.imageUrl,
              is_competitive: gameMap.isCompetitive,
            };
          } else if ('slug' in foundMap && 'image_url' in foundMap) {
            // Это уже MapResponse (из API)
            map = foundMap as any;
          }
        }
        console.log('Map found in currentPool fallback:', map);
      }
    }
    
    if (!map) {
      console.error('Map not found:', mapName, {
        sessionMapPool: session.value.map_pool,
        currentPool: options?.currentPool?.(),
        availableMaps: session.value.map_pool?.maps?.map(m => m.name)
      });
      error.value = t('vetoSession.mapNotFound');
      return false;
    }

    loading.value = true;
    error.value = null;

    try {
      const updatedSession = await vetoService.banMap(
        sessionId.value,
        map.id,
        state.value.currentTeam
      );

      session.value = updatedSession;
      const teamName = state.value.currentTeam === 'A' ? session.value.team_a_name : session.value.team_b_name;
      log(t('vetoSession.banAction', { teamName, mapName }));

      // Если сессия завершена
      if (updatedSession.status === 'finished' && updatedSession.selected_map_id) {
        const selectedMap = session.value.map_pool?.maps.find(m => m.id === updatedSession.selected_map_id);
        if (selectedMap) {
          log(t('vetoSession.autoPick', { mapName: selectedMap.name }));
        }
      }

      return true;
    } catch (err: any) {
      error.value = err.message || 'Не удалось забанить карту';
      showError(error.value);
      return false;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Выбор карты (pick)
   */
  async function pickMap(mapName: MapName): Promise<boolean> {
    if (!session.value || !sessionId.value) {
      error.value = 'Сессия не загружена';
      return false;
    }

    const map = session.value.map_pool?.maps.find(m => m.name === mapName);
    if (!map) {
      error.value = t('vetoSession.mapNotFound');
      return false;
    }

    loading.value = true;
    error.value = null;

    try {
      const updatedSession = await vetoService.pickMap(
        sessionId.value,
        map.id,
        state.value.currentTeam
      );

      session.value = updatedSession;
      const teamName = state.value.currentTeam === 'A' ? session.value.team_a_name : session.value.team_b_name;
      log(t('vetoSession.pickAction', { teamName, mapName }));

      return true;
    } catch (err: any) {
      error.value = err.message || t('vetoSession.pickError');
      showError(error.value);
      return false;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Выбор стороны (attack/defence) после пика карты
   */
  async function selectSide(side: 'attack' | 'defence', team: 'A' | 'B'): Promise<boolean> {
    if (!session.value || !sessionId.value) {
      error.value = 'Сессия не загружена';
      return false;
    }

    loading.value = true;
    error.value = null;

    try {
      const updatedSession = await vetoService.selectSide(
        sessionId.value,
        side,
        team
      );

      session.value = updatedSession;
      
      // Пересобираем лог из действий после обновления сессии
      rebuildLogFromActions();
      
      const sideText = side === 'attack' ? t('vetoSession.sideAttack') : t('vetoSession.sideDefence');
      const teamName = team === 'A' ? session.value.team_a_name : session.value.team_b_name;
      log(t('vetoSession.sideSelected', { teamName, sideText }));

      return true;
    } catch (err: any) {
      error.value = err.message || t('vetoSession.selectSideError');
      showError(error.value);
      return false;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Сброс сессии
   */
  async function resetSession(): Promise<boolean> {
    if (!sessionId.value) {
      return false;
    }

    loading.value = true;
    error.value = null;

    try {
      const resetSession = await vetoService.resetSession(sessionId.value);
      session.value = resetSession;
      logEntries.value = [];
      log(t('vetoSession.reset'));

      return true;
    } catch (err: any) {
      error.value = err.message || t('vetoSession.resetError');
      showError(error.value);
      return false;
    } finally {
      loading.value = false;
    }
  }

  /**
   * Получение следующего действия
   */
  async function getNextAction(): Promise<NextActionResponse | null> {
    if (!sessionId.value) {
      return null;
    }

    try {
      return await vetoService.getNextAction(sessionId.value);
    } catch (err: any) {
      error.value = err.message || t('vetoSession.nextActionError');
      return null;
    }
  }

  /**
   * Восстановление лога из действий сессии
   */
  function rebuildLogFromActions() {
    if (!session.value || !session.value.actions) {
      return;
    }

    logEntries.value = [];
    const sortedActions = [...session.value.actions].sort((a, b) => a.step_number - b.step_number);

    // Создаем мапу для быстрого поиска карт по ID
    const mapsById = new Map<number, string>();
    if (session.value.map_pool?.maps) {
      session.value.map_pool.maps.forEach(map => {
        mapsById.set(map.id, map.name);
      });
    }

    sortedActions.forEach(action => {
      const teamName = action.team === 'A' ? session.value!.team_a_name : session.value!.team_b_name;
      
      // Используем map_id для поиска карты, так как action.map может быть undefined
      const mapId = action.map?.id || action.map_id;
      const mapName = mapId && mapsById.has(mapId) 
        ? mapsById.get(mapId)! 
        : t('vetoSession.unknownMap');

      if (action.action_type === 'ban') {
        log(t('vetoSession.banAction', { teamName, mapName }));
      } else if (action.action_type === 'pick') {
        log(t('vetoSession.pickAction', { teamName, mapName }));
      }
    });

    if (session.value.status === 'finished' && session.value.selected_map_id) {
      const selectedMap = session.value.map_pool?.maps.find(m => m.id === session.value!.selected_map_id);
      if (selectedMap) {
        log(t('vetoSession.autoPick', { mapName: selectedMap.name }));
      }
    }
  }

  /**
   * Добавление записи в лог
   */
  function log(message: string) {
    logEntries.value.push({
      message,
      timestamp: Date.now(),
    });
  }

  /**
   * Получение доступных карт
   */
  const availableMaps = computed<MapName[]>(() => {
    if (!session.value || !session.value.map_pool) {
      return [];
    }
    return session.value.map_pool.maps.map(m => m.name as MapName);
  });

  /**
   * Получение оставшихся карт
   */
  const remainingMaps = computed<MapName[]>(() => {
    const allMaps = availableMaps.value;
    const bannedAndPicked = [...state.value.bans, ...state.value.pickedMaps];
    return allMaps.filter(map => !bannedAndPicked.includes(map));
  });

  return {
    sessionId,
    session,
    state,
    loading,
    error,
    logEntries,
    availableMaps,
    remainingMaps,
    createSession,
    loadSession,
    updateSessionFromWebSocket,
    banMap,
    pickMap,
    selectSide,
    resetSession,
    getNextAction,
    log,
  };
}
