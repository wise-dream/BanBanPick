<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick, computed, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useVetoSession } from '../composables/useVetoSession'
import { getPoolById } from '../services/mapPoolService'
import { getRoom, getRoomBySessionId, roomResponseToRoom } from '../services/api/roomService'
import { useAuthStore } from '../store/auth'
import type { MapPool, Room } from '../types'
import type { MapName } from '../types/veto'
import VetoHeader from '../components/VetoHeader.vue'
import MapsGrid from '../components/MapsGrid.vue'
import SummaryPanel from '../components/SummaryPanel.vue'
import FinalOverlay from '../components/FinalOverlay.vue'
import SideOverlay from '../components/SideOverlay.vue'
import * as vetoService from '../services/api/vetoService'
import { useRoomWebSocket } from '../composables/useRoomWebSocket'
import { useErrorToast } from '../composables/useErrorToast'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const poolId = route.params.poolId ? Number(route.params.poolId) : null
const shareToken = route.query.token as string | undefined
const roomIdFromQuery = route.query.room ? Number(route.query.room) : null

// Актуальный roomId - из query или из найденной комнаты
const actualRoomId = computed(() => {
  return roomIdFromQuery || room.value?.id || null
})

const vetoSession = useVetoSession({
  currentPool: () => currentPool.value
})
const { showError: showErrorToast } = useErrorToast()

const showFinalOverlay = ref(false)
const showSideOverlay = ref(false)
const currentPool = ref<MapPool | null>(null)
const room = ref<Room | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const teamAName = ref('Team A')
const teamBName = ref('Team B')
const vetoType = ref<'bo1' | 'bo3' | 'bo5'>('bo1')

// Ключ для принудительного перерендера при сбросе
const mapsGridKey = ref(0)

// Состояние карт: объект с ключами-названиями карт
// Используем объект вместо Map для правильной реактивности Vue
const mapsState = ref<Record<MapName, { isBanned: boolean; isPicked: boolean }>>({} as Record<MapName, { isBanned: boolean; isPicked: boolean }>)

// Оптимистичное обновление для банов (только для конкретной карты)
const optimisticBannedMap = ref<MapName | null>(null)

// WebSocket для real-time обновлений
const roomWs = shallowRef<ReturnType<typeof useRoomWebSocket> | null>(null)
const lastProcessedMessageIndex = ref(-1)

const currentTeamName = computed(() => {
  if (!vetoSession.session.value) return ''
  return vetoSession.state.value.currentTeam === 'A'
    ? vetoSession.session.value.team_a_name
    : vetoSession.session.value.team_b_name
})

// Показываем выбранную карту только если процесс начат И завершен
const pickedMap = computed(() => {
  if (!vetoSession.state.value.started || !vetoSession.state.value.finished) {
    return null
  }
  return vetoSession.state.value.selectedMap
})

const allMaps = computed(() => {
  if (vetoSession.availableMaps.value.length > 0) {
    return vetoSession.availableMaps.value
  }
  if (currentPool.value?.maps) {
    return currentPool.value.maps.map(m => m.name as any)
  }
  return []
})

// Инициализация состояния карт из сессии
function initializeMapsState() {
  const newState: Record<MapName, { isBanned: boolean; isPicked: boolean }> = {} as Record<MapName, { isBanned: boolean; isPicked: boolean }>
  
  if (allMaps.value.length > 0) {
    const sessionBans = new Set(vetoSession.state.value.bans || [])
    // Показываем выбранную карту только если процесс начат И завершен
    const sessionPicked = (vetoSession.state.value.started && vetoSession.state.value.finished) 
      ? vetoSession.state.value.selectedMap 
      : null
    
    allMaps.value.forEach((map) => {
      const mapName = map as MapName
      newState[mapName] = {
        isBanned: sessionBans.has(mapName),
        isPicked: mapName === sessionPicked
      }
    })
  }
  
  mapsState.value = newState
  optimisticBannedMap.value = null
}

// Обновление состояния конкретной карты (без перерендера всего списка)
function updateMapState(mapName: MapName, updates: Partial<{ isBanned: boolean; isPicked: boolean }>) {
  if (mapsState.value[mapName]) {
    mapsState.value[mapName] = {
      ...mapsState.value[mapName],
      ...updates
    }
  }
}

// Проверка, забанена ли карта (computed для каждой карты)
function isMapBanned(mapName: MapName): boolean {
  const state = mapsState.value[mapName]
  if (!state) return false
  
  // Учитываем оптимистичное обновление
  if (optimisticBannedMap.value === mapName) {
    return true
  }
  
  return state.isBanned
}

// Проверка, выбрана ли карта (только если процесс начат и завершен)
function isMapPicked(mapName: MapName): boolean {
  // Не показываем выбранную карту до начала процесса или до завершения
  if (!vetoSession.state.value.started || !vetoSession.state.value.finished) {
    return false
  }
  
  const state = mapsState.value[mapName]
  if (!state) return false
  return state.isPicked
}

const userTeam = computed<'A' | 'B' | null>(() => {
  const isAuthenticated = authStore.isAuthenticated
  const user = authStore.user
  const participants = room.value?.participants
  
  // Логирование для отладки
  console.log('🔍 [userTeam] Computing userTeam:', {
    isAuthenticated,
    hasUser: !!user,
    userId: user?.id,
    hasRoom: !!room.value,
    hasParticipants: !!participants,
    participantsCount: participants?.length || 0,
    participants: participants?.map(p => ({ id: p.id, userId: p.userId, username: p.username, roomId: p.roomId, role: p.role })) || []
  })
  
  if (!isAuthenticated || !user || !participants) {
    console.warn('⚠️ [userTeam] Missing requirements:', {
      isAuthenticated,
      hasUser: !!user,
      hasParticipants: !!participants
    })
    return null
  }

  // Пробуем найти пользователя по userId (с приведением типов для надежности)
  const userIndex = participants.findIndex(
    p => {
      // Приводим к числам для сравнения, так как могут быть строки
      const participantUserId = Number(p.userId)
      const authUserId = Number(user.id)
      const matches = participantUserId === authUserId
      
      console.log('🔍 [userTeam] Comparing:', {
        participantUserId: p.userId,
        participantUserIdNumber: participantUserId,
        userAuthId: user.id,
        userAuthIdNumber: authUserId,
        matches,
        participant: { id: p.id, userId: p.userId, username: p.username, roomId: p.roomId, role: p.role }
      })
      
      return matches
    }
  )
  
  console.log('🔍 [userTeam] User index result:', {
    userId: user.id,
    userIndex,
    participants: participants.map((p, idx) => ({ 
      index: idx, 
      id: p.id,
      userId: p.userId, 
      username: p.username,
      roomId: p.roomId,
      role: p.role
    }))
  })

  if (userIndex === 0) {
    console.log('✅ [userTeam] User is Team A')
    return 'A'
  }
  if (userIndex === 1) {
    console.log('✅ [userTeam] User is Team B')
    return 'B'
  }
  
  console.warn('⚠️ [userTeam] User not found in participants')
  return null
})

const teamAParticipantUsername = computed(() => {
  if (!room.value?.participants || room.value.participants.length === 0) {
    return null
  }
  return room.value.participants[0]?.username || null
})

const teamBParticipantUsername = computed(() => {
  if (!room.value?.participants || room.value.participants.length < 2) {
    return null
  }
  return room.value.participants[1]?.username || null
})

