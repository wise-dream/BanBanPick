# Сущности бэкенда (Clean Architecture + Go + Gin + GORM + PostgreSQL)

## Структура проекта (Clean Architecture)

```
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── user.go
│   │   │   ├── game.go
│   │   │   ├── map.go
│   │   │   ├── map_pool.go
│   │   │   ├── veto_session.go
│   │   │   ├── veto_action.go
│   │   │   ├── room.go
│   │   │   └── room_participant.go
│   │   └── repositories/
│   │       ├── user_repository.go
│   │       ├── game_repository.go
│   │       ├── map_repository.go
│   │       ├── map_pool_repository.go
│   │       ├── veto_session_repository.go
│   │       └── room_repository.go
│   ├── usecase/
│   │   ├── auth/
│   │   │   ├── register.go
│   │   │   ├── login.go
│   │   │   └── logout.go
│   │   ├── user/
│   │   │   ├── get_profile.go
│   │   │   └── update_profile.go
│   │   ├── map_pool/
│   │   │   ├── get_pools.go
│   │   │   ├── create_custom_pool.go
│   │   │   └── delete_pool.go
│   │   └── veto/
│   │       ├── create_session.go
│   │       ├── get_session.go
│   │       ├── ban_map.go
│   │       ├── pick_map.go
│   │       ├── select_side.go
│   │       └── reset_session.go
│   ├── repository/
│   │   ├── postgres/
│   │   │   ├── user_repository.go
│   │   │   ├── game_repository.go
│   │   │   ├── map_repository.go
│   │   │   ├── map_pool_repository.go
│   │   │   └── veto_session_repository.go
│   │   └── models/
│   │       ├── user_model.go
│   │       ├── game_model.go
│   │       ├── map_model.go
│   │       ├── map_pool_model.go
│   │       ├── veto_session_model.go
│   │       └── veto_action_model.go
│   ├── handler/
│   │   ├── http/
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── map_pool_handler.go
│   │   │   └── veto_handler.go
│   │   └── dto/
│   │       ├── auth_dto.go
│   │       ├── user_dto.go
│   │       ├── map_pool_dto.go
│   │       └── veto_dto.go
│   └── middleware/
│       ├── auth.go
│       ├── cors.go
│       └── logger.go
├── pkg/
│   ├── database/
│   │   └── postgres.go
│   ├── jwt/
│   │   └── jwt.go
│   └── password/
│       └── hasher.go
├── config/
│   └── config.go
├── migrations/
│   └── *.sql
└── go.mod
```

---

## Domain Entities (Сущности домена)

### 1. User (Пользователь)

```go
// internal/domain/entities/user.go
package entities

import "time"

type User struct {
    ID        uint      `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    Password  string    `json:"-"` // не возвращается в JSON
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // Relations
    VetoSessions []VetoSession `json:"veto_sessions,omitempty" gorm:"foreignKey:UserID"`
    MapPools      []MapPool     `json:"map_pools,omitempty" gorm:"foreignKey:UserID"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `Email` - email пользователя (уникальный)
- `Username` - имя пользователя (уникальное)
- `Password` - хеш пароля (bcrypt)
- `CreatedAt` - дата создания
- `UpdatedAt` - дата обновления

---

### 2. Game (Игра)

```go
// internal/domain/entities/game.go
package entities

import "time"

type Game struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`      // "Valorant"
    Slug      string    `json:"slug"`      // "valorant"
    IsActive  bool      `json:"is_active"`  // пока только Valorant = true
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // Relations
    Maps      []Map      `json:"maps,omitempty" gorm:"foreignKey:GameID"`
    MapPools  []MapPool  `json:"map_pools,omitempty" gorm:"foreignKey:GameID"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `Name` - название игры
- `Slug` - URL-friendly идентификатор
- `IsActive` - активна ли игра (пока только Valorant)
- `CreatedAt` - дата создания
- `UpdatedAt` - дата обновления

**Начальные данные:**
- Valorant (ID: 1, Slug: "valorant", IsActive: true)

---

### 3. Map (Карта)

```go
// internal/domain/entities/map.go
package entities

import "time"

type Map struct {
    ID        uint      `json:"id"`
    GameID    uint      `json:"game_id"`
    Name      string    `json:"name"`        // "Ascent", "Bind", etc.
    Slug      string    `json:"slug"`        // "ascent", "bind", etc.
    ImageURL  string    `json:"image_url"`   // путь к изображению
    IsActive  bool      `json:"is_active"`    // активна ли карта в пуле
    IsCompetitive bool  `json:"is_competitive"` // в соревновательном пуле
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // Relations
    Game      Game      `json:"game,omitempty" gorm:"foreignKey:GameID"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `GameID` - ID игры (пока только Valorant)
