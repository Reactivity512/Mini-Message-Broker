# Этап 1: Сборка
FROM golang:1.21-alpine AS builder

# Устанавливаем зависимости для сборки (включая git для go mod)
RUN apk add --no-cache git ca-certificates && \
    # Добавляем пользователя для корректных прав в финальном образе
    adduser -D -g '' appuser

WORKDIR /workspace

# Копируем файлы зависимостей для эффективного кеширования слоев Docker
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Устанавливаем protoc и генерируем код.
RUN apk add --no-cache protobuf-dev && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    export PATH="$PATH:$(go env GOPATH)/bin" && \
    protoc --proto_path=api/proto \
           --go_out=internal/delivery/grpc/pb \
           --go_opt=paths=source_relative \
           --go-grpc_out=internal/delivery/grpc/pb \
           --go-grpc_opt=paths=source_relative \
           api/proto/broker.proto

# Собираем статически скомпилированный бинарный файл
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o broker-server ./cmd/server

# Этап 2: Финальный образ
FROM alpine:latest

# Копируем сертификаты из builder для работы с TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем информацию о пользователе из builder
COPY --from=builder /etc/passwd /etc/passwd

# Создаем непривилегированного пользователя RUN adduser -D -g '' appuser
USER appuser

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /workspace/broker-server .

# Копируем конфигурационный файл
COPY --from=builder /workspace/config.yaml .

# Открываем порт, который использует сервер (по умолчанию 50051)
EXPOSE 50051

# Запускаем сервер
CMD ["./broker-server"]
# Если конфиг не нужен по умолчанию, можно просто:
# CMD ["./broker-server"]
# Или с конфигом:
# CMD ["./broker-server", "config.yaml"]