const canBan = computed(() => {
  const started = vetoSession.state.value.started
  const finished = vetoSession.state.value.finished
  const currentTeam = vetoSession.state.value.currentTeam
  const userTeamValue = userTeam.value
  
  // Логирование для отладки
  if (started && !finished) {
    console.log('🔍 [canBan] Computing canBan:', {
      started,
      finished,
      currentTeam,
      userTeam: userTeamValue,
      isAuthenticated: authStore.isAuthenticated,
      hasUser: !!authStore.user,
      userId: authStore.user?.id,
      hasRoom: !!room.value,
      hasParticipants: !!room.value?.participants,
      participantsCount: room.value?.participants?.length || 0,
      participants: room.value?.participants?.map(p => ({ userId: p.userId, username: p.username })) || []
    })
  }
  
  if (!started) {
    return false
  }
  
  if (finished) {
    return false
  }
  
  // Если пользователь не авторизован или не найден в участниках комнаты, запрещаем банить
  if (userTeamValue === null) {
    console.warn('⚠️ [canBan] userTeam is null:', {
      isAuthenticated: authStore.isAuthenticated,
      hasUser: !!authStore.user,
      userId: authStore.user?.id,
      hasRoom: !!room.value,
      hasParticipants: !!room.value?.participants,
      participantsCount: room.value?.participants?.length || 0
    })
    return false
  }
  
  // Разрешаем банить только если очередь текущего пользователя
  const result = currentTeam === userTeamValue
  console.log('✅ [canBan] Result:', {
    currentTeam,
    userTeam: userTeamValue,
    canBan: result
  })
  return result
})

// Функция для обновления названий команд из сессии с fallback на никнеймы из комнаты
function updateTeamNamesFromSession() {
  if (!vetoSession.session.value) return
  
  const sessionTeamA = vetoSession.session.value.team_a_name
  const sessionTeamB = vetoSession.session.value.team_b_name
  
  const isOldTeamA = sessionTeamA === 'Team 1' || sessionTeamA === 'Team A' || sessionTeamA?.startsWith('Team ')
  const isOldTeamB = sessionTeamB === 'Team 2' || sessionTeamB === 'Team B' || sessionTeamB?.startsWith('Team ')
  
  if (isOldTeamA && teamAParticipantUsername.value) {
    teamAName.value = teamAParticipantUsername.value
  } else if (sessionTeamA) {
    teamAName.value = sessionTeamA
  } else if (teamAParticipantUsername.value) {
    teamAName.value = teamAParticipantUsername.value
  }
  
  if (isOldTeamB && teamBParticipantUsername.value) {
    teamBName.value = teamBParticipantUsername.value
  } else if (sessionTeamB) {
    teamBName.value = sessionTeamB
  } else if (teamBParticipantUsername.value) {
    teamBName.value = teamBParticipantUsername.value
  }
}

// Следим за изменениями сессии и обновляем состояние карт ТОЛЬКО при реальных изменениях
// НЕ обновляем все карты - только те, что действительно изменились
watch(
  () => vetoSession.state.value.bans,
  (newBans, oldBans) => {
    if (!allMaps.value.length) return
    
    const newBansSet = new Set(newBans || [])
    const oldBansSet = new Set(oldBans || [])
    
    // Находим только измененные карты
    allMaps.value.forEach(map => {
      const wasBanned = oldBansSet.has(map)
      const isBanned = newBansSet.has(map)
      
      if (wasBanned !== isBanned) {
        // Обновляем только эту карту
        updateMapState(map, { isBanned })
        if (optimisticBannedMap.value === map && isBanned) {
          optimisticBannedMap.value = null
        }
      }
    })
  },
  { deep: true }
)

// Следим за изменениями выбранной карты (только если процесс начат и завершен)
watch(
  () => {
    // Возвращаем выбранную карту только если процесс завершен
    if (!vetoSession.state.value.started || !vetoSession.state.value.finished) {
      return null
    }
    return vetoSession.state.value.selectedMap
  },
  (newPicked, oldPicked) => {
    console.log('👀 [Watch selectedMap] Changed:', {
      oldPicked,
      newPicked,
      started: vetoSession.state.value.started,
      finished: vetoSession.state.value.finished,
      allMapsCount: allMaps.value.length
    })
    
    if (!allMaps.value.length) return
    
    // Сбрасываем старую выбранную карту (включая случай когда newPicked === null)
    if (oldPicked && mapsState.value[oldPicked]) {
      console.log('🔄 [Watch selectedMap] Clearing old picked map:', oldPicked)
      updateMapState(oldPicked, { isPicked: false })
    }
    
    // Устанавливаем новую выбранную карту (или очищаем если newPicked === null)
    if (newPicked && mapsState.value[newPicked]) {
      console.log('✅ [Watch selectedMap] Setting new picked map:', newPicked)
      updateMapState(newPicked, { isPicked: true })
    } else if (!newPicked) {
      // Если newPicked === null, очищаем все карты
      console.log('🧹 [Watch selectedMap] No picked map, ensuring all maps are cleared')
      allMaps.value.forEach((map) => {
        const mapName = map as MapName
        if (mapsState.value[mapName]?.isPicked) {
          updateMapState(mapName, { isPicked: false })
        }
      })
    }
  }
)

onMounted(async () => {
  console.log('🚀 [VetoProcessPage] onMounted started:', {
    roomIdFromQuery,
    poolId,
    shareToken: !!shareToken,
    sessionId: route.query.session,
    isAuthenticated: authStore.isAuthenticated,
    timestamp: new Date().toISOString()
  })
  
  if (roomIdFromQuery) {
    console.log('🏠 [VetoProcessPage] Loading room by roomId:', {
      roomId: roomIdFromQuery,
      timestamp: new Date().toISOString()
    })
    
    try {
      if (!roomIdFromQuery) {
        console.error('❌ [VetoProcessPage] roomIdFromQuery is null')
        return
      }
      const roomData = await getRoom(roomIdFromQuery)
      console.log('✅ [VetoProcessPage] Room loaded:', {
        roomId: roomData.id,
        name: roomData.name,
        vetoSessionId: roomData.veto_session_id,
        participantsCount: roomData.participants?.length || 0
      })
      
      // Используем функцию roomResponseToRoom для правильного маппинга participants
      room.value = roomResponseToRoom(roomData)
      
      if (vetoSession.session.value) {
        updateTeamNamesFromSession()
      } else {
        if (teamAParticipantUsername.value) {
          teamAName.value = teamAParticipantUsername.value
        }
        if (teamBParticipantUsername.value) {
          teamBName.value = teamBParticipantUsername.value
        }
      }
      
      if (room.value && authStore.isAuthenticated) {
        console.log('🔌 [VetoProcessPage] Initializing WebSocket connection:', {
          roomId: room.value.id,
          hasRoom: !!room.value,
          isAuthenticated: authStore.isAuthenticated,
          timestamp: new Date().toISOString()
        })
        
        roomWs.value = useRoomWebSocket(room.value.id)
        roomWs.value.connect()
        
        console.log('✅ [VetoProcessPage] WebSocket initialized:', {
          roomId: room.value.id,
          hasWebSocket: !!roomWs.value,
          isConnected: roomWs.value?.isConnected.value
        })
      } else {
        console.warn('⚠️ [VetoProcessPage] Cannot initialize WebSocket:', {
          roomId: roomIdFromQuery,
          hasRoom: !!room.value,
          isAuthenticated: authStore.isAuthenticated
        })
      }
    } catch (err: any) {
      console.error('❌ [VetoProcessPage] Error loading room:', {
        roomIdFromQuery,
        error: err,
        message: err.message
      })
    }
  } else {
    console.log('ℹ️ [VetoProcessPage] No roomId in URL, will try to get from session')
  }

  const sessionId = route.query.session as string | undefined
  if (sessionId) {
    console.log('📋 [VetoProcessPage] Loading session from query:', {
      sessionId: Number(sessionId),
      timestamp: new Date().toISOString()
    })
    await loadSessionById(Number(sessionId))
    return
  }

  if (room.value?.vetoSessionId) {
    console.log('📋 [VetoProcessPage] Loading session from room:', {
      sessionId: room.value.vetoSessionId,
      timestamp: new Date().toISOString()
    })
    await loadSessionById(room.value.vetoSessionId)
    return
  }

  if (shareToken) {
    console.log('🔗 [VetoProcessPage] Loading session by token:', {
      hasToken: !!shareToken,
      timestamp: new Date().toISOString()
    })
    await loadSessionByToken(shareToken)
    return
  }

  if (poolId) {
    console.log('🎮 [VetoProcessPage] Loading pool:', {
      poolId,
      timestamp: new Date().toISOString()
    })
    await loadPool()
  } else {
    console.error('❌ [VetoProcessPage] No poolId or session token provided')
    error.value = 'Не указан пул карт или токен сессии'
    loading.value = false
  }
  
  console.log('✅ [VetoProcessPage] onMounted completed:', {
    hasSession: !!vetoSession.session.value,
    sessionId: vetoSession.sessionId.value,
    hasRoom: !!room.value,
    roomId: room.value?.id,
    hasWebSocket: !!roomWs.value,
    isConnected: roomWs.value?.isConnected.value,
    timestamp: new Date().toISOString()
  })
})