- `Name` - название карты
- `Slug` - URL-friendly идентификатор
- `ImageURL` - путь к изображению карты
- `IsActive` - активна ли карта
- `IsCompetitive` - входит ли в соревновательный пул
- `CreatedAt` - дата создания
- `UpdatedAt` - дата обновления

**Начальные данные для Valorant:**
- All Maps: Abyss, Ascent, Bind, Breeze, Corrode, Fracture, Haven, Icebox, Lotus, Pearl, Split, Sunset
- Competitive Maps: Abyss, Bind, Corrode, Haven, Pearl, Split, Sunset

---

### 4. MapPool (Пул карт)

```go
// internal/domain/entities/map_pool.go
package entities

import "time"

type MapPoolType string

const (
    MapPoolTypeAll         MapPoolType = "all"         // все карты
    MapPoolTypeCompetitive MapPoolType = "competitive" // соревновательный
    MapPoolTypeCustom      MapPoolType = "custom"     // пользовательский
)

type MapPool struct {
    ID          uint         `json:"id"`
    GameID      uint         `json:"game_id"`
    UserID      *uint        `json:"user_id,omitempty"` // null для системных пулов
    Name        string       `json:"name"`               // "All Maps", "Competitive Maps", или кастомное имя
    Type        MapPoolType  `json:"type"`               // all, competitive, custom
    IsSystem    bool         `json:"is_system"`          // системный пул (нельзя удалить)
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
    
    // Relations
    Game        Game         `json:"game,omitempty" gorm:"foreignKey:GameID"`
    User        *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Maps        []Map        `json:"maps" gorm:"many2many:map_pool_maps;"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `GameID` - ID игры
- `UserID` - ID пользователя (null для системных пулов)
- `Name` - название пула
- `Type` - тип пула (all, competitive, custom)
- `IsSystem` - системный пул (нельзя удалить)
- `CreatedAt` - дата создания
- `UpdatedAt` - дата обновления

**Связь many-to-many:**
- Таблица `map_pool_maps` для связи пулов и карт

---

### 5. VetoSession (Сессия вето)

```go
// internal/domain/entities/veto_session.go
package entities

import "time"

type VetoStatus string

const (
    VetoStatusNotStarted VetoStatus = "not_started"
    VetoStatusInProgress VetoStatus = "in_progress"
    VetoStatusFinished   VetoStatus = "finished"
    VetoStatusCancelled  VetoStatus = "cancelled"
)

type VetoType string

const (
    VetoTypeBo1 VetoType = "bo1" // ban до последней карты
    VetoTypeBo3 VetoType = "bo3" // бан, бан, пик, бан, бан, пик, бан, бан, десидер
    VetoTypeBo5 VetoType = "bo5" // бан, бан, бан или пик, бан или пик, пик, бан, бан, пик, бан, бан, пик, десидер
)

type VetoSession struct {
    ID              uint         `json:"id"`
    UserID          *uint        `json:"user_id,omitempty"` // null для анонимных сессий
    GameID          uint         `json:"game_id"`
    MapPoolID       uint         `json:"map_pool_id"`
    Type            VetoType     `json:"type"`              // bo1, bo3, bo5
    Status          VetoStatus   `json:"status"`            // not_started, in_progress, finished, cancelled
    TeamAName       string       `json:"team_a_name"`       // "Team A"
    TeamBName       string       `json:"team_b_name"`       // "Team B"
    CurrentTeam     string       `json:"current_team"`      // "A" или "B"
    SelectedMapID   *uint        `json:"selected_map_id,omitempty"` // выбранная карта
    SelectedSide    *string      `json:"selected_side,omitempty"`   // "attack" или "defence"
    TimerSeconds    int          `json:"timer_seconds"`     // таймер для бана (0 = отключен)
    ShareToken      string       `json:"share_token"`       // токен для публичного доступа
    CreatedAt       time.Time    `json:"created_at"`
    UpdatedAt       time.Time    `json:"updated_at"`
    FinishedAt      *time.Time   `json:"finished_at,omitempty"`
    
    // Relations
    User            *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Game            Game         `json:"game,omitempty" gorm:"foreignKey:GameID"`
    MapPool         MapPool      `json:"map_pool,omitempty" gorm:"foreignKey:MapPoolID"`
    SelectedMap     *Map         `json:"selected_map,omitempty" gorm:"foreignKey:SelectedMapID"`
    Actions         []VetoAction `json:"actions,omitempty" gorm:"foreignKey:VetoSessionID"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `UserID` - ID пользователя (null для анонимных сессий)
