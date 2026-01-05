# Multi-stage build для backend
# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Устанавливаем необходимые пакеты (для SQLite нужен C runtime)
RUN apk --no-cache add ca-certificates sqlite

# Копируем бинарник из builder stage
COPY --from=builder /app/server .

# Создаем директорию для базы данных
RUN mkdir -p /app/data

# Открываем порт
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запускаем приложение
CMD ["./server"]