onUnmounted(() => {
  console.log('🔌 [VetoProcessPage] onUnmounted: cleaning up WebSocket:', {
    hasWebSocket: !!roomWs.value,
    roomIdFromQuery,
    actualRoomId: actualRoomId.value,
    roomValueId: room.value?.id,
    isConnected: roomWs.value?.isConnected.value,
    timestamp: new Date().toISOString()
  })
  
  if (roomWs.value) {
    roomWs.value.disconnect()
    roomWs.value = null
  }
  
  console.log('✅ [VetoProcessPage] WebSocket cleaned up')
})

const loadPool = async () => {
  if (!poolId) return

  loading.value = true
  error.value = null

  try {
    const pool = await getPoolById(poolId)
    if (!pool) {
      error.value = 'Пул карт не найден'
      loading.value = false
      return
    }
    currentPool.value = pool
    initializeMapsState()
  } catch (err: any) {
    error.value = err.message || 'Не удалось загрузить пул карт'
  } finally {
    loading.value = false
  }
}

const loadSessionById = async (id: number) => {
  const currentRoomIdAtStart = actualRoomId.value
  console.log('📋 [loadSessionById] Starting to load session:', {
    sessionId: id,
    roomIdFromQuery,
    actualRoomId: currentRoomIdAtStart,
    roomValueId: room.value?.id,
    hasRoom: !!room.value,
    hasWebSocket: !!roomWs.value,
    isAuthenticated: authStore.isAuthenticated,
    timestamp: new Date().toISOString()
  })
  
  loading.value = true
  error.value = null

  try {
    const success = await vetoSession.loadSession(id)
    console.log('📋 [loadSessionById] Session loaded from API:', {
      success,
      sessionId: vetoSession.sessionId.value,
      hasSession: !!vetoSession.session.value,
      sessionStatus: vetoSession.session.value?.status,
      roomIdFromQuery,
      actualRoomId: actualRoomId.value,
      roomValueId: room.value?.id,
      hasWebSocket: !!roomWs.value,
      roomVetoSessionId: room.value?.vetoSessionId
    })
    
    if (success && vetoSession.session.value) {
      updateTeamNamesFromSession()
      
      // КРИТИЧНО: Если нет roomId, пытаемся получить комнату по sessionId
      if (!currentRoomIdAtStart && !roomWs.value) {
      console.log('🔍 [loadSessionById] No actualRoomId, attempting to find room by sessionId:', {
        sessionId: id,
        actualRoomId: actualRoomId.value,
        roomIdFromQuery,
        roomValueId: room.value?.id,
        timestamp: new Date().toISOString()
      })
        
        try {
          const roomData = await getRoomBySessionId(id)
          
          if (roomData) {
            console.log('✅ [loadSessionById] Room found by sessionId:', {
              sessionId: id,
              roomId: roomData.id,
              roomName: roomData.name,
              hasParticipants: !!roomData.participants?.length,
              participantsCount: roomData.participants?.length || 0
            })
            
            // Обновляем room.value с данными найденной комнаты
            // Используем функцию roomResponseToRoom для правильного маппинга participants
            room.value = roomResponseToRoom(roomData)
            
            // Обновляем teamAName и teamBName из участников комнаты
            if (room.value.participants && room.value.participants.length > 0) {
              if (room.value.participants[0]?.username) {
                teamAName.value = room.value.participants[0].username
              }
              if (room.value.participants[1]?.username) {
                teamBName.value = room.value.participants[1].username
              }
            }
            
            // Подключаем WebSocket если пользователь авторизован
            if (authStore.isAuthenticated && room.value) {
              console.log('🔌 [loadSessionById] Initializing WebSocket for found room:', {
                sessionId: id,
                roomId: room.value.id,
                isAuthenticated: authStore.isAuthenticated,
                timestamp: new Date().toISOString()
              })
              
              roomWs.value = useRoomWebSocket(room.value.id)
              roomWs.value.connect()
              
              console.log('✅ [loadSessionById] WebSocket connection initiated:', {
                sessionId: id,
                roomId: room.value.id,
                hasWebSocket: !!roomWs.value,
                isConnected: roomWs.value?.isConnected.value
              })
            } else {
              console.warn('⚠️ [loadSessionById] Cannot connect WebSocket:', {
                sessionId: id,
                roomId: room.value.id,
                isAuthenticated: authStore.isAuthenticated,
                hasRoom: !!room.value
              })
            }
          } else {
            console.warn('⚠️ [loadSessionById] Room not found by sessionId:', {
              sessionId: id,
              message: 'Session is not linked to any room'
            })
          }
        } catch (err: any) {
          console.error('❌ [loadSessionById] Error getting room by sessionId:', {
            sessionId: id,
            error: err,
            message: err.message,
            timestamp: new Date().toISOString()
          })
        }
      } else if (actualRoomId.value && !roomWs.value && authStore.isAuthenticated) {
        // Если roomId есть, но WebSocket не подключен, подключаем
        console.log('🔌 [loadSessionById] RoomId exists but WebSocket not connected, connecting:', {
          sessionId: id,
          roomId: actualRoomId.value,
          isAuthenticated: authStore.isAuthenticated,
          timestamp: new Date().toISOString()
        })
        
        roomWs.value = useRoomWebSocket(actualRoomId.value)
        roomWs.value.connect()
        
        console.log('✅ [loadSessionById] WebSocket connection initiated for existing roomId:', {
          sessionId: id,
          roomId: actualRoomId.value,
          isConnected: roomWs.value?.isConnected.value
        })
      } else {
        console.log('ℹ️ [loadSessionById] WebSocket connection status:', {
          sessionId: id,
          actualRoomId: actualRoomId.value,
          roomIdFromQuery,
          roomValueId: room.value?.id,
          hasRoomId: !!actualRoomId.value,
          hasWebSocket: !!roomWs.value,
          isAuthenticated: authStore.isAuthenticated,
          reason: !actualRoomId.value ? 'No roomId' : !authStore.isAuthenticated ? 'Not authenticated' : 'Already connected'
        })
      }

      const type = vetoSession.session.value.type.toLowerCase()
      if (type === 'bo1' || type === 'bo3' || type === 'bo5') {
        vetoType.value = type
      }

      if (vetoSession.session.value.map_pool) {
        const mapPool = vetoSession.session.value.map_pool
        if (mapPool.maps && mapPool.maps.length > 0) {
          currentPool.value = {
            id: mapPool.id,
            gameId: mapPool.game_id,
            name: mapPool.name,
            type: mapPool.type as 'all' | 'competitive' | 'custom',
            isSystem: mapPool.is_system || false,
            maps: mapPool.maps.map(m => ({
              id: m.id,
              name: m.name,
              slug: m.slug,
              imageUrl: m.image_url,
              isCompetitive: m.is_competitive,
            })),
          }
          initializeMapsState()
        } else if (vetoSession.session.value.map_pool_id) {
          const pool = await getPoolById(vetoSession.session.value.map_pool_id)
          if (pool && pool.maps?.length) {
            currentPool.value = pool
            initializeMapsState()
          } else {
            error.value = 'Пул карт не содержит карт'
          }
        }
      } else if (vetoSession.session.value.map_pool_id) {
        const pool = await getPoolById(vetoSession.session.value.map_pool_id)
        if (pool && pool.maps?.length) {
          currentPool.value = pool
          initializeMapsState()
        } else {
          error.value = 'Пул карт не найден или не содержит карт'
        }
      } else {
        error.value = 'Пул карт не указан в сессии'
      }
    }
  } catch (err: any) {
    error.value = err.message || 'Не удалось загрузить сессию'
  } finally {
    loading.value = false
  }
}