- `GameID` - ID игры
- `MapPoolID` - ID пула карт
- `Type` - тип вето (bo1, bo3, bo5)
- `Status` - статус сессии
- `TeamAName` - название команды A
- `TeamBName` - название команды B
- `CurrentTeam` - текущая команда ("A" или "B")
- `SelectedMapID` - ID выбранной карты
- `SelectedSide` - выбранная сторона ("attack" или "defence")
- `TimerSeconds` - таймер для бана в секундах
- `ShareToken` - токен для публичного доступа к сессии
- `CreatedAt` - дата создания
- `UpdatedAt` - дата обновления
- `FinishedAt` - дата завершения

---

### 6. VetoAction (Действие в вето)

```go
// internal/domain/entities/veto_action.go
package entities

import "time"

type VetoActionType string

const (
    VetoActionTypeBan  VetoActionType = "ban"  // бан карты
    VetoActionTypePick VetoActionType = "pick" // пик карты
)

type VetoAction struct {
    ID            uint            `json:"id"`
    VetoSessionID uint            `json:"veto_session_id"`
    MapID         uint            `json:"map_id"`
    Team          string          `json:"team"`          // "A" или "B"
    ActionType    VetoActionType  `json:"action_type"`   // "ban" или "pick"
    StepNumber    int             `json:"step_number"`   // номер шага в процессе
    CreatedAt     time.Time       `json:"created_at"`
    
    // Relations
    VetoSession   VetoSession     `json:"veto_session,omitempty" gorm:"foreignKey:VetoSessionID"`
    Map           Map             `json:"map,omitempty" gorm:"foreignKey:MapID"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `VetoSessionID` - ID сессии вето
- `MapID` - ID карты
- `Team` - команда, выполнившая действие ("A" или "B")
- `ActionType` - тип действия (ban или pick)
- `StepNumber` - номер шага в процессе
- `CreatedAt` - дата создания

---

### 7. Room (Комната)

```go
// internal/domain/entities/room.go
package entities

import "time"

type RoomType string

const (
    RoomTypePublic  RoomType = "public"  // публичная комната
    RoomTypePrivate RoomType = "private" // приватная комната (по коду)
)

type RoomStatus string

const (
    RoomStatusWaiting  RoomStatus = "waiting"  // ожидание участников
    RoomStatusActive   RoomStatus = "active"   // активная (вето в процессе)
    RoomStatusFinished RoomStatus = "finished" // завершена
)

type Room struct {
    ID            uint         `json:"id"`
    OwnerID       uint         `json:"owner_id"`
    Name          string       `json:"name"`           // название комнаты
    Code          string       `json:"code"`           // уникальный код для присоединения
    Type          RoomType     `json:"type"`           // public или private
    Status        RoomStatus   `json:"status"`         // waiting, active, finished
    GameID        uint         `json:"game_id"`        // пока только Valorant
    MapPoolID     *uint        `json:"map_pool_id,omitempty"` // выбранный пул карт
    VetoSessionID *uint        `json:"veto_session_id,omitempty"` // активная сессия вето
    MaxParticipants int        `json:"max_participants"` // максимум участников (по умолчанию 10)
    CreatedAt     time.Time    `json:"created_at"`
    UpdatedAt     time.Time    `json:"updated_at"`
    
    // Relations
    Owner         User         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
    Game          Game         `json:"game,omitempty" gorm:"foreignKey:GameID"`
    MapPool       *MapPool     `json:"map_pool,omitempty" gorm:"foreignKey:MapPoolID"`
    VetoSession   *VetoSession `json:"veto_session,omitempty" gorm:"foreignKey:VetoSessionID"`
    Participants  []RoomParticipant `json:"participants,omitempty" gorm:"foreignKey:RoomID"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `OwnerID` - ID владельца комнаты
