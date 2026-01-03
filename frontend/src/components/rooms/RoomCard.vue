<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from '../../composables/useI18n';
import type { Room } from '../../types';
import { Users, Lock, Unlock, Clock, Play } from 'lucide-vue-next';

interface Props {
  room: Room;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  join: [room: Room];
}>();

const router = useRouter();
const { t } = useI18n();

const participantsCount = computed(() => props.room.participants?.length || 0);
const isFull = computed(() => participantsCount.value >= 2); // Fixed: max 2 participants
const canJoin = computed(() => props.room.status === 'waiting' && !isFull.value);

const statusLabel = computed(() => {
  switch (props.room.status) {
    case 'waiting':
      return t('rooms.status.waiting');
    case 'active':
      return t('rooms.status.active');
    case 'finished':
      return t('rooms.status.finished');
    default:
      return props.room.status;
  }
});

const statusClass = computed(() => {
  switch (props.room.status) {
    case 'waiting':
      return 'status-waiting';
    case 'active':
      return 'status-active';
    case 'finished':
      return 'status-finished';
    default:
      return '';
  }
});

const handleClick = () => {
  if (canJoin.value || props.room.status === 'active') {
    // Для приватных комнат показываем popup для ввода пароля
    if (props.room.type === 'private' && canJoin.value) {
      emit('join', props.room);
    } else {
      router.push(`/room/${props.room.id}`);
    }
  }
};
</script>

<template>
  <div
    class="room-card"
    :class="{ 'can-join': canJoin, 'is-full': isFull, 'is-finished': room.status === 'finished' }"
    @click="handleClick"
  >
    <div class="card-header">
      <div class="room-title-section">
        <h3 class="room-name">{{ room.name }}</h3>
        <span class="room-code">#{{ room.code }}</span>
      </div>
      <div class="room-badges">
        <span :class="['status-badge', statusClass]">
          <Clock v-if="room.status === 'waiting'" :size="14" />
          <Play v-if="room.status === 'active'" :size="14" />
          {{ statusLabel }}
        </span>
        <span v-if="room.type === 'private'" class="type-badge private">
          <Lock :size="12" />
        </span>
        <span v-else class="type-badge public">
          <Unlock :size="12" />
        </span>
      </div>
    </div>

    <div class="card-content">
      <div class="room-info">
        <div class="info-item">
          <Users :size="16" />
          <span>{{ participantsCount }} / 2</span>
        </div>
        <div v-if="room.mapPoolId" class="info-item">
          <span>{{ t('rooms.mapPool') }}: {{ room.mapPoolId }}</span>
        </div>
      </div>
    </div>

    <div class="card-footer">
      <button
        v-if="canJoin"
        class="btn-join"
        @click.stop="handleClick"
      >
        {{ t('rooms.join') }}
      </button>
      <button
        v-else-if="room.status === 'active'"
        class="btn-view"
        @click.stop="handleClick"
      >
        {{ t('rooms.view') }}
      </button>
      <span v-else class="room-unavailable">
        {{ isFull ? t('rooms.full') : t('rooms.unavailable') }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.room-card {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.room-card:hover {
  border-color: rgba(102, 126, 234, 0.5);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.2);
}

.room-card.can-join {
  border-color: rgba(76, 175, 80, 0.5);
}

.room-card.can-join:hover {
  border-color: #4caf50;
  box-shadow: 0 8px 24px rgba(76, 175, 80, 0.3);
}

.room-card.is-full,
.room-card.is-finished {
  opacity: 0.6;
  cursor: not-allowed;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.room-title-section {
  flex: 1;
}

.room-name {
  margin: 0 0 0.25rem 0;
  font-size: 1.25rem;
  color: white;
  font-weight: 600;
}

.room-code {
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.6);
  font-family: monospace;
}

.room-badges {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-badge.status-waiting {
  background: rgba(255, 193, 7, 0.2);
  color: #ffc107;
}

.status-badge.status-active {
  background: rgba(76, 175, 80, 0.2);
  color: #4caf50;
}

.status-badge.status-finished {
  background: rgba(158, 158, 158, 0.2);
  color: #9e9e9e;
}

.type-badge {
  display: flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  border-radius: 8px;
  font-size: 0.75rem;
}

.type-badge.private {
  background: rgba(244, 67, 54, 0.2);
  color: #f44336;
}

.type-badge.public {
  background: rgba(33, 150, 243, 0.2);
  color: #2196f3;
}

.card-content {
  flex: 1;
}

.room-info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.9rem;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: auto;
}

.btn-join,
.btn-view {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 0.5rem 1.5rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
  font-size: 0.9rem;
}

.btn-join:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-view {
  background: rgba(102, 126, 234, 0.3);
}

.btn-view:hover {
  background: rgba(102, 126, 234, 0.5);
}

.room-unavailable {
  color: rgba(255, 255, 255, 0.5);
  font-size: 0.85rem;
  font-style: italic;
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .room-badges {
    width: 100%;
  }
}
</style>