const loadSessionByToken = async (token: string) => {
  loading.value = true
  error.value = null

  try {
    const success = await vetoSession.loadSession(token)
    if (success && vetoSession.session.value) {
      updateTeamNamesFromSession()

      const type = vetoSession.session.value.type.toLowerCase()
      if (type === 'bo1' || type === 'bo3' || type === 'bo5') {
        vetoType.value = type
      }

      if (vetoSession.session.value.map_pool) {
        const mapPool = vetoSession.session.value.map_pool
        if (mapPool.maps && mapPool.maps.length > 0) {
          currentPool.value = {
            id: mapPool.id,
            gameId: mapPool.game_id,
            name: mapPool.name,
            type: mapPool.type as 'all' | 'competitive' | 'custom',
            isSystem: mapPool.is_system || false,
            maps: mapPool.maps.map(m => ({
              id: m.id,
              name: m.name,
              slug: m.slug,
              imageUrl: m.image_url,
              isCompetitive: m.is_competitive,
            })),
          }
          initializeMapsState()
        } else if (vetoSession.session.value.map_pool_id) {
          const pool = await getPoolById(vetoSession.session.value.map_pool_id)
          if (pool && pool.maps?.length) {
            currentPool.value = pool
            initializeMapsState()
          } else {
            error.value = 'Пул карт не содержит карт'
          }
        }
      } else if (vetoSession.session.value.map_pool_id) {
        const pool = await getPoolById(vetoSession.session.value.map_pool_id)
        if (pool && pool.maps?.length) {
          currentPool.value = pool
          initializeMapsState()
        } else {
          error.value = 'Пул карт не найден или не содержит карт'
        }
      } else {
        error.value = 'Пул карт не указан в сессии'
      }
    }
  } catch (err: any) {
    const errorMessage = err.message || 'Не удалось загрузить сессию'
    error.value = errorMessage
    showErrorToast(err)
  } finally {
    loading.value = false
  }
}

// Обрабатываем WebSocket сообщения для real-time обновлений
watch(
  () => roomWs.value?.messages.value,
  (messages) => {
    const currentRoomIdInWatch = actualRoomId.value
    console.log('👀 [WebSocket Watch] Messages changed:', {
      messagesCount: messages?.length || 0,
      lastProcessedIndex: lastProcessedMessageIndex.value,
      hasWebSocket: !!roomWs.value,
      isConnected: roomWs.value?.isConnected.value,
      roomIdFromQuery,
      actualRoomId: currentRoomIdInWatch,
      roomValueId: room.value?.id,
      sessionId: vetoSession.sessionId.value,
      timestamp: new Date().toISOString()
    })
    
    if (!roomWs.value) {
      console.warn('⚠️ [WebSocket Watch] No WebSocket instance:', {
        roomIdFromQuery,
        sessionId: vetoSession.sessionId.value
      })
      return
    }
    
    if (!messages || messages.length === 0) {
      console.log('ℹ️ [WebSocket Watch] No messages to process')
      return
    }
    
    const startIndex = lastProcessedMessageIndex.value + 1
    const newMessages = messages.slice(startIndex)
    
    console.log(`📋 [WebSocket Watch] Processing ${newMessages.length} new messages (from index ${startIndex}):`, {
      totalMessages: messages.length,
      newMessagesCount: newMessages.length,
      messageTypes: newMessages.map(m => m?.type).filter(Boolean),
      actualRoomId: currentRoomIdInWatch,
      sessionId: vetoSession.sessionId.value
    })
    
    for (let i = 0; i < newMessages.length; i++) {
      const message = newMessages[i]
      if (!message) {
        console.warn(`⚠️ [WebSocket Watch] Message at index ${startIndex + i} is undefined`)
        continue
      }
      
      console.log(`🔄 [WebSocket Watch] Processing message ${i + 1}/${newMessages.length}:`, {
        type: message.type,
        hasData: !!message.data,
        index: startIndex + i,
        roomIdFromQuery,
        sessionId: vetoSession.sessionId.value
      })
      handleWebSocketMessage(message)
    }
    
    lastProcessedMessageIndex.value = messages.length - 1
    console.log(`✅ [WebSocket Watch] Processed all messages. Last index: ${lastProcessedMessageIndex.value}`, {
      totalProcessed: newMessages.length,
      actualRoomId: currentRoomIdInWatch,
      sessionId: vetoSession.sessionId.value
    })
  },
  { deep: true, flush: 'post' }
)

watch(roomWs, (newWs, oldWs) => {
  console.log('👀 [WebSocket Instance Watch] WebSocket instance changed:', {
    hadWebSocket: !!oldWs,
    hasWebSocket: !!newWs,
    oldRoomId: oldWs ? 'exists' : 'none',
    newRoomId: newWs ? 'exists' : 'none',
    isConnected: newWs?.isConnected.value,
    messagesCount: newWs?.messages.value?.length || 0,
    timestamp: new Date().toISOString()
  })
  
  lastProcessedMessageIndex.value = -1
  console.log('🔄 [WebSocket Instance Watch] Reset lastProcessedMessageIndex to -1')
})

