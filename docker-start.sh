#!/bin/bash

# Скрипт для быстрого запуска Docker контейнеров
# Использование: ./docker-start.sh

set -e

echo "🚀 Запуск Docker контейнеров..."

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Ошибка: Docker не установлен. Установите Docker и повторите попытку."
    exit 1
fi

# Проверяем, что Docker работает
if ! docker info > /dev/null 2>&1; then
    echo "❌ Ошибка: Docker не запущен. Запустите Docker и повторите попытку."
    exit 1
fi

# Определяем команду docker compose (поддерживаем оба варианта)
# Делаем это ПОСЛЕ проверки, что Docker работает
if docker compose version > /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
elif docker-compose version > /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
else
    echo "❌ Ошибка: docker compose не установлен"
    echo "Установите Docker Compose и повторите попытку."
    exit 1
fi

# Проверяем наличие .env файла
if [ ! -f ".env" ]; then
    echo "⚠️  Файл .env не найден!"
    if [ -f ".env.example" ]; then
        echo "📝 Создаю .env из .env.example..."
        cp .env.example .env
        echo "✅ Файл .env создан. Отредактируйте его при необходимости."
    else
        echo "❌ Файл .env.example также не найден!"
        exit 1
    fi
fi

# Создаем директорию для БД, если её нет
mkdir -p backend/data

# Собираем и запускаем контейнеры
echo "📦 Сборка образов..."
$DOCKER_COMPOSE build

echo "🚀 Запуск контейнеров..."
$DOCKER_COMPOSE up -d

echo "⏳ Ожидание запуска контейнеров..."
sleep 5

# Проверяем статус
echo ""
echo "📊 Статус контейнеров:"
$DOCKER_COMPOSE ps

echo ""
echo "✅ Контейнеры запущены!"
echo ""
echo "🔍 Проверка health checks..."
sleep 3

# Проверяем backend health
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Backend доступен: http://localhost:8080/health"
else
    echo "⚠️  Backend еще не готов, проверьте логи: $DOCKER_COMPOSE logs backend"
fi

# Проверяем frontend
if curl -s http://localhost:80 > /dev/null 2>&1; then
    echo "✅ Frontend доступен: http://localhost:80"
else
    echo "⚠️  Frontend еще не готов, проверьте логи: $DOCKER_COMPOSE logs frontend"
fi

echo ""
echo "📝 Полезные команды:"
echo "  - Логи: $DOCKER_COMPOSE logs -f"
echo "  - Остановка: $DOCKER_COMPOSE down"
echo "  - Пересборка: $DOCKER_COMPOSE up --build -d"
echo ""
echo "🗄️  Не забудьте инициализировать БД (если первый запуск):"
echo "  docker exec -it bbp-backend /bin/sh -c 'cd /app && go run cmd/seed/main.go'"
echo ""
