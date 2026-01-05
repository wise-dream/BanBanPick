# Развертывание через Docker

Быстрое развертывание backend и frontend через Docker.

## Быстрый старт

### 1. Клонируйте репозиторий

```bash
git clone <your-repo-url>
cd BBP
```

### 2. Разверните всё одной командой

```bash
./deploy.sh
```

Скрипт `deploy.sh` автоматически:
- ✅ Проверит наличие Docker
- ✅ Создаст `.env` из `.env.example` (если нужно)
- ✅ Соберет Docker образы
- ✅ Запустит контейнеры (backend + frontend)
- ✅ Проверит health checks
- ✅ Инициализирует БД (если первый запуск)

### 3. Готово!

- **Frontend**: http://localhost:80
- **Backend API**: http://localhost:8080/api
- **Health check**: http://localhost:8080/health

## Настройка переменных окружения

### Для локальной разработки

Файл `.env.example` уже настроен для локальной разработки:

```bash
# Просто скопируйте (или используйте deploy.sh, он сделает это автоматически)
cp .env.example .env
```

Настройки по умолчанию:
- `VITE_API_URL=http://localhost:8080/api`
- `VITE_WS_URL=ws://localhost:8080/ws`
- `CORS_ORIGIN=*` (разрешает все origins)

### Для production на сервере

Отредактируйте `.env`:

```env
# Backend
ENVIRONMENT=production
JWT_SECRET=<сгенерируйте уникальный ключ минимум 32 символа>
CORS_ORIGIN=https://your-domain.com

# Frontend (если через nginx на одном домене)
VITE_API_URL=/api
VITE_WS_URL=wss://your-domain.com/ws

# Frontend (если на разных доменах)
# VITE_API_URL=https://api.your-domain.com/api
# VITE_WS_URL=wss://api.your-domain.com/ws
```

**Важно**: После изменения `VITE_*` переменных нужно пересобрать frontend:

```bash
docker compose build frontend
docker compose up -d frontend
```

> **Примечание**: Используется `docker compose` (новый синтаксис Docker Compose V2). Если у вас старая версия, можно использовать `docker-compose` (через дефис).

## Полезные команды

### Просмотр логов

```bash
# Все сервисы
docker compose logs -f

# Только backend
docker compose logs -f backend

# Только frontend
docker compose logs -f frontend
```

### Управление контейнерами

```bash
# Остановить
docker compose down

# Остановить и удалить volumes (удалит БД!)
docker compose down -v

# Перезапустить
docker compose restart

# Пересобрать и перезапустить
docker compose up --build -d
```

### Инициализация/переинициализация БД

```bash
# Если нужно пересоздать БД
docker compose down -v
docker compose up -d backend
docker exec -it bbp-backend /bin/sh -c 'cd /app && go run cmd/seed/main.go'
```

### Статус контейнеров

```bash
# Статус
docker compose ps

# Health check статус
docker ps
```

## Скрипты

- **`deploy.sh`** - Полное развертывание (рекомендуется)
  - Проверяет зависимости
  - Создает .env если нужно
  - Собирает и запускает всё
  - Проверяет health checks
  - Инициализирует БД

- **`docker-start.sh`** - Быстрый запуск (если уже настроено)
  - Собирает и запускает контейнеры
  - Не инициализирует БД автоматически

- **`setup-env.sh`** - Создание .env из .env.example

## Интеграция с nginx на сервере

Если на сервере уже установлен nginx, используйте `nginx-example.conf`:

1. Скопируйте конфигурацию:
   ```bash
   sudo cp nginx-example.conf /etc/nginx/sites-available/bbp
   ```

2. Отредактируйте:
   ```bash
   sudo nano /etc/nginx/sites-available/bbp
   ```
   - Измените `server_name` на ваш домен
   - Проверьте upstream (должны указывать на Docker контейнеры)

3. Активируйте:
   ```bash
   sudo ln -s /etc/nginx/sites-available/bbp /etc/nginx/sites-enabled/
   sudo nginx -t
   sudo systemctl reload nginx
   ```

Подробнее: см. `DOCKER_LOCAL_TESTING.md` и `nginx-example.conf`

## Структура файлов

```
BBP/
├── .env                    # Переменные окружения (не коммитится)
├── .env.example            # Пример переменных окружения
├── docker-compose.yml      # Конфигурация Docker Compose
├── deploy.sh               # Скрипт развертывания (основной)
├── docker-start.sh         # Быстрый запуск
├── setup-env.sh            # Создание .env
├── nginx-example.conf      # Пример конфигурации nginx
├── backend/
│   ├── Dockerfile          # Backend образ
│   └── data/               # База данных (монтируется в контейнер)
└── frontend/
    └── Dockerfile          # Frontend образ
```

## Решение проблем

### Контейнеры не запускаются

1. Проверьте логи:
   ```bash
   docker compose logs
   ```

2. Проверьте, что порты свободны:
   ```bash
   lsof -i :80    # Frontend
   lsof -i :8080  # Backend
   ```

3. Проверьте .env файл:
   ```bash
   cat .env
   ```

### Backend не отвечает

1. Проверьте логи:
   ```bash
   docker compose logs backend
   ```

2. Проверьте health check:
   ```bash
   curl http://localhost:8080/health
   ```

3. Проверьте БД (если нужно, инициализируйте):
   ```bash
   docker exec -it bbp-backend /bin/sh -c 'cd /app && go run cmd/seed/main.go'
   ```

### Frontend не собирается

1. Проверьте переменные окружения VITE_* в .env
2. Пересоберите:
   ```bash
   docker compose build --no-cache frontend
   ```

### WebSocket не работает

1. Проверьте VITE_WS_URL в .env
2. Убедитесь, что используете правильный протокол (ws:// для HTTP, wss:// для HTTPS)
3. Если через nginx - проверьте конфигурацию WebSocket в nginx-example.conf
