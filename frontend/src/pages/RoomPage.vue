  <script setup lang="ts">
  import { ref, computed, onMounted, onUnmounted, watch, shallowRef } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { useI18n } from '../composables/useI18n'
  import { useAuthStore } from '../store/auth'
  import RoomPasswordModal from '../components/rooms/RoomPasswordModal.vue'
  import type { Room } from '../types'
  import { Users, Copy, Play, ArrowLeft, Settings } from 'lucide-vue-next'
  import * as roomApi from '../services/api/roomService'
  import * as vetoService from '../services/api/vetoService'
  import type { ApiError } from '../services/api/types'
  import { useRoomWebSocket } from '../composables/useRoomWebSocket'
  import { getAllPools } from '../services/mapPoolService'
  import type { MapPool } from '../types'
  import Select from 'primevue/select'

  interface Props {
    roomId?: string
  }

  const props = defineProps<Props>()
  const route = useRoute()
  const router = useRouter()
  const { t } = useI18n()
  const authStore = useAuthStore()

  const roomId = computed(() => props.roomId || route.params.roomId as string)
  const room = ref<Room | null>(null)

  const isLoading = ref(true)
  const error = ref<string | null>(null)

  const showCodeCopied = ref(false)
  const showPasswordModal = ref(false)
  const pendingJoin = ref(false)
  const showSettingsModal = ref(false)

  const hasRedirectedToVeto = ref(false)
  
  // Состояние для редактирования настроек
  const availablePools = ref<MapPool[]>([])
  const editingVetoType = ref<'bo1' | 'bo3' | 'bo5' | null>(null)
  const editingMapPoolId = ref<number | null>(null)
  const isSavingSettings = ref(false)

  // Опции для PrimeVue Select
  const vetoTypeOptions = computed(() => [
    { label: t('bestOf.bo1'), value: 'bo1' },
    { label: t('bestOf.bo3'), value: 'bo3' },
    { label: t('bestOf.bo5'), value: 'bo5' }
  ])

  const mapPoolOptionsForSettings = computed(() => [
    { label: t('rooms.settings.selectPoolPlaceholder'), value: null },
    ...availablePools.value.map(pool => ({
      label: pool.name,
      value: pool.id
    }))
  ])

  const roomWs = shallowRef<ReturnType<typeof useRoomWebSocket> | null>(null)
  const lastProcessedMessageIndex = ref(-1)

  const isOwner = computed(() =>
    authStore.isAuthenticated &&
    authStore.user &&
    room.value?.ownerId === authStore.user.id
  )

  const isParticipant = computed(() => {
    if (!authStore.isAuthenticated || !authStore.user || !room.value) return false
    return room.value.participants?.some(p => p.userId === authStore.user!.id) || false
  })

  const canJoin = computed(() => {
    if (!room.value) return false
    const count = room.value.participants?.length || 0
    return room.value.status === 'waiting' && count < 2 && !isParticipant.value
  })

  const participantsCount = computed(() => room.value?.participants?.length || 0)
  const isFull = computed(() => participantsCount.value >= 2)

  const teamAParticipant = computed(() => room.value?.participants?.[0] || null)
  const teamBParticipant = computed(() => room.value?.participants?.[1] || null)

  const hasActiveVetoSession = computed(() =>
    room.value?.status === 'active' && !!room.value?.vetoSessionId
  )

  onMounted(() => {
    if (!authStore.isAuthenticated) {
      router.push({ path: '/login', query: { redirect: route.fullPath } })
      return
    }
    loadRoom()
  })

  watch(
    () => ({
      status: room.value?.status,
      vetoSessionId: room.value?.vetoSessionId,
    }),
    (curr, prev) => {
      // Убираем автоматический редирект при загрузке страницы
      // Редирект только при изменении статуса с waiting на active
      if (
        !prev ||
        !prev.status ||
        hasRedirectedToVeto.value ||
        !isParticipant.value
      ) return

      // Редирект только если статус изменился с waiting на active
      if (
        prev.status === 'waiting' &&
        curr.status === 'active' &&
        curr.vetoSessionId
      ) {
        const poolId = room.value?.mapPoolId
        if (poolId) {
          hasRedirectedToVeto.value = true
          router.push(`/veto/valorant/${poolId}?session=${curr.vetoSessionId}`)
        }
      }
    }
  )

  watch([room, isParticipant], ([newRoom, newIsParticipant]) => {
    // Проверяем, нужно ли создавать новое соединение
    const needsConnection = newRoom && newIsParticipant
    const currentRoomId = roomWs.value && newRoom ? Number(newRoom.id) : null
    const newRoomId = needsConnection ? Number(newRoom.id) : null
    
    // Если комната не изменилась и соединение уже есть, ничего не делаем
    if (needsConnection && currentRoomId === newRoomId && roomWs.value?.isConnected.value) {
      return
    }
    
    if (needsConnection) {
      // Отключаем старое соединение, если оно есть и комната изменилась
      if (roomWs.value && currentRoomId !== newRoomId) {
        roomWs.value.disconnect()
        roomWs.value = null
      }
      
      // Создаем новое соединение только если его нет или комната изменилась
      if (!roomWs.value || currentRoomId !== newRoomId) {
        roomWs.value = useRoomWebSocket(Number(newRoom.id))
        roomWs.value.connect()
      }
    } else {
      if (roomWs.value) {
        roomWs.value.disconnect()
        roomWs.value = null
      }
    }
  }, { immediate: true })

  // Обрабатываем сообщения WebSocket (только новые)
  watch(
    () => roomWs.value?.messages.value,
    (messages) => {
      if (!messages || messages.length === 0) return
      
      // Обрабатываем только новые сообщения
      const startIndex = lastProcessedMessageIndex.value + 1
      const newMessages = messages.slice(startIndex)
      
      for (const message of newMessages) {
        handleWebSocketMessage(message)
      }
      
      lastProcessedMessageIndex.value = messages.length - 1
    },
    { deep: true }
  )
  
  // Сбрасываем индекс при смене соединения
  watch(roomWs, () => {
    lastProcessedMessageIndex.value = -1
  })

  onUnmounted(() => {
    roomWs.value?.disconnect()
    roomWs.value = null
  })

  const loadRoom = async () => {
    isLoading.value = true
    error.value = null

    try {
      const response = await roomApi.getRoom(Number(roomId.value))
      room.value = roomApi.roomResponseToRoom(response)
    } catch (err) {
      const apiError = err as ApiError
      error.value =
        apiError.code === 'HTTP_404'
          ? t('errors.roomNotFound')
          : apiError.message || t('errors.roomLoadError')
    } finally {
      isLoading.value = false
    }
  }

  const copyRoomCode = async () => {
    if (!room.value) return
    await navigator.clipboard.writeText(room.value.code)
    showCodeCopied.value = true
    setTimeout(() => (showCodeCopied.value = false), 2000)
  }

  const handleJoin = async () => {
    if (!canJoin.value || !room.value) return

    if (room.value.type === 'private' && room.value.password) {
      showPasswordModal.value = true
      pendingJoin.value = true
      return
    }

    await performJoin()
  }

  const handlePasswordSubmit = async (password: string) => {
    if (!room.value) return

    isLoading.value = true
    error.value = null

    try {
      const response = await roomApi.joinRoom(Number(roomId.value), password)
      room.value = roomApi.roomResponseToRoom(response)
    } catch (err) {
      const apiError = err as ApiError
      error.value = apiError.message || t('rooms.invalidPassword')
      setTimeout(() => (error.value = null), 3000)
    } finally {
      showPasswordModal.value = false
      pendingJoin.value = false
      isLoading.value = false
    }
  }

  const performJoin = async () => {
    if (!room.value) return

    isLoading.value = true
    error.value = null

    try {
      const response = await roomApi.joinRoom(Number(roomId.value))
      room.value = roomApi.roomResponseToRoom(response)
    } catch (err) {
      const apiError = err as ApiError
      error.value = apiError.message || t('errors.roomJoinError')
      setTimeout(() => (error.value = null), 3000)
    } finally {
      pendingJoin.value = false
      isLoading.value = false
    }
  }

  const handleStartVeto = async () => {
    if (!room.value || !isOwner.value || !room.value.mapPoolId) return

    isLoading.value = true
    error.value = null

    try {
      // Создаем сессию вето с никнеймами игроков
      const teamAName = teamAParticipant.value?.username || `Team ${teamAParticipant.value?.userId || 'A'}`
      const teamBName = teamBParticipant.value?.username || `Team ${teamBParticipant.value?.userId || 'B'}`
      
      // Используем veto_type из комнаты, или 'bo1' по умолчанию
      const vetoType = room.value.vetoType || 'bo1'
      
      const session = await vetoService.createSession({
        game_id: room.value.gameId,
        map_pool_id: room.value.mapPoolId,
        type: vetoType,
        team_a_name: teamAName,
        team_b_name: teamBName,
        timer_seconds: 60,
      })

      // Активируем сессию (меняем статус с not_started на in_progress)
      await vetoService.startSession(session.id)

      // Обновляем комнату
      await roomApi.updateRoom(Number(roomId.value), {
        veto_session_id: session.id,
        status: 'active',
      })

      // Отправляем WebSocket сообщение для уведомления других участников
      roomWs.value?.send({
        type: 'veto:start',
        data: {
          session_id: session.id,
          room_id: Number(roomId.value),
          map_pool_id: room.value.mapPoolId,
        },
      })

      // Редирект на страницу вето
      hasRedirectedToVeto.value = true
      router.push(`/veto/valorant/${room.value.mapPoolId}?session=${session.id}`)
    } catch (err) {
      const apiError = err as ApiError
      error.value = apiError.message || 'Не удалось начать вето'
      setTimeout(() => (error.value = null), 3000)
    } finally {
      isLoading.value = false
    }
  }

  const handleLeave = async () => {
    if (!isParticipant.value) return

    isLoading.value = true
    error.value = null

    try {
      roomWs.value?.disconnect()
      roomWs.value = null

      await roomApi.leaveRoom(Number(roomId.value))
      // Всегда редиректим на список комнат после выхода
      router.push('/rooms')
    } catch (err) {
      const apiError = err as ApiError
      error.value = apiError.message || t('errors.leaveRoomError')
      setTimeout(() => (error.value = null), 3000)
      // В случае ошибки тоже редиректим (пользователь больше не участник)
      router.push('/rooms')
    } finally {
      isLoading.value = false
    }
  }

  const canEditSettings = computed(() => {
    // Можно редактировать настройки если:
    // 1. Пользователь - владелец комнаты
    // 2. Комната в статусе waiting ИЛИ сессия не начата (not_started)
    if (!isOwner.value || !room.value) return false
    
    // Если сессии нет - можно редактировать
    if (!room.value.vetoSessionId) return true
    
    // Если сессия есть, нужно проверить её статус (будет реализовано позже через загрузку сессии)
    // Пока разрешаем редактирование если статус комнаты waiting
    return room.value.status === 'waiting'
  })
  
  const handleOpenSettings = async () => {
    if (!room.value) return
    
    // Загружаем доступные пулы карт
    try {
      const pools = await getAllPools()
      availablePools.value = pools
    } catch (err) {
      console.error('Error loading pools:', err)
      error.value = t('errors.poolsLoadError')
      return
    }
    
    // Устанавливаем текущие значения
    editingVetoType.value = room.value.vetoType || 'bo1'
    editingMapPoolId.value = room.value.mapPoolId || null
    
    showSettingsModal.value = true
  }
  
  const handleSaveSettings = async () => {
    if (!room.value || !roomId.value || isSavingSettings.value) return
    
    isSavingSettings.value = true
    error.value = null
    
    try {
      await roomApi.updateRoom(Number(roomId.value), {
        veto_type: editingVetoType.value || undefined,
        map_pool_id: editingMapPoolId.value || undefined,
      })
      
      // Обновляем комнату после сохранения
      await loadRoom()
      showSettingsModal.value = false
    } catch (err) {
      const apiError = err as ApiError
      error.value = apiError.message || t('errors.saveSettingsError')
      setTimeout(() => (error.value = null), 3000)
    } finally {
      isSavingSettings.value = false
    }
  }
  
  const handleGoToVeto = async () => {
    if (!room.value?.vetoSessionId) return

    let poolId = room.value.mapPoolId
    if (!poolId) {
      const session = await vetoService.getSession(room.value.vetoSessionId)
      poolId = session.map_pool_id
    }

    if (poolId) {
      router.push(`/veto/valorant/${poolId}?session=${room.value.vetoSessionId}`)
    }
  }

  const handleWebSocketMessage = (message: any) => {
    switch (message.type) {
      case 'room:join':
      case 'room:leave':
      case 'room:participants:updated':
        loadRoom()
        break
      case 'room:state':
        // Обновляем veto_type и map_pool_id из WebSocket сообщения
        if (message.data && room.value) {
          if (message.data.veto_type !== undefined) {
            room.value.vetoType = message.data.veto_type || undefined
          }
          if (message.data.map_pool_id !== undefined) {
            room.value.mapPoolId = message.data.map_pool_id || undefined
          }
        }
        loadRoom() // Перезагружаем для получения полных данных
        break
      case 'room:deleted':
        router.push('/rooms')
        break
    }
  }
  </script>
    

  <template>
    <div class="room-page">
      <div class="page-container">
        <button class="back-btn" @click="router.push('/rooms')">
          <ArrowLeft :size="18" />
          {{ t('common.back') }}
        </button>

        <div v-if="isLoading" class="loading-state">
          <p>{{ t('common.loading') }}</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p>{{ error }}</p>
          <button class="btn btn-primary" @click="loadRoom">
            {{ t('common.retry') }}
          </button>
        </div>

        <div v-else-if="room" class="room-content">
          <!-- Room Header -->
          <div class="room-header">
            <div class="room-info">
              <h1 class="room-name">{{ room.name }}</h1>
              <div class="room-code-section">
                <span class="room-code-label">{{ t('rooms.roomCode') }}:</span>
                <div class="room-code-box">
                  <span class="room-code-value">{{ room.code }}</span>
                  <button class="copy-btn" @click="copyRoomCode" :title="t('rooms.copyCode')">
                    <Copy :size="16" />
                  </button>
                  <span v-if="showCodeCopied" class="copied-message">{{ t('rooms.codeCopied') }}</span>
                </div>
              </div>
            </div>

            <div class="room-status-badge" :class="`status-${room.status}`">
              <Play v-if="room.status === 'active'" :size="16" />
              <span>{{ t(`rooms.status.${room.status}`) }}</span>
            </div>
          </div>

          <!-- Participants Section -->
          <div class="participants-section">
            <h2 class="section-title">
              <Users :size="20" />
              {{ t('rooms.participants') }} ({{ participantsCount }} / 2)
            </h2>

            <div class="teams-grid">
              <!-- Team A -->
              <div class="team-card team-a">
                <h3 class="team-title">{{ t('veto.teamA') }}</h3>
                <div v-if="teamAParticipant" class="participant-card">
                  <div class="participant-info">
                    <span class="participant-name">{{ teamAParticipant.username || `User #${teamAParticipant.userId}` }}</span>
                    <span v-if="teamAParticipant.role === 'owner'" class="owner-badge">
                      {{ t('rooms.owner') }}
                    </span>
                  </div>
                </div>
                <div v-else class="empty-slot">
                  <span>{{ t('rooms.waitingForPlayer') }}</span>
                </div>
              </div>

              <!-- Team B -->
              <div class="team-card team-b">
                <h3 class="team-title">{{ t('veto.teamB') }}</h3>
                <div v-if="teamBParticipant" class="participant-card">
                  <div class="participant-info">
                    <span class="participant-name">{{ teamBParticipant.username || `User #${teamBParticipant.userId}` }}</span>
                    <span v-if="teamBParticipant.role === 'owner'" class="owner-badge">
                      {{ t('rooms.owner') }}
                    </span>
                  </div>
                </div>
                <div v-else class="empty-slot">
                  <span>{{ t('rooms.waitingForPlayer') }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Actions Section -->
          <div class="actions-section">
            <div v-if="canJoin" class="action-buttons">
              <button class="btn btn-primary btn-large" @click="handleJoin">
                <Users :size="18" />
                {{ t('rooms.joinAsTeamB') }}
              </button>
            </div>

            <div v-else-if="isParticipant" class="action-buttons">
              <button
                v-if="canEditSettings"
                class="btn btn-secondary"
                @click="handleOpenSettings"
              >
                <Settings :size="18" />
                Настройки
              </button>
              <button
                v-if="isOwner && isFull && room.status === 'waiting'"
                class="btn btn-primary btn-large"
                @click="handleStartVeto"
              >
                <Play :size="18" />
                {{ t('rooms.startVeto') }}
              </button>
              <button
                v-if="hasActiveVetoSession"
                class="btn btn-primary btn-large"
                @click="handleGoToVeto"
              >
                <Play :size="18" />
                {{ t('rooms.goToVeto') }}
              </button>
              <button class="btn btn-secondary" @click="handleLeave">
                {{ t('rooms.leave') }}
              </button>
            </div>

            <div v-else-if="isFull" class="info-message">
              <p>{{ t('rooms.roomFull') }}</p>
            </div>
          </div>
        </div>

        <!-- Password Modal -->
        <RoomPasswordModal
          v-if="room"
          :show="showPasswordModal"
          :room-name="room.name"
          @close="showPasswordModal = false; pendingJoin = false"
          @submit="handlePasswordSubmit"
        />
        
        <!-- Settings Modal -->
        <div v-if="showSettingsModal" class="modal-overlay" @click.self="showSettingsModal = false">
          <div class="modal-content">
            <div class="modal-header">
              <h2>Настройки комнаты</h2>
              <button class="modal-close" @click="showSettingsModal = false">×</button>
            </div>
            
            <div class="modal-body">
              <div class="form-group">
                <label class="form-label">{{ t('rooms.settings.vetoTypeLabel') }}</label>
                <Select
                  v-model="editingVetoType"
                  :options="vetoTypeOptions"
                  optionLabel="label"
                  optionValue="value"
                  :placeholder="t('rooms.settings.vetoTypePlaceholder')"
                  fluid
                />
              </div>
              
              <div class="form-group">
                <label class="form-label">{{ t('rooms.settings.mapPoolLabel') }}</label>
                <Select
                  v-model="editingMapPoolId"
                  :options="mapPoolOptionsForSettings"
                  optionLabel="label"
                  optionValue="value"
                  :placeholder="t('rooms.settings.mapPoolPlaceholder')"
                  fluid
                />
              </div>
              
              <div v-if="error" class="error-message">{{ error }}</div>
            </div>
            
            <div class="modal-footer">
              <button class="btn btn-secondary" @click="showSettingsModal = false" :disabled="isSavingSettings">
                {{ t('common.cancel') }}
              </button>
              <button class="btn btn-primary" @click="handleSaveSettings" :disabled="isSavingSettings">
                {{ isSavingSettings ? t('rooms.settings.saving') : t('common.save') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </template>

  <style scoped>
  .room-page {
    min-height: calc(100vh - 200px);
    padding: 2rem;
    position: relative;
    z-index: 1;
    width: 100%;
  }

  .page-container {
    max-width: 1200px;
    margin: 0 auto;
    color: white;
  }

  .back-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    cursor: pointer;
    margin-bottom: 2rem;
    transition: all 0.2s;
  }

  .back-btn:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .loading-state,
  .error-state {
    text-align: center;
    padding: 3rem;
    color: rgba(255, 255, 255, 0.7);
  }

  .room-content {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .room-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 2rem;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 2rem;
  }

  .room-info {
    flex: 1;
  }

  .room-name {
    font-size: 2rem;
    font-weight: 700;
    margin: 0 0 1rem 0;
    color: white;
  }

  .room-code-section {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .room-code-label {
    color: rgba(255, 255, 255, 0.7);
    font-size: 0.9rem;
  }

  .room-code-box {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(0, 0, 0, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 6px;
    padding: 0.5rem 1rem;
  }

  .room-code-value {
    font-family: monospace;
    font-size: 1.1rem;
    font-weight: 600;
    color: white;
    letter-spacing: 0.1em;
  }

  .copy-btn {
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.7);
    cursor: pointer;
    padding: 0.25rem;
    display: flex;
    align-items: center;
    transition: color 0.2s;
  }

  .copy-btn:hover {
    color: white;
  }

  .copied-message {
    font-size: 0.85rem;
    color: #4caf50;
    margin-left: 0.5rem;
  }

  .room-status-badge {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.9rem;
  }

  .room-status-badge.status-waiting {
    background: rgba(255, 193, 7, 0.2);
    color: #ffc107;
  }

  .room-status-badge.status-active {
    background: rgba(76, 175, 80, 0.2);
    color: #4caf50;
  }

  .room-status-badge.status-finished {
    background: rgba(158, 158, 158, 0.2);
    color: #9e9e9e;
  }

  .participants-section {
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 2rem;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0 0 1.5rem 0;
    color: white;
  }

  .teams-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
  }

  .team-card {
    background: rgba(0, 0, 0, 0.9);
    border: 2px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 1.5rem;
  }

  .team-card.team-a {
    border-color: rgba(33, 150, 243, 0.5);
  }

  .team-card.team-b {
    border-color: rgba(244, 67, 54, 0.5);
  }

  .team-title {
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0 0 1rem 0;
    color: white;
  }

  .participant-card {
    background: rgba(255, 255, 255, 0.05);
    border-radius: 6px;
    padding: 1rem;
  }

  .participant-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .participant-name {
    color: white;
    font-weight: 500;
  }

  .owner-badge {
    background: rgba(76, 175, 80, 0.2);
    color: #4caf50;
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .empty-slot {
    text-align: center;
    padding: 2rem 1rem;
    color: rgba(255, 255, 255, 0.5);
    font-style: italic;
    border: 2px dashed rgba(255, 255, 255, 0.2);
    border-radius: 6px;
  }

  .actions-section {
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 2rem;
  }

  .action-buttons {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
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

  .btn-primary:hover:not(:disabled) {
    opacity: 0.9;
    transform: translateY(-1px);
  }

  .btn-secondary {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  .btn-secondary:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .btn-large {
    padding: 1rem 2rem;
    font-size: 1.1rem;
  }

  .info-message {
    text-align: center;
    color: rgba(255, 255, 255, 0.7);
  }
  
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.75);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }
  
  .modal-content {
    background: rgba(15, 23, 42, 0.95);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    color: white;
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .modal-header h2 {
    margin: 0;
    font-size: 1.5rem;
  }
  
  .modal-close {
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.7);
    font-size: 2rem;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: all 0.2s;
  }
  
  .modal-close:hover {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }
  
  .modal-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
  }
  
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding: 1.5rem;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .form-group {
    margin-bottom: 1.5rem;
  }
  
  .form-label {
    display: block;
    margin-bottom: 0.5rem;
    color: rgba(255, 255, 255, 0.9);
    font-weight: 500;
  }
  
  .form-input {
    width: 100%;
    padding: 0.75rem 1rem;
    background: rgba(0, 0, 0, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 6px;
    color: white;
    font-size: 1rem;
    transition: all 0.2s;
  }
  
  .form-input::placeholder {
    color: rgba(255, 255, 255, 0.5);
  }
  
  .form-input:focus {
    outline: none;
    border-color: #667eea;
    background: rgba(255, 255, 255, 0.15);
  }
  
  .form-input option {
    background: rgba(0, 0, 0, 0.9);
    color: white;
  }
  

  @media (max-width: 768px) {
    .room-header {
      flex-direction: column;
    }

    .teams-grid {
      grid-template-columns: 1fr;
    }

    .action-buttons {
      flex-direction: column;
    }

    .btn-large {
      width: 100%;
    }
  }
  </style>