- `Name` - название комнаты
- `Code` - уникальный код для присоединения (6-8 символов)
- `Type` - тип комнаты (public или private)
- `Status` - статус комнаты
- `GameID` - ID игры (пока только Valorant)
- `MapPoolID` - ID выбранного пула карт
- `VetoSessionID` - ID активной сессии вето
- `MaxParticipants` - максимальное количество участников
- `CreatedAt` - дата создания
- `UpdatedAt` - дата обновления

---

### 8. RoomParticipant (Участник комнаты)

```go
// internal/domain/entities/room_participant.go
package entities

import "time"

type ParticipantRole string

const (
    ParticipantRoleOwner  ParticipantRole = "owner"  // владелец
    ParticipantRoleMember ParticipantRole = "member" // участник
)

type RoomParticipant struct {
    ID        uint             `json:"id"`
    RoomID    uint             `json:"room_id"`
    UserID    uint             `json:"user_id"`
    Role      ParticipantRole  `json:"role"`         // owner или member
    JoinedAt  time.Time        `json:"joined_at"`
    
    // Relations
    Room      Room             `json:"room,omitempty" gorm:"foreignKey:RoomID"`
    User      User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
```

**Поля:**
- `ID` - уникальный идентификатор
- `RoomID` - ID комнаты
- `UserID` - ID пользователя
- `Role` - роль участника (owner или member)
- `JoinedAt` - дата присоединения

---

## Database Models (GORM Models)

### User Model

```go
// internal/repository/models/user_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type UserModel struct {
    ID        uint           `gorm:"primaryKey"`
    Email     string         `gorm:"uniqueIndex;not null;size:255"`
    Username  string         `gorm:"uniqueIndex;not null;size:100"`
    Password  string         `gorm:"not null;size:255"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

---

### Game Model

```go
// internal/repository/models/game_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type GameModel struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"not null;size:100"`
    Slug      string         `gorm:"uniqueIndex;not null;size:50"`
    IsActive  bool           `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

---

### Map Model

```go
// internal/repository/models/map_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type MapModel struct {
    ID            uint           `gorm:"primaryKey"`
    GameID        uint           `gorm:"not null;index"`
    Name          string         `gorm:"not null;size:100"`
    Slug          string         `gorm:"not null;size:100"`
    ImageURL      string         `gorm:"size:255"`
    IsActive      bool           `gorm:"default:true"`
    IsCompetitive bool           `gorm:"default:false"`
    CreatedAt     time.Time
    UpdatedAt   time.Time
    DeletedAt     gorm.DeletedAt `gorm:"index"`
}
```

---

### MapPool Model

```go
// internal/repository/models/map_pool_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type MapPoolModel struct {
    ID        uint           `gorm:"primaryKey"`
    GameID    uint           `gorm:"not null;index"`
    UserID    *uint          `gorm:"index"`
    Name      string         `gorm:"not null;size:255"`
    Type      string         `gorm:"not null;size:50"` // all, competitive, custom
    IsSystem  bool           `gorm:"default:false"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // Many-to-many relation
    Maps      []MapModel     `gorm:"many2many:map_pool_maps;"`
}
```

---

### VetoSession Model

```go
// internal/repository/models/veto_session_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type VetoSessionModel struct {
    ID            uint           `gorm:"primaryKey"`
    UserID        *uint          `gorm:"index"`
    GameID        uint           `gorm:"not null;index"`
    MapPoolID     uint           `gorm:"not null;index"`
    Type          string         `gorm:"not null;size:10"` // bo1, bo3, bo5
    Status        string         `gorm:"not null;size:20"` // not_started, in_progress, finished, cancelled
    TeamAName     string         `gorm:"not null;size:100"`
    TeamBName     string         `gorm:"not null;size:100"`
    CurrentTeam   string         `gorm:"not null;size:1"` // A или B
    SelectedMapID *uint          `gorm:"index"`
    SelectedSide  *string         `gorm:"size:20"` // attack или defence
    TimerSeconds  int            `gorm:"default:0"`
    ShareToken    string         `gorm:"uniqueIndex;not null;size:64"`
    CreatedAt     time.Time
    UpdatedAt      time.Time
    FinishedAt     *time.Time
    DeletedAt      gorm.DeletedAt `gorm:"index"`
}
```

---

### VetoAction Model

```go
// internal/repository/models/veto_action_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type VetoActionModel struct {
    ID            uint           `gorm:"primaryKey"`
    VetoSessionID uint           `gorm:"not null;index"`
    MapID         uint           `gorm:"not null;index"`
    Team          string         `gorm:"not null;size:1"` // A или B
    ActionType    string         `gorm:"not null;size:10"` // ban или pick
    StepNumber    int            `gorm:"not null"`
    CreatedAt     time.Time
    DeletedAt     gorm.DeletedAt `gorm:"index"`
}
```

