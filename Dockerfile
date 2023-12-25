# Используйте официальный образ Go как базовый образ
FROM golang:1.21.4 as builder

# Установите рабочий каталог в контейнере
WORKDIR /app

# Скопируйте модули Go (go.mod и go.sum) в рабочий каталог
COPY go.mod go.sum ./

# Скачайте зависимости
RUN go mod download

# Скопируйте исходный код в контейнер
COPY . .

# Соберите приложение. Убедитесь, что путь к main.go отражает структуру вашего проекта.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp ./cmd

# Начните новый этап сборки с более легкого образа
FROM alpine:latest

RUN apk --no-cache add ca-certificates


WORKDIR /root/

# Скопируйте собранный бинарный файл из предыдущего этапа
COPY --from=builder /app/myapp .

# Откройте порт, который использует ваше приложение
EXPOSE 8082

# Запустите приложение
CMD ["./myapp"]
