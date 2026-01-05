<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from '../composables/useI18n';
import { useAuthStore } from '../store/auth';
import RoomCard from '../components/rooms/RoomCard.vue';
import CreateRoomModal from '../components/rooms/CreateRoomModal.vue';
import RoomPasswordModal from '../components/rooms/RoomPasswordModal.vue';
import type { Room } from '../types';
import { Search, Plus, LogIn, X } from 'lucide-vue-next';
import * as roomApi from '../services/api/roomService';
import type { ApiError } from '../services/api/types';

const router = useRouter();
const { t } = useI18n();
const authStore = useAuthStore();

// State
const rooms = ref<Room[]>([]);
const filteredRooms = ref<Room[]>([]);
const searchQuery = ref('');
const filterType = ref<'all' | 'my' | 'public' | 'private'>('all');
const showCreateModal = ref(false);
const showJoinModal = ref(false);
const showPasswordModal = ref(false);
const selectedRoomForPassword = ref<Room | null>(null);
const joinCode = ref('');
const isLoading = ref(false);
const error = ref<string | null>(null);


onMounted(() => {
  // Проверка авторизации (дополнительная защита, хотя guard должен это обработать)
  if (!authStore.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: '/rooms' } });
    return;
  }
  loadRooms();
});

const loadRooms = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    // Определяем тип комнат для загрузки в зависимости от фильтра
    let roomType: 'public' | 'private' | undefined = undefined;
    if (filterType.value === 'public') {
      roomType = 'public';
    } else if (filterType.value === 'private') {
      roomType = 'private';
    }
    // Для 'all' и 'my' загружаем все комнаты (type = undefined)

    const response = await roomApi.getRooms(20, 0, roomType);
    rooms.value = response.data.map(roomApi.roomResponseToRoom);
    applyFilters();
  } catch (err) {
    const apiError = err as ApiError;
    error.value = apiError.message || t('errors.roomsLoadError');
    console.error('Error loading rooms:', err);
  } finally {
    isLoading.value = false;
  }
};

const applyFilters = () => {
  let filtered = [...rooms.value];

  // Apply search filter
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(room =>
      room.name.toLowerCase().includes(query) ||
      room.code.toLowerCase().includes(query)
    );
  }

  // Apply type filter
  if (filterType.value === 'my') {
    if (authStore.isAuthenticated && authStore.user) {
      filtered = filtered.filter(room => room.ownerId === authStore.user!.id);
    } else {
      filtered = [];
    }
  } else if (filterType.value === 'public') {
    filtered = filtered.filter(room => room.type === 'public');
  } else if (filterType.value === 'private') {
    filtered = filtered.filter(room => room.type === 'private');
  }

  filteredRooms.value = filtered;
};

const handleSearch = () => {
  applyFilters();
};

const handleFilterChange = () => {
  // При изменении фильтра перезагружаем комнаты
  loadRooms();
};

const handleCreateRoom = async (roomData: {
  name: string;
  type: 'public' | 'private';
  mapPoolId?: number;
  bestOf?: 'bo1' | 'bo3' | 'bo5';
  password?: string;
}) => {
  isLoading.value = true;
  error.value = null;

  try {
    // Создаем комнату через API
    const response = await roomApi.createRoom({
      name: roomData.name,
      type: roomData.type,
      game_id: 1, // Valorant
      map_pool_id: roomData.mapPoolId,
      veto_type: roomData.bestOf, // Передаем veto_type из bestOf
      max_participants: 2, // Fixed: 2 participants (one from each team)
      password: roomData.password, // Пароль для приватных комнат
    });

    const newRoom = roomApi.roomResponseToRoom(response);
    
    // Если указан пул карт и BO, создаем VetoSession
    if (roomData.mapPoolId && roomData.bestOf) {
      // TODO: Создать VetoSession через API и связать с комнатой
      // Это можно сделать на странице комнаты или здесь
    }

    // Обновляем список комнат
    await loadRooms();
    showCreateModal.value = false;
    
    // Navigate to the new room
    router.push(`/room/${newRoom.id}`);
  } catch (err) {
    const apiError = err as ApiError;
    error.value = apiError.message || 'Не удалось создать комнату';
    console.error('Error creating room:', err);
  } finally {
    isLoading.value = false;
  }
};

