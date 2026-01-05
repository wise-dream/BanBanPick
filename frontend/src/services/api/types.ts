// Типы для API ответов и запросов

export interface ApiResponse<T = any> {
  data?: T;
  error?: ApiError;
  message?: string;
}

export interface ApiError {
  code?: string;
  message: string;
  details?: Record<string, any>;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

// Типы для авторизации
export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: UserResponse;
}

export interface UserResponse {
  id: number;
  email: string;
  username: string;
  created_at: string;
}

// Типы для Veto Sessions
export interface CreateVetoSessionRequest {
  game_id: number;
  map_pool_id: number;
  type: 'bo1' | 'bo3' | 'bo5';
  team_a_name: string;
  team_b_name: string;
  timer_seconds?: number;
}

export interface VetoSessionResponse {
  id: number;
  game_id: number;
  map_pool_id: number;
  type: string;
  status: string;
  team_a_name: string;
  team_b_name: string;
  current_team: string;
  selected_map_id?: number;
  selected_side?: string;
  timer_seconds: number;
  share_token: string;
  created_at: string;
  updated_at: string;
  finished_at?: string;
  map_pool?: MapPoolResponse;
  actions?: VetoActionResponse[];
}

export interface VetoActionResponse {
  id: number;
  map_id: number;
  map?: MapResponse;
  team: string;
  action_type: string;
  step_number: number;
  selected_side?: 'attack' | 'defence'; // Выбранная сторона для этого действия (если это pick)
  created_at: string;
}

export interface NextActionResponse {
  action_type: string;
  current_step: number;
  current_team: string;
  can_ban: boolean;
  can_pick: boolean;
  needs_side_selection: boolean;
  side_selection_team?: string;
  message?: string;
}

// Типы для Map Pools
export interface MapPoolResponse {
  id: number;
  game_id: number;
  name: string;
  type: string;
  is_system: boolean;
  maps: MapResponse[];
  created_at: string;
}

export interface CreateCustomMapPoolRequest {
  name: string;
  map_ids: number[];
}

// Типы для Maps
export interface MapResponse {
  id: number;
  name: string;
  slug: string;
  image_url: string;
  is_competitive: boolean;
}

// Типы для Rooms
export interface RoomResponse {
  id: number;
  owner_id: number;
  owner?: UserResponse;
  name: string;
  code: string;
  type: string;
  status: string;
  game_id: number;
  map_pool_id?: number;
  map_pool?: MapPoolResponse;
  veto_type?: 'bo1' | 'bo3' | 'bo5'; // Тип вето
  veto_session_id?: number;
  veto_session?: VetoSessionResponse;
  max_participants: number;
  participants_count: number;
  participants: ParticipantResponse[];
  created_at: string;
  updated_at: string;
}

export interface CreateRoomRequest {
  name: string;
  type: 'public' | 'private';
  game_id: number;
  map_pool_id?: number;
  veto_type?: 'bo1' | 'bo3' | 'bo5'; // Тип вето
  max_participants?: number;
  password?: string; // Пароль для приватных комнат (опционально)
}

export interface UpdateRoomRequest {
  map_pool_id?: number;
  veto_type?: 'bo1' | 'bo3' | 'bo5'; // Тип вето
  veto_session_id?: number;
  status?: 'waiting' | 'active' | 'finished';
}

export interface JoinRoomRequest {
  password?: string; // Пароль для приватных комнат
}

export interface UpdateRoomRequest {
  veto_session_id?: number; // ID сессии вето
  status?: 'waiting' | 'active' | 'finished'; // Статус комнаты
}

export interface ParticipantResponse {
  id: number;
  user_id: number;
  user?: UserResponse;
  username?: string; // Никнейм пользователя (может приходить напрямую от бэкенда или из user)
  role: string;
  joined_at: string;
}

export interface RoomParticipant {
  id: number
  roomId: number
  userId: number
  username?: string // Никнейм пользователя
  role: 'owner' | 'member'
  joinedAt: string
}

// Типы для Profile
export interface UpdateProfileRequest {
  username?: string;
}

export interface ProfileResponse {
  id: number;
  email: string;
  username: string;
  created_at: string;
}