---

### Room Model

```go
// internal/repository/models/room_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type RoomModel struct {
    ID              uint           `gorm:"primaryKey"`
    OwnerID          uint           `gorm:"not null;index"`
    Name             string         `gorm:"not null;size:255"`
    Code             string         `gorm:"uniqueIndex;not null;size:8"`
    Type             string         `gorm:"not null;size:20"` // public или private
    Status           string         `gorm:"not null;size:20"` // waiting, active, finished
    GameID           uint           `gorm:"not null;index"`
    MapPoolID        *uint          `gorm:"index"`
    VetoSessionID    *uint          `gorm:"index"`
    MaxParticipants  int            `gorm:"default:10"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        gorm.DeletedAt `gorm:"index"`
}
```

---

### RoomParticipant Model

```go
// internal/repository/models/room_participant_model.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type RoomParticipantModel struct {
    ID        uint           `gorm:"primaryKey"`
    RoomID    uint           `gorm:"not null;index"`
    UserID    uint           `gorm:"not null;index"`
    Role      string         `gorm:"not null;size:20"` // owner или member
    JoinedAt  time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // Unique constraint: один пользователь может быть только в одной комнате одновременно
    // (можно реализовать через составной индекс)
}
```

---

## DTOs (Data Transfer Objects)

### Auth DTOs

```go
// internal/handler/dto/auth_dto.go
package dto

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
    Token string      `json:"token"`
    User  UserResponse `json:"user"`
}
```

---

### User DTOs

```go
// internal/handler/dto/user_dto.go
package dto

import "time"

type UserResponse struct {
    ID        uint      `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"created_at"`
}

type UpdateProfileRequest struct {
    Username string `json:"username" binding:"min=3,max=50"`
}
```

---

### MapPool DTOs

```go
// internal/handler/dto/map_pool_dto.go
package dto

import "time"

type MapPoolResponse struct {
    ID        uint      `json:"id"`
    GameID  uint      `json:"game_id"`
    Name      string    `json:"name"`
    Type      string    `json:"type"`
    IsSystem  bool      `json:"is_system"`
    Maps      []MapResponse `json:"maps"`
    CreatedAt time.Time `json:"created_at"`
}

type MapResponse struct {
    ID            uint   `json:"id"`
    Name          string `json:"name"`
    Slug          string `json:"slug"`
    ImageURL      string `json:"image_url"`
    IsCompetitive bool   `json:"is_competitive"`
}

type CreateCustomMapPoolRequest struct {
    Name  string   `json:"name" binding:"required,min=1,max=255"`
    MapIDs []uint  `json:"map_ids" binding:"required,min=1"`
}
```

---

### Veto DTOs

```go
// internal/handler/dto/veto_dto.go
package dto

import "time"

type CreateVetoSessionRequest struct {
    GameID     uint   `json:"game_id" binding:"required"`
    MapPoolID  uint   `json:"map_pool_id" binding:"required"`
    Type       string `json:"type" binding:"required,oneof=bo1 bo3 bo5"`
    TeamAName  string `json:"team_a_name" binding:"required,min=1,max=100"`
    TeamBName  string `json:"team_b_name" binding:"required,min=1,max=100"`
    TimerSeconds int  `json:"timer_seconds" binding:"min=0,max=300"`
}

type VetoSessionResponse struct {
    ID            uint                `json:"id"`
    GameID        uint                `json:"game_id"`
    MapPoolID     uint                `json:"map_pool_id"`
    Type          string              `json:"type"`
    Status        string              `json:"status"`
    TeamAName     string              `json:"team_a_name"`
    TeamBName     string              `json:"team_b_name"`
    CurrentTeam   string              `json:"current_team"`
    SelectedMapID *uint               `json:"selected_map_id"`
    SelectedSide  *string              `json:"selected_side"`
    TimerSeconds  int                 `json:"timer_seconds"`
    ShareToken    string              `json:"share_token"`
    CreatedAt     time.Time           `json:"created_at"`
    UpdatedAt    time.Time           `json:"updated_at"`
    FinishedAt    *time.Time          `json:"finished_at"`
    MapPool       MapPoolResponse     `json:"map_pool"`
    Actions       []VetoActionResponse `json:"actions"`
}

