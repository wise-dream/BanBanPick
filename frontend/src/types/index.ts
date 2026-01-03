export * from './veto';

export interface MapPool {
  id: number;
  gameId: number;
  name: string;
  type: 'all' | 'competitive' | 'custom';
  isSystem: boolean;
  maps: Map[];
}

export interface Map {
  id: number;
  name: string;
  slug: string;
  imageUrl: string;
  isCompetitive: boolean;
}

export interface Game {
  id: number;
  name: string;
  slug: string;
  isActive: boolean;
}

export interface RoomParticipant {
  id: number
  roomId: number
  userId: number
  username?: string // Никнейм пользователя
  role: 'owner' | 'member'
  joinedAt: string
}

export interface Room {
  id: number;
  ownerId: number;
  name: string;
  code: string;
  type: 'public' | 'private';
  status: 'waiting' | 'active' | 'finished';
  gameId: number;
  mapPoolId?: number;
  vetoSessionId?: number;
  maxParticipants: number;
  password?: string; // Пароль для приватных комнат (опционально)
  createdAt: string;
  updatedAt: string;
  participants?: RoomParticipant[];
}
