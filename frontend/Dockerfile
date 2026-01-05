# Multi-stage build для frontend
# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app

# Копируем package файлы
COPY package*.json ./
RUN npm ci

# Копируем исходный код
COPY . .

# Аргументы сборки для переменных окружения Vite (build-time)
ARG VITE_API_URL
ARG VITE_WS_URL

# Устанавливаем переменные окружения для сборки
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_WS_URL=${VITE_WS_URL}

# Собираем приложение
RUN npm run build

# Stage 2: Production (Nginx)
FROM nginx:alpine

# Копируем собранные файлы из builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Копируем конфигурацию nginx (опционально, для SPA роутинга)
RUN echo 'server { \
    listen 80; \
    server_name _; \
    root /usr/share/nginx/html; \
    index index.html; \
    \
    # Gzip compression \
    gzip on; \
    gzip_vary on; \
    gzip_min_length 1024; \
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/javascript application/json; \
    \
    # SPA routing - все запросы перенаправляем на index.html \
    location / { \
        try_files $uri $uri/ /index.html; \
    } \
    \
    # Кеширование статических ресурсов \
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2|ttf|eot)$ { \
        expires 1y; \
        add_header Cache-Control "public, immutable"; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