type VetoActionResponse struct {
    ID         uint      `json:"id"`
    MapID      uint      `json:"map_id"`
    Map        MapResponse `json:"map"`
    Team       string    `json:"team"`
    ActionType string    `json:"action_type"`
    StepNumber int       `json:"step_number"`
    CreatedAt  time.Time `json:"created_at"`
}

type BanMapRequest struct {
    MapID uint `json:"map_id" binding:"required"`
}

type PickMapRequest struct {
    MapID uint `json:"map_id" binding:"required"`
}

type SelectSideRequest struct {
    Side string `json:"side" binding:"required,oneof=attack defence"`
}
```

---

### Room DTOs

```go
// internal/handler/dto/room_dto.go
package dto

import "time"

type CreateRoomRequest struct {
    Name         string `json:"name" binding:"required,min=1,max=255"`
    Type         string `json:"type" binding:"required,oneof=public private"`
    GameID       uint   `json:"game_id" binding:"required"`
    MapPoolID    *uint  `json:"map_pool_id"`
    MaxParticipants int `json:"max_participants" binding:"min=2,max=20"`
}

type RoomResponse struct {
    ID             uint                `json:"id"`
    OwnerID         uint                `json:"owner_id"`
    Owner           UserResponse        `json:"owner"`
    Name            string              `json:"name"`
    Code            string              `json:"code"`
    Type            string              `json:"type"`
    Status          string              `json:"status"`
    GameID          uint                `json:"game_id"`
    MapPoolID       *uint               `json:"map_pool_id"`
    MapPool         *MapPoolResponse     `json:"map_pool,omitempty"`
    VetoSessionID   *uint               `json:"veto_session_id"`
    VetoSession     *VetoSessionResponse `json:"veto_session,omitempty"`
    MaxParticipants int                 `json:"max_participants"`
    ParticipantsCount int               `json:"participants_count"`
    Participants     []ParticipantResponse `json:"participants"`
    CreatedAt        time.Time           `json:"created_at"`
    UpdatedAt        time.Time           `json:"updated_at"`
}

type ParticipantResponse struct {
    ID       uint      `json:"id"`
    UserID   uint      `json:"user_id"`
    User     UserResponse `json:"user"`
    Role     string    `json:"role"`
    JoinedAt time.Time `json:"joined_at"`
}

type JoinRoomRequest struct {
    Code string `json:"code" binding:"required,min=6,max=8"`
}

type UpdateRoomRequest struct {
    Name         *string `json:"name"`
    MapPoolID    *uint   `json:"map_pool_id"`
    MaxParticipants *int `json:"max_participants"`
}
```

---

## Repository Interfaces

### User Repository

```go
// internal/domain/repositories/user_repository.go
package repositories

import "github.com/yourproject/internal/domain/entities"

type UserRepository interface {
    Create(user *entities.User) error
    GetByID(id uint) (*entities.User, error)
    GetByEmail(email string) (*entities.User, error)
    GetByUsername(username string) (*entities.User, error)
    Update(user *entities.User) error
    Delete(id uint) error
}
```

### MapPool Repository

```go
// internal/domain/repositories/map_pool_repository.go
package repositories

import "github.com/yourproject/internal/domain/entities"

type MapPoolRepository interface {
    Create(pool *entities.MapPool) error
    GetByID(id uint) (*entities.MapPool, error)
    GetByGameID(gameID uint) ([]entities.MapPool, error)
    GetByUserID(userID uint) ([]entities.MapPool, error)
    GetSystemPools(gameID uint) ([]entities.MapPool, error)
    Update(pool *entities.MapPool) error
    Delete(id uint) error
}
```

### VetoSession Repository

```go
// internal/domain/repositories/veto_session_repository.go
package repositories

import "github.com/yourproject/internal/domain/entities"

type VetoSessionRepository interface {
    Create(session *entities.VetoSession) error
    GetByID(id uint) (*entities.VetoSession, error)
    GetByShareToken(token string) (*entities.VetoSession, error)
    GetByUserID(userID uint) ([]entities.VetoSession, error)
    Update(session *entities.VetoSession) error
    Delete(id uint) error
}
```

### Room Repository

```go
// internal/domain/repositories/room_repository.go
package repositories

import "github.com/yourproject/internal/domain/entities"