const handleJoinByCode = async () => {
  if (!joinCode.value.trim()) {
    return;
  }

  isLoading.value = true;
  error.value = null;

  try {
    // Пытаемся найти комнату по коду через API
    // Для этого нужно сначала получить список комнат или использовать специальный endpoint
    // Пока используем поиск в загруженных комнатах
    const room = rooms.value.find(r => r.code.toLowerCase() === joinCode.value.trim().toLowerCase());
    
    if (room) {
      // Если комната приватная, показываем popup для ввода пароля
      if (room.type === 'private') {
        selectedRoomForPassword.value = room;
        showPasswordModal.value = true;
        joinCode.value = '';
        showJoinModal.value = false;
      } else {
        // Присоединяемся к публичной комнате
        await roomApi.joinRoom(room.id);
        router.push(`/room/${room.id}`);
        joinCode.value = '';
        showJoinModal.value = false;
      }
    } else {
      error.value = t('rooms.invalidCode');
      setTimeout(() => {
        error.value = null;
      }, 3000);
    }
  } catch (err) {
    const apiError = err as ApiError;
    error.value = apiError.message || t('rooms.invalidCode');
    setTimeout(() => {
      error.value = null;
    }, 3000);
  } finally {
    isLoading.value = false;
  }
};

const handleRoomJoin = async (room: Room) => {
  isLoading.value = true;
  error.value = null;

  try {
    // Показываем popup для ввода пароля приватной комнаты
    if (room.type === 'private') {
      selectedRoomForPassword.value = room;
      showPasswordModal.value = true;
    } else {
      // Присоединяемся к публичной комнате
      await roomApi.joinRoom(room.id);
      try {
        await router.push(`/room/${room.id}`);
      } catch (navError) {
        console.error('Navigation error:', navError);
        // Если навигация не удалась, все равно показываем ошибку
        throw navError;
      }
    }
  } catch (err) {
    const apiError = err as ApiError;
    error.value = apiError.message || t('errors.roomJoinError');
    setTimeout(() => {
      error.value = null;
    }, 3000);
  } finally {
    isLoading.value = false;
  }
};

const handlePasswordSubmit = async (password: string) => {
  if (!selectedRoomForPassword.value) {
    return;
  }

  isLoading.value = true;
  error.value = null;

  try {
    // Присоединяемся к приватной комнате с кодом
    await roomApi.joinRoom(selectedRoomForPassword.value.id, password);
    try {
      await router.push(`/room/${selectedRoomForPassword.value.id}`);
    } catch (navError) {
      console.error('Navigation error:', navError);
      // Если навигация не удалась, все равно показываем ошибку
      throw navError;
    }
    showPasswordModal.value = false;
    selectedRoomForPassword.value = null;
  } catch (err) {
    const apiError = err as ApiError;
    if (apiError.code === 'HTTP_400' && apiError.message.includes('code')) {
      error.value = t('rooms.invalidPassword');
    } else {
      error.value = apiError.message || t('errors.roomJoinError');
    }
    setTimeout(() => {
      error.value = null;
    }, 3000);
  } finally {
    isLoading.value = false;
  }
};

const hasRooms = computed(() => filteredRooms.value.length > 0);
</script>