// Обработка WebSocket сообщений
function handleWebSocketMessage(message: any) {
  const messageTimestamp = new Date().toISOString()
  const currentRoomIdInMessage = actualRoomId.value
  
  console.log('📨 [handleWebSocketMessage] Processing message:', {
    messageType: message?.type,
    hasData: !!message?.data,
    timestamp: messageTimestamp,
    roomIdFromQuery,
    actualRoomId: currentRoomIdInMessage,
    roomValueId: room.value?.id,
    sessionId: vetoSession.sessionId.value,
    fullMessage: message
  })
  
  if (!message || !message.type) {
    console.error('❌ [handleWebSocketMessage] Invalid message format:', {
      message,
      timestamp: messageTimestamp
    })
    return
  }
  
  const stateBefore = {
    started: vetoSession.state.value.started,
    finished: vetoSession.state.value.finished,
    currentTeam: vetoSession.state.value.currentTeam,
    bans: [...(vetoSession.state.value.bans || [])],
    selectedMap: vetoSession.state.value.selectedMap,
    mapsStateCount: Object.keys(mapsState.value).length
  }
  
  console.log('📊 [handleWebSocketMessage] State before processing:', {
    messageType: message.type,
    stateBefore,
    timestamp: messageTimestamp
  })
  
  switch (message.type) {
    case 'veto:ban':
      console.log('🚫 [VETO:BAN] Processing ban message:', {
        hasSession: !!message.data?.session,
        hasAction: !!message.data?.action,
        actionType: message.data?.action?.action_type,
        mapId: message.data?.action?.map_id,
        sessionId: message.data?.session?.id,
        timestamp: messageTimestamp
      })
      
      if (message.data?.session) {
        const sessionData = message.data.session
        
        if (sessionData.map_pool && sessionData.actions !== undefined) {
          console.log('📋 [VETO:BAN] Session data valid:', {
            sessionId: sessionData.id,
            status: sessionData.status,
            actionsCount: sessionData.actions?.length || 0,
            mapsCount: sessionData.map_pool?.maps?.length || 0,
            currentTeam: sessionData.current_team
          })
          
          const success = vetoSession.updateSessionFromWebSocket(sessionData)
          
          console.log('📊 [VETO:BAN] updateSessionFromWebSocket result:', {
            success,
            newStatus: vetoSession.state.value.started,
            newBans: vetoSession.state.value.bans,
            newCurrentTeam: vetoSession.state.value.currentTeam
          })
          
          if (success) {
            updateTeamNamesFromSession()
            
            // Обновляем состояние конкретной карты, если она была забанена/выбрана
            if (message.data?.action?.map_id) {
              const actionMapId = message.data.action.map_id
              const actionMap = sessionData.map_pool?.maps?.find((m: any) => m.id === actionMapId)
              
              console.log('🎯 [VETO:BAN] Processing action:', {
                actionMapId,
                foundMap: !!actionMap,
                mapName: actionMap?.name,
                actionType: message.data.action.action_type
              })
              
              if (actionMap) {
                const mapName = actionMap.name as MapName
                const isBan = message.data.action.action_type === 'ban'
                
                if (isBan) {
                  console.log('✅ [VETO:BAN] Updating map state to banned:', {
                    mapName,
                    previousState: mapsState.value[mapName],
                    optimisticBanned: optimisticBannedMap.value
                  })
                  
                  updateMapState(mapName, { isBanned: true })
                  if (optimisticBannedMap.value === mapName) {
                    optimisticBannedMap.value = null
                    console.log('🧹 [VETO:BAN] Cleared optimistic ban for:', mapName)
                  }
                }
              }
            }
            
            if (vetoSession.state.value.finished && pickedMap.value) {
              console.log('🏁 [VETO:BAN] Session finished, showing final overlay')
              showFinalOverlay.value = true
            }
            
            const stateAfter = {
              started: vetoSession.state.value.started,
              finished: vetoSession.state.value.finished,
              currentTeam: vetoSession.state.value.currentTeam,
              bans: [...(vetoSession.state.value.bans || [])],
              selectedMap: vetoSession.state.value.selectedMap
            }
            
            console.log('✅ [VETO:BAN] Processing complete:', {
              stateBefore,
              stateAfter,
              changed: JSON.stringify(stateBefore) !== JSON.stringify(stateAfter)
            })
          } else {
            console.error('❌ [VETO:BAN] updateSessionFromWebSocket failed')
          }
        } else {
          console.warn('⚠️ [VETO:BAN] WebSocket message missing map_pool or actions:', {
            hasMapPool: !!sessionData.map_pool,
            hasActions: sessionData.actions !== undefined,
            waitingForNextMessage: true
          })
        }
      } else {
        console.warn('⚠️ [VETO:BAN] WebSocket message missing session data')
      }
      break
      
    case 'veto:pick':
      console.log('🎯 [VETO:PICK] Processing pick message:', {
        hasSession: !!message.data?.session,
        hasAction: !!message.data?.action,
        actionType: message.data?.action?.action_type,
        mapId: message.data?.action?.map_id,
        sessionId: message.data?.session?.id,
        timestamp: messageTimestamp
      })
      
      if (message.data?.session) {
        const sessionData = message.data.session
        
        if (sessionData.map_pool && sessionData.actions !== undefined) {
          console.log('📋 [VETO:PICK] Session data valid:', {
            sessionId: sessionData.id,
            status: sessionData.status,
            actionsCount: sessionData.actions?.length || 0,
            mapsCount: sessionData.map_pool?.maps?.length || 0,
            currentTeam: sessionData.current_team,
            selectedMapId: sessionData.selected_map_id
          })
          
          const success = vetoSession.updateSessionFromWebSocket(sessionData)
          
          console.log('📊 [VETO:PICK] updateSessionFromWebSocket result:', {
            success,
            newStatus: vetoSession.state.value.started,
            newFinished: vetoSession.state.value.finished,
            newSelectedMap: vetoSession.state.value.selectedMap,
            newCurrentTeam: vetoSession.state.value.currentTeam
          })
          
          if (success) {
            updateTeamNamesFromSession()
            
            // Обновляем состояние конкретной карты, если она была выбрана
            if (message.data?.action?.map_id) {
              const actionMapId = message.data.action.map_id
              const actionMap = sessionData.map_pool?.maps?.find((m: any) => m.id === actionMapId)
              
              console.log('🎯 [VETO:PICK] Processing action:', {
                actionMapId,
                foundMap: !!actionMap,
                mapName: actionMap?.name,
                actionType: message.data.action.action_type,
                isFinished: vetoSession.state.value.finished
              })
              
              if (actionMap) {
                const mapName = actionMap.name as MapName
                const isPick = message.data.action.action_type === 'pick'
                
                if (isPick) {
                  // Помечаем карту как выбранную только если процесс завершен
                  if (vetoSession.state.value.finished) {
                    console.log('✅ [VETO:PICK] Updating map state to picked:', {
                      mapName,
                      previousState: mapsState.value[mapName],
                      finished: vetoSession.state.value.finished
                    })
                    
                    updateMapState(mapName, { isPicked: true })
                  } else {
                    console.log('⏳ [VETO:PICK] Process not finished yet, skipping pick update:', {
                      mapName,
                      finished: vetoSession.state.value.finished
                    })
                  }
                }
              }
            }
            
            if (vetoSession.state.value.finished && pickedMap.value) {
              console.log('🏁 [VETO:PICK] Session finished, showing final overlay')
              showFinalOverlay.value = true
            }
            
            const stateAfter = {
              started: vetoSession.state.value.started,
              finished: vetoSession.state.value.finished,
              currentTeam: vetoSession.state.value.currentTeam,
              bans: [...(vetoSession.state.value.bans || [])],
              selectedMap: vetoSession.state.value.selectedMap
            }
            
            console.log('✅ [VETO:PICK] Processing complete:', {
              stateBefore,
              stateAfter,
              changed: JSON.stringify(stateBefore) !== JSON.stringify(stateAfter)
            })
          } else {
            console.error('❌ [VETO:PICK] updateSessionFromWebSocket failed')
          }
        } else {
          console.warn('⚠️ [VETO:PICK] WebSocket message missing map_pool or actions:', {
            hasMapPool: !!sessionData.map_pool,
            hasActions: sessionData.actions !== undefined,
            waitingForNextMessage: true
          })
        }
      } else {
        console.warn('⚠️ [VETO:PICK] WebSocket message missing session data')
      }
      break
    case 'veto:start':
      console.log('▶️ [VETO:START] Processing start message:', {
        hasSession: !!message.data?.session,
        sessionId: message.data?.session?.id,
        userId: message.data?.user_id,
        timestamp: messageTimestamp
      })
      
      if (message.data?.session) {
        const sessionData = message.data.session
        
        console.log('📋 [VETO:START] Session data:', {
          sessionId: sessionData.id,
          status: sessionData.status,
          hasMapPool: !!sessionData.map_pool,
          hasActions: sessionData.actions !== undefined,
          actionsCount: sessionData.actions?.length || 0,
          mapsCount: sessionData.map_pool?.maps?.length || 0,
          teamA: sessionData.team_a_name,
          teamB: sessionData.team_b_name
        })
        
        // Для veto:start достаточно наличия map_pool, actions может быть undefined/null/[] (нормально для только что стартованной сессии)
        if (sessionData.map_pool) {
          // Если actions отсутствует, устанавливаем пустой массив для корректной обработки
          const sessionDataWithActions = {
            ...sessionData,
            actions: sessionData.actions || []
          }
          
          const success = vetoSession.updateSessionFromWebSocket(sessionDataWithActions)
          
          console.log('📊 [VETO:START] updateSessionFromWebSocket result:', {
            success,
            newStarted: vetoSession.state.value.started,
            newFinished: vetoSession.state.value.finished,
            newCurrentTeam: vetoSession.state.value.currentTeam,
            newBans: vetoSession.state.value.bans
          })
          
          if (success) {
            updateTeamNamesFromSession()
            initializeMapsState()
            
            const stateAfter = {
              started: vetoSession.state.value.started,
              finished: vetoSession.state.value.finished,
              currentTeam: vetoSession.state.value.currentTeam,
              bans: [...(vetoSession.state.value.bans || [])],
              mapsStateCount: Object.keys(mapsState.value).length
            }
            
            console.log('✅ [VETO:START] Session initialized:', {
              stateBefore,
              stateAfter,
              changed: JSON.stringify(stateBefore) !== JSON.stringify(stateAfter)
            })
          } else {
            console.error('❌ [VETO:START] updateSessionFromWebSocket failed')
          }
        } else {
          console.warn('⚠️ [VETO:START] WebSocket message missing map_pool:', {
            hasMapPool: !!sessionData.map_pool,
            hasSession: !!sessionData
          })
        }
      } else {
        console.warn('⚠️ [VETO:START] WebSocket message missing session data')
      }
      break
    case 'veto:reset':
      console.log('🔄 [VETO:RESET] Processing reset message:', {
        hasSession: !!message.data?.session,
        sessionId: message.data?.session?.id,
        userId: message.data?.user_id,
        timestamp: messageTimestamp
      })
      
      if (message.data?.session) {
        const sessionData = message.data.session
        
        console.log('📋 [VETO:RESET] Session data:', {
          sessionId: sessionData.id,
          status: sessionData.status,
          hasMapPool: !!sessionData.map_pool,
          hasActions: sessionData.actions !== undefined,
          actionsCount: sessionData.actions?.length || 0,
          mapsCount: sessionData.map_pool?.maps?.length || 0
        })
        
        // Для veto:reset достаточно наличия map_pool, actions может быть undefined/null/[] (нормально для сброшенной сессии)
        if (sessionData.map_pool) {
          // Если actions отсутствует, устанавливаем пустой массив для корректной обработки
          const sessionDataWithActions = {
            ...sessionData,
            actions: sessionData.actions || []
          }
          
          const success = vetoSession.updateSessionFromWebSocket(sessionDataWithActions)
          
          console.log('📊 [VETO:RESET] updateSessionFromWebSocket result:', {
            success,
            newStarted: vetoSession.state.value.started,
            newFinished: vetoSession.state.value.finished,
            newBans: vetoSession.state.value.bans,
            newSelectedMap: vetoSession.state.value.selectedMap
          })
          
          if (success) {
            updateTeamNamesFromSession()
            
            const uiStateBefore = {
              showFinalOverlay: showFinalOverlay.value,
              showSideOverlay: showSideOverlay.value,
              optimisticBannedMap: optimisticBannedMap.value,
              mapsGridKey: mapsGridKey.value,
              mapsStateCount: Object.keys(mapsState.value).length
            }
            
            showFinalOverlay.value = false
            showSideOverlay.value = false
            optimisticBannedMap.value = null
            
            // Принудительно перерендериваем все карты при сбросе
            mapsGridKey.value++
            
            // Полностью очищаем состояние всех карт
            const newState: Record<MapName, { isBanned: boolean; isPicked: boolean }> = {} as Record<MapName, { isBanned: boolean; isPicked: boolean }>
            if (allMaps.value.length > 0) {
              allMaps.value.forEach((map) => {
                const mapName = map as MapName
                newState[mapName] = {
                  isBanned: false,
                  isPicked: false
                }
              })
            }
            mapsState.value = newState
            
            // Инициализируем состояние из сессии (должно быть пустым после сброса)
            initializeMapsState()
            
            const stateAfter = {
              started: vetoSession.state.value.started,
              finished: vetoSession.state.value.finished,
              currentTeam: vetoSession.state.value.currentTeam,
              bans: [...(vetoSession.state.value.bans || [])],
              selectedMap: vetoSession.state.value.selectedMap,
              mapsStateCount: Object.keys(mapsState.value).length
            }
            
            console.log('✅ [VETO:RESET] Session reset, all maps cleared:', {
              stateBefore,
              stateAfter,
              uiStateBefore,
              uiStateAfter: {
                showFinalOverlay: showFinalOverlay.value,
                showSideOverlay: showSideOverlay.value,
                optimisticBannedMap: optimisticBannedMap.value,
                mapsGridKey: mapsGridKey.value
              },
              mapsState: Object.fromEntries(Object.entries(mapsState.value)),
              allMapsCount: allMaps.value.length,
              changed: JSON.stringify(stateBefore) !== JSON.stringify(stateAfter)
            })
          } else {
            console.error('❌ [VETO:RESET] updateSessionFromWebSocket failed')
          }
        } else {
          console.warn('⚠️ [VETO:RESET] WebSocket message missing map_pool:', {
            hasMapPool: !!sessionData.map_pool,
            hasSession: !!sessionData
          })
        }
      } else {
        console.warn('⚠️ [VETO:RESET] WebSocket message missing session data')
      }
      break
    case 'room:state':
      console.log('🏠 [ROOM:STATE] Processing room state message:', {
        hasVetoSession: !!message.data?.veto_session,
        vetoSessionId: message.data?.veto_session?.id,
        hasRoom: !!message.data?.room,
        timestamp: messageTimestamp
      })
      
      if (message.data?.veto_session) {
        const sessionData = message.data.veto_session
        
        console.log('📋 [ROOM:STATE] Veto session data:', {
          sessionId: sessionData.id,
          status: sessionData.status,
          hasMapPool: !!sessionData.map_pool,
          hasActions: sessionData.actions !== undefined,
          actionsCount: sessionData.actions?.length || 0,
          mapsCount: sessionData.map_pool?.maps?.length || 0
        })
        
        if (sessionData && sessionData.map_pool && sessionData.actions !== undefined) {
          const success = vetoSession.updateSessionFromWebSocket(sessionData)
          
          console.log('📊 [ROOM:STATE] updateSessionFromWebSocket result:', {
            success,
            newStarted: vetoSession.state.value.started,
            newFinished: vetoSession.state.value.finished,
            newCurrentTeam: vetoSession.state.value.currentTeam,
            newBans: vetoSession.state.value.bans
          })
          
          if (success) {
            updateTeamNamesFromSession()
            initializeMapsState()
            
            const stateAfter = {
              started: vetoSession.state.value.started,
              finished: vetoSession.state.value.finished,
              currentTeam: vetoSession.state.value.currentTeam,
              bans: [...(vetoSession.state.value.bans || [])],
              selectedMap: vetoSession.state.value.selectedMap,
              mapsStateCount: Object.keys(mapsState.value).length
            }
            
            console.log('✅ [ROOM:STATE] Session updated from room state:', {
              stateBefore,
              stateAfter,
              changed: JSON.stringify(stateBefore) !== JSON.stringify(stateAfter)
            })
          } else {
            console.error('❌ [ROOM:STATE] updateSessionFromWebSocket failed')
          }
        } else {
          console.warn('⚠️ [ROOM:STATE] Veto session data invalid:', {
            hasSession: !!sessionData,
            hasMapPool: !!sessionData?.map_pool,
            hasActions: sessionData?.actions !== undefined
          })
        }
      } else {
        console.log('ℹ️ [ROOM:STATE] No veto session in room state')
      }
      break
    case 'error':
      console.error('❌ [ERROR] WebSocket error message received:', {
        message: message.data?.message || 'Unknown error',
        fullData: message.data,
        timestamp: messageTimestamp
      })
      
      // Обрабатываем различные типы ошибок
      const errorMessage = message.data?.message || 'Unknown error'
      
      // Игнорируем ошибку "session is already started" - это нормально, если REST API уже стартовал сессию
      if (errorMessage.includes('session is already started')) {
        console.log('ℹ️ [ERROR] Ignoring "session is already started" error - session was already started via REST API')
        // Не показываем ошибку пользователю, так как это нормальная ситуация
      } else {
        // Для других ошибок показываем уведомление пользователю
        showErrorToast({
          code: 'WEBSOCKET_ERROR',
          message: errorMessage
        } as any)
      }
      break
      
    default:
      console.warn('⚠️ [handleWebSocketMessage] Unknown message type:', {
        messageType: message.type,
        hasData: !!message.data,
        timestamp: messageTimestamp
      })
      break
  }
}