type RoomRepository interface {
    Create(room *entities.Room) error
    GetByID(id uint) (*entities.Room, error)
    GetByCode(code string) (*entities.Room, error)
    GetByOwnerID(ownerID uint) ([]entities.Room, error)
    GetPublicRooms(limit, offset int) ([]entities.Room, error)
    Update(room *entities.Room) error
    Delete(id uint) error
    AddParticipant(participant *entities.RoomParticipant) error
    RemoveParticipant(roomID, userID uint) error
    GetParticipants(roomID uint) ([]entities.RoomParticipant, error)
    GetParticipant(roomID, userID uint) (*entities.RoomParticipant, error)
}
```

---

## API Endpoints

### Auth Endpoints

```
POST   /api/auth/register     - Регистрация
POST   /api/auth/login        - Вход
POST   /api/auth/logout       - Выход (требует авторизации)
GET    /api/auth/me           - Получить текущего пользователя
```

### User Endpoints

```
GET    /api/users/profile     - Получить профиль (требует авторизации)
PUT    /api/users/profile     - Обновить профиль (требует авторизации)
GET    /api/users/sessions    - История сессий (требует авторизации)
```

### MapPool Endpoints

```
GET    /api/games/:gameId/map-pools           - Получить пулы для игры
GET    /api/map-pools/:id                     - Получить пул по ID
POST   /api/map-pools                         - Создать кастомный пул (требует авторизации)
DELETE /api/map-pools/:id                     - Удалить пул (требует авторизации)
```

### Veto Endpoints

```
POST   /api/veto/sessions                     - Создать сессию вето
GET    /api/veto/sessions/:id                 - Получить сессию по ID
GET    /api/veto/sessions/share/:token       - Получить сессию по токену
POST   /api/veto/sessions/:id/ban             - Забанить карту
POST   /api/veto/sessions/:id/pick            - Выбрать карту
POST   /api/veto/sessions/:id/start           - Начать вето
POST   /api/veto/sessions/:id/reset           - Сбросить сессию
POST   /api/veto/sessions/:id/select-side     - Выбрать сторону
POST   /api/veto/sessions/:id/swap-team       - Сменить команду
```

### Room Endpoints

```
GET    /api/rooms                             - Список комнат (публичные)
POST   /api/rooms                             - Создать комнату (требует авторизации)
GET    /api/rooms/:id                          - Получить комнату по ID
POST   /api/rooms/:id/join                     - Присоединиться к комнате (по коду или ID)
POST   /api/rooms/:id/leave                    - Покинуть комнату (требует авторизации)
PUT    /api/rooms/:id                          - Обновить комнату (только владелец)
DELETE /api/rooms/:id                          - Удалить комнату (только владелец)
GET    /api/rooms/:id/participants              - Получить список участников
GET    /api/users/rooms                        - Мои комнаты (требует авторизации)
```

### WebSocket Endpoints

```
WS     /ws/room/:roomId                        - Подключение к комнате для real-time синхронизации
```

---

## Миграции базы данных

### Initial Migration

```sql
-- migrations/001_initial.sql

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- Games table
CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_games_slug ON games(slug);

