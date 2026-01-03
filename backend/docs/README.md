# API Документация

## Swagger/OpenAPI

API документация доступна в формате OpenAPI 3.0 в файле `swagger.yaml`.

### Использование

1. **Swagger UI:**
   - Установите Swagger UI: `npm install -g swagger-ui-serve`
   - Запустите: `swagger-ui-serve swagger.yaml`
   - Откройте в браузере: `http://localhost:3000`

2. **Redoc:**
   - Установите Redoc: `npm install -g redoc-cli`
   - Запустите: `redoc-cli serve swagger.yaml`
   - Откройте в браузере: `http://localhost:8080`

3. **Онлайн редактор:**
   - Откройте файл в [Swagger Editor](https://editor.swagger.io/)

### Основные endpoints

#### Авторизация
- `POST /api/auth/register` - Регистрация
- `POST /api/auth/login` - Вход
- `GET /api/auth/me` - Текущий пользователь

#### Пользователи
- `GET /api/users/profile` - Профиль
- `PUT /api/users/profile` - Обновление профиля
- `GET /api/users/sessions` - Сессии пользователя
- `GET /api/users/rooms` - Комнаты пользователя

#### Veto Sessions
- `POST /api/veto/sessions` - Создать сессию
- `GET /api/veto/sessions/:id` - Получить сессию
- `POST /api/veto/sessions/:id/ban` - Забанить карту
- `POST /api/veto/sessions/:id/pick` - Выбрать карту
- `POST /api/veto/sessions/:id/reset` - Сбросить сессию

#### Map Pools
- `GET /api/games/:gameId/map-pools` - Список пулов
- `GET /api/map-pools/:id` - Получить пул
- `POST /api/map-pools` - Создать кастомный пул
- `DELETE /api/map-pools/:id` - Удалить пул

#### Rooms
- `GET /api/rooms` - Список комнат
- `POST /api/rooms` - Создать комнату
- `GET /api/rooms/:id` - Получить комнату
- `POST /api/rooms/:id/join` - Присоединиться
- `POST /api/rooms/:id/leave` - Выйти
- `DELETE /api/rooms/:id` - Удалить комнату

#### WebSocket
- `WS /ws/room/:roomId` - WebSocket для комнаты

### Аутентификация

Большинство endpoints требуют JWT токен в заголовке:
```
Authorization: Bearer <token>
```

Токен получается при регистрации или входе.

### Коды ответов

- `200` - Успешно
- `201` - Создано
- `400` - Невалидные данные
- `401` - Не авторизован
- `403` - Нет прав
- `404` - Не найдено
- `409` - Конфликт (дубликат)
- `429` - Слишком много запросов (rate limit)
- `500` - Внутренняя ошибка сервера

### Примеры запросов

#### Регистрация
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "password123"
  }'
```

#### Вход
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Создание комнаты
```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "My Room",
    "type": "public",
    "game_id": 1,
    "max_participants": 10
  }'
```