watch(() => vetoSession.logEntries.value.length, async () => {
  await nextTick()
  const logElement = document.querySelector('.log')
  if (logElement) {
    logElement.scrollTop = logElement.scrollHeight
  }
})

watch(() => vetoSession.state.value.finished, finished => {
  if (finished && pickedMap.value) {
    showFinalOverlay.value = true
  }
})

async function handleStart() {
  const currentRoomId = actualRoomId.value
  console.log('▶️ [START] handleStart called:', {
    sessionId: vetoSession.sessionId.value,
    started: vetoSession.state.value.started,
    finished: vetoSession.state.value.finished,
    roomIdFromQuery,
    actualRoomId: currentRoomId,
    roomValueId: room.value?.id,
    hasWebSocket: !!roomWs.value,
    isConnected: roomWs.value?.isConnected.value,
    hasRoom: !!room.value,
    timestamp: new Date().toISOString()
  })
  
  if (vetoSession.sessionId.value && !vetoSession.state.value.started) {
    loading.value = true
    try {
      console.log('📤 [START] Calling startSession API:', {
        sessionId: vetoSession.sessionId.value,
        timestamp: new Date().toISOString()
      })
      
      const updatedSession = await vetoService.startSession(vetoSession.sessionId.value)
      
      console.log('✅ [START] startSession API response:', {
        sessionId: updatedSession.id,
        status: updatedSession.status,
        started: updatedSession.status === 'in_progress' || updatedSession.status === 'finished',
        timestamp: new Date().toISOString()
      })
      
      vetoSession.session.value = updatedSession
      
      // Проверяем, подключен ли WebSocket после старта
      const currentRoomIdAfterStart = actualRoomId.value
      console.log('🔍 [START] Checking WebSocket connection after start:', {
        sessionId: vetoSession.sessionId.value,
        roomIdFromQuery,
        actualRoomId: currentRoomIdAfterStart,
        roomValueId: room.value?.id,
        hasWebSocket: !!roomWs.value,
        isConnected: roomWs.value?.isConnected.value,
        hasRoom: !!room.value
      })
      
      // Отправляем WebSocket сообщение для синхронизации с другими пользователями
      if (roomWs.value && currentRoomIdAfterStart && roomWs.value.isConnected.value) {
        try {
          console.log('📤 [START] Sending veto:start via WebSocket for synchronization:', {
            sessionId: vetoSession.sessionId.value,
            roomId: currentRoomIdAfterStart,
            timestamp: new Date().toISOString()
          })
          
          roomWs.value.sendVetoStart()
          
          console.log('✅ [START] veto:start sent via WebSocket')
        } catch (err: any) {
          console.error('❌ [START] Error sending veto:start via WebSocket:', {
            error: err,
            message: err.message,
            sessionId: vetoSession.sessionId.value,
            roomId: currentRoomIdAfterStart
          })
          // Не показываем ошибку пользователю, так как REST API уже отработал успешно
        }
      } else {
        console.log('ℹ️ [START] WebSocket not available for sending veto:start:', {
          hasWebSocket: !!roomWs.value,
          hasRoomId: !!currentRoomIdAfterStart,
          isConnected: roomWs.value?.isConnected.value
        })
      }
      
      // Если WebSocket не подключен, пытаемся подключить
      if (!roomWs.value && !currentRoomIdAfterStart) {
        console.log('🔍 [START] No WebSocket and no roomId, trying to get room by sessionId:', {
          sessionId: vetoSession.sessionId.value
        })
        
        try {
          const roomData = await getRoomBySessionId(vetoSession.sessionId.value)
          
          if (roomData) {
            console.log('✅ [START] Room found after start:', {
              sessionId: vetoSession.sessionId.value,
              roomId: roomData.id,
              timestamp: new Date().toISOString()
            })
            
            // Используем функцию roomResponseToRoom для правильного маппинга participants
            room.value = roomResponseToRoom(roomData)
            
            if (authStore.isAuthenticated) {
              console.log('🔌 [START] Connecting WebSocket after finding room:', {
                sessionId: vetoSession.sessionId.value,
                roomId: room.value.id
              })
              
              roomWs.value = useRoomWebSocket(room.value.id)
              roomWs.value.connect()
              
              console.log('✅ [START] WebSocket connected after start:', {
                sessionId: vetoSession.sessionId.value,
                roomId: room.value.id,
                isConnected: roomWs.value?.isConnected.value
              })
            }
          }
        } catch (err: any) {
          console.error('❌ [START] Error getting room after start:', {
            sessionId: vetoSession.sessionId.value,
            error: err,
            message: err.message
          })
        }
      }
    } catch (err: any) {
      console.error('❌ [START] Error starting session:', {
        sessionId: vetoSession.sessionId.value,
        error: err,
        message: err.message,
        timestamp: new Date().toISOString()
      })
      showErrorToast(err)
    } finally {
      loading.value = false
    }
    return
  }

  if (!currentPool.value || !poolId) {
    error.value = 'Пул карт не загружен'
    return
  }

  loading.value = true
  try {
    const success = await vetoSession.createSession(
      poolId,
      currentPool.value.gameId,
      vetoType.value,
      teamAName.value,
      teamBName.value,
      60
    )

    if (success && vetoSession.sessionId.value) {
      const updatedSession = await vetoService.startSession(vetoSession.sessionId.value)
      vetoSession.session.value = updatedSession
      initializeMapsState()
      
      // Отправляем WebSocket сообщение для синхронизации с другими пользователями
      const currentRoomIdForNewSession = actualRoomId.value
      if (roomWs.value && currentRoomIdForNewSession && roomWs.value.isConnected.value) {
        try {
          console.log('📤 [START] Sending veto:start via WebSocket for new session:', {
            sessionId: vetoSession.sessionId.value,
            roomId: currentRoomIdForNewSession,
            timestamp: new Date().toISOString()
          })
          
          roomWs.value.sendVetoStart()
          
          console.log('✅ [START] veto:start sent via WebSocket for new session')
        } catch (err: any) {
          console.error('❌ [START] Error sending veto:start via WebSocket for new session:', {
            error: err,
            message: err.message,
            sessionId: vetoSession.sessionId.value,
            roomId: currentRoomIdForNewSession
          })
        }
      }
    } else {
      const errorMsg = vetoSession.error.value || 'Не удалось создать сессию'
      showErrorToast({ code: '', message: errorMsg } as any)
    }
  } catch (err: any) {
    showErrorToast(err)
  } finally {
    loading.value = false
  }
}

