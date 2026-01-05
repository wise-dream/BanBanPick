<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from '../composables/useI18n';
import { useAuthStore } from '../store/auth';
import * as userApi from '../services/api/userService';
import * as roomApi from '../services/api/roomService';
import type { ApiError } from '../services/api/types';
import type { Room } from '../types';
import { Edit2, Calendar, Users, Map } from 'lucide-vue-next';

const router = useRouter();
const { t } = useI18n();
const authStore = useAuthStore();

const profile = ref<{
  id: number;
  email: string;
  username: string;
  created_at: string;
} | null>(null);
const sessions = ref<any[]>([]);
const rooms = ref<Room[]>([]);
const isLoading = ref(true);
const error = ref<string | null>(null);
const isEditing = ref(false);
const editUsername = ref('');
const editErrors = ref<Record<string, string>>({});

onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push({ path: '/login', query: { redirect: '/profile' } });
    return;
  }
  loadProfile();
  loadSessions();
  loadRooms();
});

const loadProfile = async () => {
  try {
    const response = await userApi.getProfile();
    profile.value = response;
    editUsername.value = response.username;
  } catch (err) {
    const apiError = err as ApiError;
    if (apiError.code === 'HTTP_401') {
      authStore.logout();
      router.push('/login');
    } else {
      error.value = apiError.message || t('errors.profileLoadError');
    }
  }
};

const loadSessions = async () => {
  try {
    const response = await userApi.getSessions();
    sessions.value = response;
  } catch (err) {
    console.error('Error loading sessions:', err);
  }
};

const loadRooms = async () => {
  try {
    const response = await userApi.getRooms();
    rooms.value = response.map(roomApi.roomResponseToRoom);
  } catch (err) {
    console.error('Error loading rooms:', err);
  }
};

const handleUpdateProfile = async () => {
  editErrors.value = {};

  if (!editUsername.value.trim()) {
    editErrors.value.username = t('auth.usernameRequired');
    return;
  }

  if (editUsername.value.length < 3) {
    editErrors.value.username = t('auth.usernameMinLength');
    return;
  }

  if (editUsername.value.length > 50) {
    editErrors.value.username = 'Имя пользователя должно быть не более 50 символов';
    return;
  }

  try {
    const response = await userApi.updateProfile({
      username: editUsername.value.trim(),
    });
    profile.value = response;
    authStore.setAuth(
      {
        id: response.id,
        email: response.email,
        username: response.username,
      },
      authStore.token || ''
    );
    isEditing.value = false;
  } catch (err) {
    const apiError = err as ApiError;
    if (apiError.code === 'HTTP_409') {
      editErrors.value.username = t('profile.updateUsernameError');
    } else {
      editErrors.value.username = apiError.message || t('errors.profileUpdateError');
    }
  }
};

const cancelEdit = () => {
  isEditing.value = false;
  editUsername.value = profile.value?.username || '';
  editErrors.value = {};
};

const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  return date.toLocaleDateString('ru-RU', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
};
</script>

