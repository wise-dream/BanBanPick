# Backend API

Backend для проекта MapBan на Go с использованием Clean Architecture.

## Технологии

- **Go** 1.24+
- **Gin** - веб-фреймворк
- **GORM** - ORM
- **SQLite** - база данных
- **JWT** - аутентификация
- **bcrypt** - хеширование паролей
- **gorilla/websocket** - WebSocket для комнат

## Структура проекта

```
backend/
├── cmd/
│   └── server/          # Точка входа приложения
├── internal/
│   ├── domain/          # Domain layer (entities, repository interfaces)
│   ├── usecase/         # Use cases (бизнес-логика)
│   ├── repository/      # Repository implementations
│   ├── handler/         # HTTP handlers
│   └── middleware/      # Middleware (auth, CORS, etc.)
├── pkg/                 # Вспомогательные пакеты
│   ├── database/        # Работа с БД
│   ├── jwt/             # JWT утилиты
│   └── password/        # Хеширование паролей
├── config/              # Конфигурация
└── data/                # SQLite база данных (не в git)
```

## Установка и запуск

### Требования

- Go 1.24 или выше
- SQLite (встроен в драйвер)

### Установка зависимостей

```bash
cd backend
go mod download
```

### Настройка окружения

Проект использует **единый `.env` файл в корне проекта** (не в папке `backend/`).

Создайте файл `.env` в корне проекта (можно скопировать из `.env.example`):

```bash
# Из корня проекта
cp .env.example .env
```

Скрипт `start-backend.sh` автоматически загрузит переменные из корневого `.env`.

Или установите переменные окружения:

```bash
export PORT=8080
export JWT_SECRET=your-secret-key
export DB_PATH=./data/app.db
export CORS_ORIGIN=http://localhost:5173
```

### Запуск сервера

**Рекомендуемый способ (с автоматической инициализацией БД):**

```bash
./start-backend.sh
```

Скрипт автоматически:
- Проверяет наличие БД, при отсутствии выполняет инициализацию
- Загружает переменные окружения из `.env` (если файл существует)
- Запускает сервер

**Альтернативный способ (вручную):**

```bash
go run cmd/server/main.go
```

Или скомпилируйте и запустите:

```bash
go build -o bin/server ./cmd/server
./bin/server
```

Сервер запустится на `http://localhost:8080` (или порт из переменной окружения `PORT`)

### Проверка работы

```bash
curl http://localhost:8080/health
```

Должен вернуть:
```json
{
  "status": "ok",
  "message": "Server is running"
}
```

### Инициализация базы данных

Для инициализации БД (миграции + заполнение начальными данными) используйте скрипт:

```bash
./init-db.sh
```

Скрипт автоматически:
- Выполняет миграции БД
- Заполняет БД начальными данными (seed)

Создается:
- Игра Valorant (id=1)
- 12 карт Valorant
- 2 системных пула: "All Maps" и "Competitive Maps"

Команда идемпотентна - можно запускать несколько раз без дублирования данных.

**Альтернативный способ (вручную):**
```bash
go run cmd/seed/main.go
```

## База данных

SQLite база данных создается автоматически при первом запуске в файле `data/app.db`.

Миграции выполняются автоматически при старте сервера через GORM AutoMigrate.

## API Endpoints

API endpoints будут добавлены по мере реализации задач. См. `BACKEND_TASKS.md` для детального плана.

### Текущие endpoints:

- `GET /health` - проверка работоспособности сервера
- `GET /api/` - базовая информация об API

## Разработка

### Структура по Clean Architecture:

1. **Domain Layer** (`internal/domain/`)
   - Entities - чистые структуры без зависимостей
   - Repository interfaces - контракты для работы с данными

2. **Use Case Layer** (`internal/usecase/`)
   - Бизнес-логика приложения
   - Не зависит от фреймворков

3. **Repository Layer** (`internal/repository/`)
   - Реализация интерфейсов репозиториев
   - Работа с БД через GORM

4. **Handler Layer** (`internal/handler/`)
   - HTTP handlers для Gin
   - DTOs для запросов/ответов

## Задачи

Детальный список задач находится в `BACKEND_TASKS.md`.

Текущий статус:
- ✅ Задача 1: Инициализация Go проекта
- ✅ Задача 2: Настройка SQLite базы данных
- ✅ Задача 3: Domain entities
- ✅ Задача 4: Repository interfaces
- ✅ Задача 5: GORM модели и реализации репозиториев
- ⏳ Задача 6: Seed данные (следующая)