async function handleBan(mapName: MapName) {
  const currentRoomIdForBan = actualRoomId.value
  
  console.log('🚫 [BAN] handleBan called:', {
    mapName,
    sessionId: vetoSession.sessionId.value,
    roomIdFromQuery,
    actualRoomId: currentRoomIdForBan,
    roomValueId: room.value?.id,
    hasWebSocket: !!roomWs.value,
    isConnected: roomWs.value?.isConnected.value,
    currentTeam: vetoSession.state.value.currentTeam,
    userTeam: userTeam.value,
    canBan: canBan.value,
    started: vetoSession.state.value.started,
    currentBans: vetoSession.state.value.bans
  })
  
  if (!vetoSession.state.value.started) {
    showErrorToast({ code: '', message: 'Сессия еще не начата. Нажмите "Начать" для начала вето.' } as any)
    return
  }

  if (!canBan.value) {
    const name =
      vetoSession.state.value.currentTeam === 'A'
        ? vetoSession.session.value?.team_a_name
        : vetoSession.session.value?.team_b_name
    showErrorToast({ code: '', message: `Сейчас очередь команды "${name}". Дождитесь своего хода.` } as any)
    return
  }

  if (vetoSession.loading.value) return

  if (!allMaps.value.includes(mapName as any)) {
    showErrorToast({ code: '', message: `Карта "${mapName}" недоступна для бана` } as any)
    return
  }

  const map = vetoSession.session.value?.map_pool?.maps?.find(m => m.name === mapName) ||
              currentPool.value?.maps?.find(m => m.name === mapName)
  
  if (!map) {
    showErrorToast({ code: '', message: `Карта "${mapName}" не найдена` } as any)
    return
  }

  console.log('✅ [BAN] Map found, applying optimistic update:', {
    mapName,
    mapId: map.id,
    currentState: mapsState.value[mapName]
  })

  // Оптимистичное обновление - обновляем только конкретную карту
  optimisticBannedMap.value = mapName
  updateMapState(mapName, { isBanned: true })
  
  console.log('📊 [BAN] After optimistic update:', {
    optimisticBanned: optimisticBannedMap.value,
    mapState: mapsState.value[mapName]
  })

  if (roomWs.value && currentRoomIdForBan && vetoSession.sessionId.value) {
    try {
      console.log('📤 [BAN] Sending ban via WebSocket:', {
        sessionId: vetoSession.sessionId.value,
        mapId: map.id,
        mapName,
        team: vetoSession.state.value.currentTeam,
        roomId: currentRoomIdForBan,
        isConnected: roomWs.value.isConnected.value,
        timestamp: new Date().toISOString()
      })
      
      roomWs.value.sendVetoBan(
        vetoSession.sessionId.value,
        map.id,
        vetoSession.state.value.currentTeam
      )
      
      console.log('✅ [BAN] Ban sent via WebSocket, waiting for response...', {
        sessionId: vetoSession.sessionId.value,
        mapId: map.id,
        mapName,
        optimisticBanned: optimisticBannedMap.value
      })
    } catch (err: any) {
      console.error('❌ [BAN] Error sending ban via WebSocket:', {
        error: err,
        message: err.message,
        stack: err.stack,
        sessionId: vetoSession.sessionId.value,
        mapId: map.id,
        mapName,
        timestamp: new Date().toISOString()
      })
      
      // Откатываем оптимистичное обновление при ошибке
      optimisticBannedMap.value = null
      updateMapState(mapName, { isBanned: false })
      
      console.log('🔄 [BAN] Rolled back optimistic update due to error:', {
        mapName,
        newState: mapsState.value[mapName]
      })
      
      showErrorToast({ code: '', message: err.message || 'Не удалось отправить бан через WebSocket' } as any)
    }
  } else {
    console.log('📤 [BAN] Using REST API fallback (no WebSocket):', {
      hasWebSocket: !!roomWs.value,
      hasRoomId: !!currentRoomIdForBan,
      actualRoomId: currentRoomIdForBan,
      roomIdFromQuery,
      roomValueId: room.value?.id,
      hasSessionId: !!vetoSession.sessionId.value,
      reason: !roomWs.value ? 'No WebSocket' : !currentRoomIdForBan ? 'No roomId' : 'Unknown'
    })
    
    // Fallback на REST API
    const success = await vetoSession.banMap(mapName as any)

    if (!success) {
      console.error('❌ [BAN] REST API ban failed')
      // Откатываем оптимистичное обновление при ошибке
      optimisticBannedMap.value = null
      updateMapState(mapName, { isBanned: false })
      
      if (vetoSession.error.value) {
        showErrorToast({ code: '', message: vetoSession.error.value } as any)
      }
    } else {
      console.log('✅ [BAN] REST API ban successful')
      optimisticBannedMap.value = null
      
      if (vetoSession.state.value.finished) {
        showFinalOverlay.value = true
      }
    }
  }
}

