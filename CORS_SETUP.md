# Настройка CORS

## Краткий ответ

**CORS настраивается на БЭКЕНДЕ** (на фронтенде менять ничего не нужно).

Браузер автоматически блокирует запросы между разными доменами (локальный фронтенд → прод бэкенд) для безопасности. Чтобы разрешить такие запросы, нужно настроить CORS на сервере.

---

## Что было исправлено

1. ✅ **Поддержка нескольких origins** - теперь можно указать несколько доменов через запятую
2. ✅ **Автоматическое разрешение localhost** - по умолчанию разрешены популярные порты для разработки
3. ✅ **Безопасная конфигурация** - по умолчанию разрешены только нужные origins

---

## Настройка на проде

### Вариант 1: Через переменную окружения (рекомендуется)

На сервере установите переменную окружения:

```bash
export CORS_ORIGIN="http://localhost:5173,https://ban.wise-dream.site,http://localhost:3000"
```

Или в файле `.env` или системных переменных окружения:

```env
CORS_ORIGIN=http://localhost:5173,https://ban.wise-dream.site,http://localhost:3000
```

### Вариант 2: Использовать значения по умолчанию

Если не установлена `CORS_ORIGIN`, по умолчанию разрешены:
- `http://localhost:5173` (Vite dev server)
- `https://ban.wise-dream.site` (прод домен)

Также автоматически разрешены:
- `http://localhost:3000`
- `http://127.0.0.1:5173`
- `http://127.0.0.1:3000`

### Вариант 3: Разрешить все origins (только для разработки!)

⚠️ **ВНИМАНИЕ: Не используйте в проде!**

```bash
export CORS_ORIGIN="*"
```

---

## Как это работает

1. **Браузер** делает запрос с заголовком `Origin: http://localhost:5173`
2. **Сервер** проверяет, есть ли этот origin в списке разрешенных
3. **Сервер** отвечает заголовком `Access-Control-Allow-Origin: http://localhost:5173`
4. **Браузер** разрешает запрос

---

## Настройка для разных сценариев

### Локальный фронтенд → Прод бэкенд

```bash
# На проде установите:
export CORS_ORIGIN="http://localhost:5173,https://ban.wise-dream.site"
```

### Локальный фронтенд → Локальный бэкенд

```bash
# Не нужно ничего менять - localhost уже разрешен по умолчанию
# Или установите явно:
export CORS_ORIGIN="http://localhost:5173"
```

### Прод фронтенд → Прод бэкенд

```bash
# На проде установите только прод домен:
export CORS_ORIGIN="https://ban.wise-dream.site"
```

### Несколько фронтендов

```bash
# Разрешите все нужные домены:
export CORS_ORIGIN="http://localhost:5173,https://ban.wise-dream.site,https://staging.example.com"
```

---

## Перезапуск сервера

После изменения переменной окружения **необходимо перезапустить бэкенд**:

```bash
# Если используете systemd
sudo systemctl restart bbp-backend

# Если запускаете вручную
./start-backend.sh

# Или просто перезапустите процесс
```

---

## Проверка работы

### 1. Проверка CORS заголовков

Запустите запрос с локального фронтенда и проверьте заголовки ответа:

```bash
curl -H "Origin: http://localhost:5173" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     https://ban.wise-dream.site/api/auth/login \
     -v
```

Должны быть заголовки:
```
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization
Access-Control-Allow-Credentials: true
```

### 2. Проверка в браузере

Откройте консоль браузера (F12) и попробуйте сделать запрос. Если CORS настроен правильно, ошибок быть не должно.

Если видите ошибку типа:
```
Access to fetch at 'https://ban.wise-dream.site/api/...' from origin 'http://localhost:5173' 
has been blocked by CORS policy
```

Это значит, что на сервере не настроен CORS для вашего origin.

---

## Где находятся файлы настроек

- **CORS middleware**: `backend/internal/middleware/cors.go`
- **Конфигурация**: `backend/config/config.go`
- **Использование**: `backend/cmd/server/main.go` (строка 62)

---

## Важные моменты

1. ✅ **CORS настраивается ТОЛЬКО на бэкенде** - фронтенд ничего не делает
2. ✅ **Несколько origins через запятую** - поддерживается список доменов
3. ✅ **Credentials включены** - `Access-Control-Allow-Credentials: true` (нужно для отправки cookies/JWT)
4. ✅ **Preflight кеширование** - OPTIONS запросы кешируются на 12 часов
5. ⚠️ **Не используйте `*` в проде** - это небезопасно, если нужны credentials

---

## Troubleshooting

### Проблема: CORS ошибка все еще возникает

**Решение:**
1. Проверьте, что переменная окружения `CORS_ORIGIN` установлена правильно
2. Убедитесь, что бэкенд перезапущен после изменения
3. Проверьте точное значение `Origin` в запросе (проверьте в DevTools → Network → Headers)

### Проблема: Credentials не работают

**Решение:**
- Убедитесь, что используете конкретный origin, а не `*`
- Проверьте, что `Access-Control-Allow-Credentials: true` в ответе

### Проблема: WebSocket подключение не работает

**Решение:**
- WebSocket использует тот же CORS middleware
- Проверьте, что origin фронтенда разрешен
- Проверьте, что токен передается правильно в query параметре

---

## Пример полной настройки

### На проде (сервер)

```bash
# В .env или системных переменных
export CORS_ORIGIN="http://localhost:5173,https://ban.wise-dream.site"
export PORT=8080
export ENVIRONMENT=production
export JWT_SECRET=your-secret-key
```

### На локальной машине (фронтенд)

В `.env` или `.env.local`:

```env
VITE_API_URL=https://ban.wise-dream.site
VITE_WS_URL=wss://ban.wise-dream.site
```

**Готово!** Теперь локальный фронтенд может обращаться к продовому бэкенду без CORS ошибок.
