# Конфигурация переменных окружения

## Общий .env файл

Проект использует **единый `.env` файл в корне проекта** для всех переменных окружения (backend и frontend).

## Структура

```
BBP/
├── .env              # Ваши переменные окружения (не коммитится в git)
├── .env.example      # Пример конфигурации (коммитится в git)
├── backend/
├── frontend/
└── docker-compose.yml
```

## Быстрый старт

### 1. Создание .env файла

Скрипты автоматически создадут `.env` из `.env.example`:

```bash
# Автоматически при запуске docker-start.sh или deploy.sh
./docker-start.sh

# Или вручную
cp .env.example .env
```

### 2. Редактирование .env

Откройте `.env` и настройте переменные под ваше окружение:

```bash
nano .env
# или
vim .env
```

## Переменные окружения

### Backend переменные (runtime)

Эти переменные используются во время выполнения backend:

| Переменная | Описание | По умолчанию | Обязательно |
|-----------|----------|--------------|-------------|
| `PORT` | Порт на котором запускается backend | `8080` | Нет |
| `DB_PATH` | Путь к файлу базы данных SQLite | `./data/app.db` | Нет |
| `JWT_SECRET` | Секретный ключ для подписи JWT токенов | - | ⚠️ **Да (в проде)** |
| `JWT_EXPIRY` | Время жизни JWT токена | `24h` | Нет |
| `CORS_ORIGIN` | Разрешенные origins для CORS (через запятую или `*`) | `*` | Нет |
| `ENVIRONMENT` | Окружение: `development` или `production` | `development` | Нет |

### Frontend переменные (build time)

⚠️ **ВАЖНО**: Эти переменные используются **только во время сборки** (build time). После сборки они "вшиваются" в код и не могут быть изменены без пересборки контейнера.

| Переменная | Описание | Пример для локальной разработки | Пример для production |
|-----------|----------|--------------------------------|----------------------|
| `VITE_API_URL` | URL API бэкенда | `http://localhost:8080/api` | `https://api.example.com/api` |
| `VITE_WS_URL` | URL WebSocket сервера | `ws://localhost:8080/ws` | `wss://api.example.com/ws` |

**Для изменения VITE переменных требуется пересборка frontend контейнера:**
```bash
docker compose up --build frontend
```

### Docker Compose переменные

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `COMPOSE_PROJECT_NAME` | Префикс для имен контейнеров | `bbp` |

## Примеры конфигурации

### Локальная разработка

```env
PORT=8080
DB_PATH=./data/app.db
JWT_SECRET=local-dev-secret-key
JWT_EXPIRY=24h
CORS_ORIGIN=*
ENVIRONMENT=development

VITE_API_URL=http://localhost:8080/api
VITE_WS_URL=ws://localhost:8080/ws
```

### Production (один домен через nginx)

```env
PORT=8080
DB_PATH=/app/data/app.db
JWT_SECRET=your-super-secret-production-key-min-32-chars
JWT_EXPIRY=24h
CORS_ORIGIN=https://ban.wise-dream.site
ENVIRONMENT=production

VITE_API_URL=/api
VITE_WS_URL=wss://ban.wise-dream.site/ws
```

### Production (разные домены)

```env
PORT=8080
DB_PATH=/app/data/app.db
JWT_SECRET=your-super-secret-production-key-min-32-chars
JWT_EXPIRY=24h
CORS_ORIGIN=https://app.example.com,https://www.example.com
ENVIRONMENT=production

VITE_API_URL=https://api.example.com/api
VITE_WS_URL=wss://api.example.com/ws
```

## Использование

### Docker Compose

`docker-compose.yml` автоматически загружает переменные из корневого `.env` файла:

```bash
docker compose up
```

### Backend (без Docker)

Скрипт `backend/start-backend.sh` автоматически загружает переменные из корневого `.env`:

```bash
cd backend
./start-backend.sh
```

Если `.env` в корне не найден, будет использоваться локальный `.env` в папке `backend` (обратная совместимость).

### Frontend (без Docker)

Frontend использует переменные из `.env` только во время сборки:

```bash
cd frontend
npm install
npm run build  # Vite читает VITE_* переменные из .env
```

## Важные замечания

1. ⚠️ **`.env` не коммитится в git** - добавлен в `.gitignore`
2. ⚠️ **JWT_SECRET обязательно измените в production** - используйте надежный случайный ключ (минимум 32 символа)
3. ⚠️ **VITE переменные требуют пересборки** - после изменения `VITE_API_URL` или `VITE_WS_URL` нужно пересобрать frontend
4. ✅ **Единый файл для всего проекта** - один `.env` в корне вместо отдельных файлов для backend и frontend

## Миграция со старых конфигураций

Если у вас были отдельные `.env` файлы в `backend/` и `frontend/`:

1. Создайте единый `.env` в корне проекта из `.env.example`
2. Скопируйте значения из старых файлов
3. Удалите старые `.env` файлы (они больше не нужны)