function handleSwap() {
  if (!roomWs.value || !actualRoomId.value) {
    showErrorToast({ code: '', message: 'WebSocket не подключен. Смена хода недоступна.' } as any)
    return
  }
  console.warn('Swap functionality not implemented via WebSocket yet')
}

async function handleReset() {
  const currentRoomIdForReset = actualRoomId.value
  
  console.log('🔄 [RESET] handleReset called:', {
    sessionId: vetoSession.sessionId.value,
    currentBans: vetoSession.state.value.bans,
    pickedMap: vetoSession.state.value.selectedMap,
    mapsState: Object.fromEntries(Object.entries(mapsState.value)),
    roomIdFromQuery,
    actualRoomId: currentRoomIdForReset,
    roomValueId: room.value?.id,
    hasWebSocket: !!roomWs.value,
    isConnected: roomWs.value?.isConnected.value
  })
  
  const success = await vetoSession.resetSession()
  if (success) {
    console.log('✅ [RESET] Session reset successful, clearing UI state')
    
    // Отправляем WebSocket сообщение для синхронизации с другими пользователями
    if (roomWs.value && currentRoomIdForReset && roomWs.value.isConnected.value) {
      try {
        console.log('📤 [RESET] Sending veto:reset via WebSocket for synchronization:', {
          sessionId: vetoSession.sessionId.value,
          roomId: currentRoomIdForReset,
          timestamp: new Date().toISOString()
        })
        
        roomWs.value.sendVetoReset()
        
        console.log('✅ [RESET] veto:reset sent via WebSocket')
      } catch (err: any) {
        console.error('❌ [RESET] Error sending veto:reset via WebSocket:', {
          error: err,
          message: err.message,
          sessionId: vetoSession.sessionId.value,
          roomId: currentRoomIdForReset
        })
        // Не показываем ошибку пользователю, так как REST API уже отработал успешно
      }
    } else {
      console.log('ℹ️ [RESET] WebSocket not available for sending veto:reset:', {
        hasWebSocket: !!roomWs.value,
        hasRoomId: !!currentRoomIdForReset,
        isConnected: roomWs.value?.isConnected.value
      })
    }
    
    showFinalOverlay.value = false
    showSideOverlay.value = false
    optimisticBannedMap.value = null
    
    // Принудительно перерендериваем все карты при сбросе
    mapsGridKey.value++
    
    // Полностью очищаем состояние всех карт
    const newState: Record<MapName, { isBanned: boolean; isPicked: boolean }> = {} as Record<MapName, { isBanned: boolean; isPicked: boolean }>
    if (allMaps.value.length > 0) {
      allMaps.value.forEach((map) => {
        const mapName = map as MapName
        newState[mapName] = {
          isBanned: false,
          isPicked: false
        }
      })
    }
    mapsState.value = newState
    
    // Инициализируем состояние из сессии (должно быть пустым после сброса)
    initializeMapsState()
    
    console.log('📊 [RESET] After reset:', {
      bannedMaps: vetoSession.state.value.bans,
      pickedMap: vetoSession.state.value.selectedMap,
      mapsState: Object.fromEntries(Object.entries(mapsState.value)),
      mapsGridKey: mapsGridKey.value,
      allMapsCount: allMaps.value.length
    })
  } else {
    console.error('❌ [RESET] Session reset failed')
  }
}

function handleSide() {
  if (!vetoSession.state.value.finished || !pickedMap.value) {
    alert('Сначала завершите вето и выберите карту.')
    return
  }
  showSideOverlay.value = true
}

</script>

<template>
  <div class="container" style="position: relative; z-index: 1;">
    <div v-if="loading || vetoSession.loading.value" class="loading-message">
      {{ shareToken ? 'Загрузка сессии...' : 'Загрузка пула карт...' }}
    </div>

    <div v-else-if="error" class="error-message">
      {{ error }}
      <button @click="router.push('/ban/valorant')" class="btn btn-primary">
        Вернуться к выбору пула
      </button>
    </div>

    <template v-else-if="vetoSession.session.value || currentPool">
      <VetoHeader
        :current-team="vetoSession.state.value.currentTeam"
        :team-a-name="teamAName"
        :team-b-name="teamBName"
        :started="vetoSession.state.value.started"
        :finished="vetoSession.state.value.finished"
        @start="handleStart"
        @swap="handleSwap"
        @reset="handleReset"
        @side="handleSide"
      />

      <main>
        <section class="panel">
          <div class="panel-header">
            <div class="panel-title">All maps</div>
            <div class="current-step">
              Шаг:
              <span :class="['pill', vetoSession.state.value.finished ? 'done' : 'step']">
                <template v-if="!vetoSession.state.value.started">
                  Нажмите «Начать», чтобы начать вето
                </template>
                <template v-else-if="vetoSession.state.value.finished">
                  Veto завершён
                </template>
                <template v-else>
                  Ход бана: {{ currentTeamName }}
                </template>
              </span>
            </div>
          </div>
          <MapsGrid
            :key="mapsGridKey"
            :all-maps="allMaps"
            :picked-map="pickedMap"
            :finished="vetoSession.state.value.finished"
            :started="vetoSession.state.value.started"
            :can-ban="canBan"
            :is-map-banned="isMapBanned"
            :is-map-picked="isMapPicked"
            @ban="handleBan"
          />
        </section>

        <SummaryPanel
          :picked-map="pickedMap"
          :log-entries="vetoSession.logEntries.value"
        />
      </main>
    </template>
  </div>

  <FinalOverlay
    :show="showFinalOverlay"
    :map-name="pickedMap"
    @close="showFinalOverlay = false"
  />

  <SideOverlay
    :show="showSideOverlay"
    :team-a-name="teamAName"
    :team-b-name="teamBName"
    @close="showSideOverlay = false"
  />
</template>

<style scoped>
.loading-message,
.error-message {
  text-align: center;
  color: white;
  padding: 2rem;
  font-size: 1.1rem;
}

.error-message {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}
</style>