<template>
  <div class="profile-page">
    <div class="page-container">
      <div v-if="isLoading && !profile" class="loading-state">
        <p>{{ t('common.loading') }}</p>
      </div>

      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <button class="btn btn-primary" @click="loadProfile">
          {{ t('common.retry') }}
        </button>
      </div>

      <div v-else-if="profile" class="profile-content">
        <div class="profile-header">
          <h1 class="page-title">{{ t('profile.title') }}</h1>
        </div>

        <!-- Profile Info Section -->
        <div class="profile-section">
          <div class="section-header">
            <h2 class="section-title">{{ t('profile.info') }}</h2>
            <button
              v-if="!isEditing"
              class="btn btn-secondary btn-small"
              @click="isEditing = true"
            >
              <Edit2 :size="16" />
              {{ t('profile.edit') }}
            </button>
          </div>

          <div v-if="!isEditing" class="profile-info">
            <div class="info-item">
              <span class="info-label">{{ t('profile.email') || 'Email' }}:</span>
              <span class="info-value">{{ profile.email }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('profile.username') }}:</span>
              <span class="info-value">{{ profile.username }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('profile.memberSince') }}:</span>
              <span class="info-value">
                <Calendar :size="16" />
                {{ formatDate(profile.created_at) }}
              </span>
            </div>
          </div>

          <div v-else class="edit-form">
            <div class="form-group">
              <label for="username" class="form-label">
                {{ t('profile.username') }}
              </label>
              <input
                id="username"
                v-model="editUsername"
                type="text"
                class="form-input"
                :class="{ error: editErrors.username }"
                maxlength="50"
              />
              <span v-if="editErrors.username" class="error-message">
                {{ editErrors.username }}
              </span>
            </div>
            <div class="form-actions">
              <button class="btn btn-secondary" @click="cancelEdit">
                {{ t('common.cancel') }}
              </button>
              <button class="btn btn-primary" @click="handleUpdateProfile">
                {{ t('common.save') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Sessions Section -->
        <div class="profile-section">
          <div class="section-header">
            <h2 class="section-title">
              <Map :size="20" />
              {{ t('profile.sessions') }} ({{ sessions.length }})
            </h2>
          </div>

          <div v-if="sessions.length === 0" class="empty-state">
            <p>{{ t('profile.noSessions') }}</p>
          </div>

          <div v-else class="sessions-list">
            <div
              v-for="session in sessions"
              :key="session.id"
              class="session-card"
            >
              <div class="session-info">
                <span class="session-type">{{ session.type.toUpperCase() }}</span>
                <span class="session-teams">
                  {{ session.team_a_name }} vs {{ session.team_b_name }}
                </span>
                <span class="session-status" :class="`status-${session.status}`">
                  {{ session.status }}
                </span>
              </div>
              <div class="session-date">
                {{ formatDate(session.created_at) }}
              </div>
            </div>
          </div>
        </div>

        <!-- Rooms Section -->
        <div class="profile-section">
          <div class="section-header">
            <h2 class="section-title">
              <Users :size="20" />
              {{ t('profile.rooms') }} ({{ rooms.length }})
            </h2>
          </div>

          <div v-if="rooms.length === 0" class="empty-state">
            <p>{{ t('profile.noRooms') }}</p>
            <button class="btn btn-primary" @click="router.push('/rooms')">
              {{ t('rooms.createRoom') }}
            </button>
          </div>

          <div v-else class="rooms-list">
            <div
              v-for="room in rooms"
              :key="room.id"
              class="room-card"
              @click="router.push(`/room/${room.id}`)"
            >
              <div class="room-info">
                <h3 class="room-name">{{ room.name }}</h3>
                <span class="room-code">{{ room.code }}</span>
                <span class="room-status" :class="`status-${room.status}`">
                  {{ room.status }}
                </span>
              </div>
              <div class="room-meta">
                <span class="room-participants">
                  {{ room.participants?.length || 0 }} / {{ room.maxParticipants }}
                </span>
                <span class="room-date">{{ formatDate(room.createdAt) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
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

.loading-state,
.error-state {
  text-align: center;
  padding: 3rem;
  color: rgba(255, 255, 255, 0.7);
}

.error-state {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}

.profile-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0;
  color: white;
}

.profile-section {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 2rem;
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: white;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.profile-info {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-weight: 500;
  color: rgba(255, 255, 255, 0.7);
  min-width: 150px;
}

.info-value {
  color: white;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.edit-form {
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
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
}

.form-input.error {
  border-color: #ef4444;
}

.error-message {
  font-size: 0.85rem;
  color: #ef4444;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: rgba(255, 255, 255, 0.7);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}

.sessions-list,
.rooms-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.session-card,
.room-card {
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.room-card:hover {
  border-color: rgba(102, 126, 234, 0.5);
  background: rgba(102, 126, 234, 0.1);
}

.session-info,
.room-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-bottom: 0.5rem;
}

.session-type {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.session-teams {
  color: white;
  font-weight: 500;
}

.session-status,
.room-status {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-waiting {
  background: rgba(255, 193, 7, 0.2);
  color: #ffc107;
}

.status-active {
  background: rgba(76, 175, 80, 0.2);
  color: #4caf50;
}

.status-finished {
  background: rgba(158, 158, 158, 0.2);
  color: #9e9e9e;
}

.session-date,
.room-date {
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.5);
}

.room-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: white;
  margin: 0;
}

.room-code {
  background: rgba(102, 126, 234, 0.2);
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.9rem;
  color: #667eea;
}

.room-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.5);
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
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
}

.btn-small {
  padding: 0.5rem 1rem;
  font-size: 0.9rem;
}

@media (max-width: 768px) {
  .profile-section {
    padding: 1.5rem;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .info-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .info-label {
    min-width: auto;
  }
}
</style>