<template>
  <div class="rooms-list-page">
    <div class="page-container">
      <div class="page-header">
        <div>
          <h1 class="page-title">{{ t('rooms.title') }}</h1>
          <p class="page-subtitle">{{ t('rooms.subtitle') }}</p>
        </div>
        <div class="header-actions">
          <button
            v-if="authStore.isAuthenticated"
            class="btn btn-primary"
            @click="showCreateModal = true"
          >
            <Plus :size="18" />
            {{ t('rooms.createRoom') }}
          </button>
          <button
            class="btn btn-secondary"
            @click="showJoinModal = true"
          >
            <LogIn :size="18" />
            {{ t('rooms.joinByCode') }}
          </button>
        </div>
      </div>

      <div class="filters-section">
        <div class="search-box">
          <Search :size="20" />
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            :placeholder="t('rooms.searchPlaceholder')"
            @input="handleSearch"
          />
        </div>

        <div class="filter-buttons">
          <button
            v-for="filter in [
              { key: 'all', label: t('rooms.filterAll') },
              { key: 'my', label: t('rooms.filterMyRooms') },
              { key: 'public', label: t('rooms.filterPublic') },
              { key: 'private', label: t('rooms.filterPrivate') }
            ]"
            :key="filter.key"
            class="filter-btn"
            :class="{ active: filterType === filter.key }"
            @click="filterType = filter.key as any; handleFilterChange()"
          >
            {{ filter.label }}
          </button>
        </div>
      </div>

      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <div v-if="isLoading" class="loading-state">
        <p>Loading rooms...</p>
      </div>

      <div v-else-if="hasRooms" class="rooms-grid">
        <RoomCard
          v-for="room in filteredRooms"
          :key="room.id"
          :room="room"
          @join="handleRoomJoin"
        />
      </div>

      <div v-else class="empty-state">
        <h3>{{ t('rooms.noRooms') }}</h3>
        <p>{{ t('rooms.noRoomsDescription') }}</p>
        <button
          v-if="authStore.isAuthenticated"
          class="btn btn-primary"
          @click="showCreateModal = true"
        >
          <Plus :size="18" />
          {{ t('rooms.createRoom') }}
        </button>
      </div>
    </div>

    <!-- Create Room Modal -->
    <CreateRoomModal
      :show="showCreateModal"
      @close="showCreateModal = false"
      @created="handleCreateRoom"
    />

    <!-- Password Modal -->
    <RoomPasswordModal
      v-if="selectedRoomForPassword"
      :show="showPasswordModal"
      :room-name="selectedRoomForPassword.name"
      @close="showPasswordModal = false; selectedRoomForPassword = null"
      @submit="handlePasswordSubmit"
    />

    <!-- Join by Code Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showJoinModal" class="modal-overlay" @click.self="showJoinModal = false">
          <div class="modal-content" @click.stop>
            <div class="modal-header">
              <h2>{{ t('rooms.enterCode') }}</h2>
              <button class="close-btn" @click="showJoinModal = false">
                <X :size="20" />
              </button>
            </div>
            <div class="modal-form">
              <div class="form-group">
                <label for="joinCode" class="form-label">{{ t('rooms.codePlaceholder') }}</label>
                <input
                  id="joinCode"
                  v-model="joinCode"
                  type="text"
                  class="form-input"
                  :placeholder="t('rooms.codePlaceholder')"
                  @keyup.enter="handleJoinByCode"
                />
              </div>
              <div class="form-actions">
                <button type="button" class="btn btn-secondary" @click="showJoinModal = false">
                  {{ t('common.cancel') }}
                </button>
                <button type="button" class="btn btn-primary" @click="handleJoinByCode">
                  {{ t('rooms.join') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.rooms-list-page {
  min-height: calc(100vh - 200px);
  padding: 2rem;
  position: relative;
  z-index: 1;
  width: 100%;
}

.page-container {
  max-width: 1400px;
  margin: 0 auto;
  color: white;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  gap: 2rem;
}

.page-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
  color: white;
}

.page-subtitle {
  font-size: 1.1rem;
  color: rgba(255, 255, 255, 0.7);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 1rem;
  flex-shrink: 0;
}

.filters-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 2rem;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 0.75rem 1rem;
}

.search-box svg {
  color: rgba(255, 255, 255, 0.5);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: white;
  font-size: 1rem;
  outline: none;
}

.search-input::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

.filter-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.filter-btn {
  padding: 0.5rem 1rem;
  background: rgba(0, 0, 0, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.9rem;
}

.filter-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.filter-btn.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
  color: white;
}

.error-message {
  background: rgba(239, 68, 68, 0.2);
  border: 1px solid rgba(239, 68, 68, 0.5);
  color: #ef4444;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.loading-state {
  text-align: center;
  padding: 3rem;
  color: rgba(255, 255, 255, 0.7);
}

.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.empty-state {
  text-align: center;
  padding: 4rem 0rem;
  color: rgba(255, 255, 255, 0.7);
}

.empty-state h3 {
  font-size: 1.5rem;
  margin: 0 0 1rem 0;
  color: white;
}

.empty-state p {
  margin: 0 0 2rem 0;
  font-size: 1rem;
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

.btn-primary:hover {
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

/* Join Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  width: 100%;
  max-width: 400px;
  backdrop-filter: blur(10px);
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
  color: white;
  font-weight: 600;
}

.close-btn {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.close-btn:hover {
  color: white;
}

.modal-form {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
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

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.3s;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.9);
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    flex-direction: column;
  }

  .rooms-grid {
    grid-template-columns: 1fr;
  }

  .filter-buttons {
    justify-content: center;
  }
}
</style>
