#!/bin/bash

# Скрипт для полного развертывания backend и frontend через Docker
# Использование: ./deploy.sh

set -e

echo "🚀 Развертывание BBP (Backend + Frontend)"
echo "=========================================="
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Функция для проверки команды
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}❌ Ошибка: $1 не установлен${NC}"
        echo "Установите $1 и повторите попытку."
        exit 1
    fi
}

# Проверяем необходимые команды
echo "📋 Проверка зависимостей..."
check_command docker

# Проверяем, что Docker работает
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Ошибка: Docker не запущен${NC}"
    echo "Запустите Docker и повторите попытку."
    exit 1
fi

# Определяем команду docker compose (поддерживаем оба варианта)
# Делаем это ПОСЛЕ проверки, что Docker работает
if docker compose version > /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
elif docker-compose version > /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
else
    echo -e "${RED}❌ Ошибка: docker compose не установлен${NC}"
    echo "Установите Docker Compose и повторите попытку."
    exit 1
fi

echo -e "${GREEN}✅ Docker доступен${NC}"
echo -e "${GREEN}✅ Docker Compose доступен ($DOCKER_COMPOSE)${NC}"
echo ""

# Проверяем наличие .env файла
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}⚠️  Файл .env не найден${NC}"
    if [ -f ".env.example" ]; then
        echo "📝 Создаю .env из .env.example..."
        cp .env.example .env
        echo -e "${GREEN}✅ Файл .env создан${NC}"
        echo ""
        echo -e "${YELLOW}⚠️  ВАЖНО: Отредактируйте .env файл перед использованием в production!${NC}"
        echo "   Особенно измените JWT_SECRET на уникальный ключ!"
        echo ""
        read -p "Продолжить с настройками по умолчанию? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Отменено. Отредактируйте .env и запустите скрипт снова."
            exit 0
        fi
    else
        echo -e "${RED}❌ Ошибка: файл .env.example не найден!${NC}"
        exit 1
    fi
fi

# Создаем необходимые директории
echo "📁 Создание директорий..."
mkdir -p backend/data
echo -e "${GREEN}✅ Директории созданы${NC}"
echo ""

# Останавливаем старые контейнеры (если есть)
echo "🛑 Остановка старых контейнеров (если есть)..."
$DOCKER_COMPOSE down 2>/dev/null || true
echo ""

# Собираем образы
echo "🔨 Сборка Docker образов..."
echo "   Backend..."
$DOCKER_COMPOSE build backend
echo "   Frontend..."
$DOCKER_COMPOSE build frontend
echo -e "${GREEN}✅ Образы собраны${NC}"
echo ""

# Запускаем контейнеры
echo "🚀 Запуск контейнеров..."
$DOCKER_COMPOSE up -d
echo -e "${GREEN}✅ Контейнеры запущены${NC}"
echo ""

# Ждем запуска
echo "⏳ Ожидание запуска сервисов..."
sleep 5

# Проверяем статус
echo "📊 Статус контейнеров:"
$DOCKER_COMPOSE ps
echo ""

# Проверяем health checks
echo "🔍 Проверка health checks..."

# Backend
BACKEND_HEALTH=false
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        BACKEND_HEALTH=true
        break
    fi
    sleep 1
done

if [ "$BACKEND_HEALTH" = true ]; then
    echo -e "${GREEN}✅ Backend доступен: http://localhost:8080/health${NC}"
else
    echo -e "${YELLOW}⚠️  Backend еще не готов${NC}"
    echo "   Проверьте логи: $DOCKER_COMPOSE logs backend"
fi

# Frontend
FRONTEND_HEALTH=false
for i in {1..30}; do
    if curl -s http://localhost:80 > /dev/null 2>&1; then
        FRONTEND_HEALTH=true
        break
    fi
    sleep 1
done

if [ "$FRONTEND_HEALTH" = true ]; then
    echo -e "${GREEN}✅ Frontend доступен: http://localhost:80${NC}"
else
    echo -e "${YELLOW}⚠️  Frontend еще не готов${NC}"
    echo "   Проверьте логи: $DOCKER_COMPOSE logs frontend"
fi

echo ""

# Проверяем, нужно ли инициализировать БД
echo "🗄️  Проверка базы данных..."
DB_EXISTS=$(docker exec bbp-backend sh -c "test -f /app/data/app.db && echo 'yes' || echo 'no'" 2>/dev/null || echo "no")

if [ "$DB_EXISTS" != "yes" ]; then
    echo -e "${YELLOW}⚠️  База данных не найдена${NC}"
    echo "📦 Инициализация базы данных..."
    if docker exec -it bbp-backend /bin/sh -c "cd /app && go run cmd/seed/main.go" 2>/dev/null; then
        echo -e "${GREEN}✅ База данных инициализирована${NC}"
    else
        echo -e "${RED}❌ Ошибка при инициализации базы данных${NC}"
        echo "   Попробуйте вручную:"
        echo "   docker exec -it bbp-backend /bin/sh -c 'cd /app && go run cmd/seed/main.go'"
    fi
else
    echo -e "${GREEN}✅ База данных уже существует${NC}"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}✅ Развертывание завершено!${NC}"
echo ""
echo "📝 Полезные команды:"
echo "   - Логи всех сервисов:  $DOCKER_COMPOSE logs -f"
echo "   - Логи backend:        $DOCKER_COMPOSE logs -f backend"
echo "   - Логи frontend:       $DOCKER_COMPOSE logs -f frontend"
echo "   - Остановка:           $DOCKER_COMPOSE down"
echo "   - Перезапуск:          $DOCKER_COMPOSE restart"
echo "   - Пересборка:          $DOCKER_COMPOSE up --build -d"
echo ""
echo "🌐 Доступ к приложению:"
echo "   - Frontend:  http://localhost:80"
echo "   - Backend:   http://localhost:8080/api"
echo "   - Health:    http://localhost:8080/health"
echo ""
