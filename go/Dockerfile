# go-app/Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

# Установка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o main .

# Экспорт порта
EXPOSE 8000

# Запуск приложения
CMD ["./main"]