-- Maps table
CREATE TABLE maps (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    image_url VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    is_competitive BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_maps_game_id ON maps(game_id);

-- Map pools table
CREATE TABLE map_pools (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id),
    user_id INTEGER REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_map_pools_game_id ON map_pools(game_id);
CREATE INDEX idx_map_pools_user_id ON map_pools(user_id);

-- Map pool maps (many-to-many)
CREATE TABLE map_pool_maps (
    map_pool_id INTEGER NOT NULL REFERENCES map_pools(id) ON DELETE CASCADE,
    map_id INTEGER NOT NULL REFERENCES maps(id) ON DELETE CASCADE,
    PRIMARY KEY (map_pool_id, map_id)
);

-- Veto sessions table
CREATE TABLE veto_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    game_id INTEGER NOT NULL REFERENCES games(id),
    map_pool_id INTEGER NOT NULL REFERENCES map_pools(id),
    type VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL,
    team_a_name VARCHAR(100) NOT NULL,
    team_b_name VARCHAR(100) NOT NULL,
    current_team VARCHAR(1) NOT NULL,
    selected_map_id INTEGER REFERENCES maps(id),
    selected_side VARCHAR(20),
    timer_seconds INTEGER DEFAULT 0,
    share_token VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_veto_sessions_user_id ON veto_sessions(user_id);
CREATE INDEX idx_veto_sessions_game_id ON veto_sessions(game_id);
CREATE INDEX idx_veto_sessions_map_pool_id ON veto_sessions(map_pool_id);
CREATE INDEX idx_veto_sessions_share_token ON veto_sessions(share_token);

-- Veto actions table
CREATE TABLE veto_actions (
    id SERIAL PRIMARY KEY,
    veto_session_id INTEGER NOT NULL REFERENCES veto_sessions(id) ON DELETE CASCADE,
    map_id INTEGER NOT NULL REFERENCES maps(id),
    team VARCHAR(1) NOT NULL,
    action_type VARCHAR(10) NOT NULL,
    step_number INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_veto_actions_veto_session_id ON veto_actions(veto_session_id);
CREATE INDEX idx_veto_actions_map_id ON veto_actions(map_id);

-- Rooms table
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(8) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    game_id INTEGER NOT NULL REFERENCES games(id),
    map_pool_id INTEGER REFERENCES map_pools(id),
    veto_session_id INTEGER REFERENCES veto_sessions(id),
    max_participants INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX idx_rooms_code ON rooms(code);
CREATE INDEX idx_rooms_game_id ON rooms(game_id);
CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_rooms_type ON rooms(type);

-- Room participants table
CREATE TABLE room_participants (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    role VARCHAR(20) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(room_id, user_id) -- один пользователь может быть только в одной комнате одновременно
);

CREATE INDEX idx_room_participants_room_id ON room_participants(room_id);
CREATE INDEX idx_room_participants_user_id ON room_participants(user_id);
```

### Seed Data

```sql
-- migrations/002_seed_data.sql

-- Insert Valorant game
INSERT INTO games (id, name, slug, is_active) VALUES
(1, 'Valorant', 'valorant', true);

-- Insert Valorant maps
INSERT INTO maps (game_id, name, slug, image_url, is_active, is_competitive) VALUES
(1, 'Abyss', 'abyss', '/images/abyss.png', true, true),
(1, 'Ascent', 'ascent', '/images/ascent.png', true, false),
(1, 'Bind', 'bind', '/images/bind.png', true, true),
(1, 'Breeze', 'breeze', '/images/breeze.png', true, false),
(1, 'Corrode', 'corrode', '/images/corrode.png', true, true),
(1, 'Fracture', 'fracture', '/images/fracture.png', true, false),
(1, 'Haven', 'haven', '/images/haven.png', true, true),
(1, 'Icebox', 'icebox', '/images/icebox.png', true, false),
(1, 'Lotus', 'lotus', '/images/lotus.png', true, false),
(1, 'Pearl', 'pearl', '/images/pearl.png', true, true),
(1, 'Split', 'split', '/images/split.png', true, true),
(1, 'Sunset', 'sunset', '/images/sunset.png', true, true);

-- Insert system map pools
INSERT INTO map_pools (game_id, name, type, is_system) VALUES
(1, 'All Maps', 'all', true),
(1, 'Competitive Maps', 'competitive', true);

-- Insert maps into "All Maps" pool
INSERT INTO map_pool_maps (map_pool_id, map_id)
SELECT 1, id FROM maps WHERE game_id = 1;

-- Insert maps into "Competitive Maps" pool
INSERT INTO map_pool_maps (map_pool_id, map_id)
SELECT 2, id FROM maps WHERE game_id = 1 AND is_competitive = true;
```

---

## Примечания

1. **Пока только Valorant**: Все сущности настроены на работу с одной игрой (Valorant), но архитектура позволяет легко добавить другие игры в будущем.

2. **Анонимные сессии**: Пользователь может создавать сессии вето без регистрации (UserID = null). Для комнат требуется авторизация.

3. **Share Token**: Каждая сессия имеет уникальный токен для публичного доступа.

4. **Комнаты**: 
   - Комнаты требуют авторизации для создания и участия
   - Уникальный код комнаты (6-8 символов) для приватных комнат
   - Один пользователь может быть только в одной комнате одновременно
   - WebSocket для real-time синхронизации состояния вето в комнате

5. **Soft Delete**: Все основные таблицы используют soft delete (deleted_at).

6. **Индексы**: Добавлены индексы для часто используемых полей для оптимизации запросов.

7. **Валидация**: Все DTO содержат теги для валидации через Gin binding.

8. **JWT**: Для авторизации используется JWT токен.

9. **Password Hashing**: Пароли хешируются с помощью bcrypt.

10. **WebSocket**: Для комнат используется WebSocket соединение для синхронизации состояния между участниками в реальном времени.
