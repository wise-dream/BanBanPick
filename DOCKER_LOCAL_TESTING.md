# Локальное тестирование Docker

Эта инструкция поможет вам протестировать Docker контейнеры локально перед деплоем на сервер.

## Предварительные требования

1. **Docker** установлен и запущен:
   ```bash
   docker --version
   docker compose version
   ```
   
   > **Примечание**: Используется `docker compose` (Docker Compose V2). Если у вас старая версия, может быть `docker-compose` (через дефис). Скрипты автоматически определяют, какой вариант использовать.

2. **Docker Desktop** (для Windows/Mac) или **Docker Engine** (для Linux)

## Быстрый старт

### 1. Клонируйте репозиторий (если еще не клонировали)

```bash
git clone <your-repo-url>
cd BBP
```

### 2. Подготовьте переменные окружения

Создайте файл `.env` в корне проекта:

```bash
# Автоматически (из .env.example)
./setup-env.sh

# Или вручную
cp .env.example .env
```

Отредактируйте `.env` файл и настройте переменные:
- **JWT_SECRET** - обязательно измените на уникальный ключ для production!
- **VITE_API_URL** - URL для API запросов (для локальной разработки: `http://localhost:8080/api`)
- **VITE_WS_URL** - URL для WebSocket соединений (для локальной разработки: `ws://localhost:8080/ws`)
- **CORS_ORIGIN** - разрешенные origins для CORS (для локальной разработки: `*` или конкретные домены)

**Важно**: VITE переменные используются только во время сборки (build time), поэтому после изменения `.env` нужно пересобрать frontend контейнер:
```bash
docker compose build frontend
docker compose up -d frontend
```

### 3. Соберите и запустите контейнеры

```bash
# Собрать и запустить все сервисы
docker compose up --build

# Или в фоновом режиме
docker compose up -d --build
```

### 4. Инициализируйте базу данных (первый запуск)

После первого запуска контейнера backend нужно инициализировать БД:

```bash
# Войти в контейнер backend
docker exec -it bbp-backend sh

# Или выполнить команду напрямую
docker exec -it bbp-backend /bin/sh -c "cd /app && go run cmd/seed/main.go"
```

**Альтернативный способ** (если у вас установлен Go локально):
```bash
cd backend
DB_PATH=./data/app.db go run cmd/seed/main.go
```

### 5. Проверьте работу

- **Frontend**: http://localhost:80 (или http://localhost)
- **Backend API**: http://localhost:8080/api
- **Health check**: http://localhost:8080/health
- **Backend health**: http://localhost:8080/health

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

### Остановка контейнеров

```bash
# Остановить и удалить контейнеры
docker compose down

# Остановить и удалить контейнеры + volumes (удалит БД!)
docker compose down -v
```

### Пересборка после изменений

```bash
# Пересобрать и перезапустить
docker compose up --build -d

# Пересобрать только backend
docker compose build backend
docker compose up -d backend

# Пересобрать только frontend
docker compose build frontend
docker compose up -d frontend
```

### Проверка статуса

```bash
# Статус контейнеров
docker compose ps

# Health check статус
docker ps  # смотрим STATUS колонку
```

### Доступ к контейнерам

```bash
# Backend shell
docker exec -it bbp-backend sh

# Frontend shell
docker exec -it bbp-frontend sh

# Проверить файлы в контейнере
docker exec -it bbp-backend ls -la /app/data
```

## Тестирование с nginx (как на сервере)

Если хотите протестировать конфигурацию с nginx (как будет на продакшене):

### Вариант 1: Использовать nginx на хосте

1. Установите nginx локально (если еще не установлен):
   ```bash
   # Ubuntu/Debian
   sudo apt-get install nginx
   
   # MacOS
   brew install nginx
   ```

2. Скопируйте конфигурацию:
   ```bash
   sudo cp nginx-example.conf /etc/nginx/sites-available/bbp-local
   sudo ln -s /etc/nginx/sites-available/bbp-local /etc/nginx/sites-enabled/
   ```

3. Отредактируйте конфигурацию:
   ```bash
   sudo nano /etc/nginx/sites-available/bbp-local
   ```
   
   Измените:
   - `server_name localhost;` (вместо your-domain.com)
   - Убедитесь что upstream указывает на правильные порты (8080 для backend)

4. Проверьте и перезагрузите nginx:
   ```bash
   sudo nginx -t
   sudo systemctl reload nginx  # или sudo nginx -s reload
   ```

5. Запустите Docker контейнеры:
   ```bash
   docker compose up -d
   ```

6. Откройте в браузере:
   - http://localhost (через nginx)

### Вариант 2: Использовать nginx в Docker (через docker compose)

Можно добавить nginx сервис в docker compose.yml, но для локального тестирования проще использовать nginx на хосте.

## Отладка проблем

### Backend не запускается

1. Проверьте логи:
   ```bash
   docker compose logs backend
   ```

2. Проверьте, что порт 8080 свободен:
   ```bash
   lsof -i :8080  # MacOS/Linux
   netstat -ano | findstr :8080  # Windows
   ```

3. Проверьте переменные окружения:
   ```bash
   docker compose config
   ```

### Frontend не собирается

1. Проверьте логи сборки:
   ```bash
   docker compose logs frontend
   ```

2. Проверьте, что все зависимости установлены:
   ```bash
   docker compose build --no-cache frontend
   ```

### База данных не инициализируется

1. Проверьте, что директория data существует:
   ```bash
   ls -la backend/data/
   ```

2. Попробуйте инициализировать вручную:
   ```bash
   docker exec -it bbp-backend sh
   cd /app
   go run cmd/seed/main.go
   ```

### Проблемы с WebSocket

1. Проверьте, что WebSocket endpoint доступен:
   ```bash
   curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
        -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" \
        http://localhost:8080/ws
   ```

2. Если используете nginx - проверьте конфигурацию WebSocket:
   - Должны быть заголовки `Upgrade` и `Connection`
   - Увеличенные таймауты

## Тестирование перед деплоем на сервер

1. ✅ Убедитесь, что все контейнеры запускаются:
   ```bash
   docker compose up -d
   docker compose ps
   ```

2. ✅ Проверьте health checks:
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:80/
   ```

3. ✅ Протестируйте основную функциональность:
   - Регистрация/авторизация
   - Создание комнаты
   - WebSocket соединения
   - Процесс вето

4. ✅ Проверьте логи на ошибки:
   ```bash
   docker compose logs | grep -i error
   ```

5. ✅ Если используете nginx локально - проверьте через nginx:
   - http://localhost (через nginx)
   - Все запросы должны проходить через nginx

## Следующие шаги

После успешного локального тестирования:

1. Создайте production конфигурацию (`.env.production`)
2. Настройте CI/CD pipeline
3. Деплой на сервер (см. `DEPLOYMENT.md`)
