# Mini Message Broker

Упрощённый брокер сообщений в стиле Kafka: топики, очереди, подписки, гарантии доставки **at-most-once** и **at-least-once**. API по gRPC.

## Запуск

```bash
# Зависимости
go mod download

# Один раз сгенерировать Go-код из .proto (обязательно для работы grpcurl)
task proto          # если protoc уже установлен (brew/apt/winget)
task proto:win      # Windows: скачает protoc и сгенерирует код

# Запуск сервера (порт 50051 по умолчанию)
go run ./cmd/server

# Или с конфигом
go run ./cmd/server config.yaml

# Через Taskfile
task run
```

Конфиг опционально: при его отсутствии используются значения по умолчанию (см. `config.yaml`).

В консоли сервера выводятся логи каждого gRPC-запроса: метод, адрес клиента и результат (ok/error).

### Ошибка «target server does not expose service "broker.Broker"»

Она возникает, если вызывать grpcurl **до** генерации кода из proto. Серверу нужны сгенерированные типы для gRPC reflection.

**Что сделать:** один раз выполнить генерацию, затем перезапустить сервер.

- **Windows:** `task proto:win` (скачает protoc в `.tools/protoc` и сгенерирует файлы в `internal/delivery/grpc/pb/`).
- **Linux/macOS:** установить [protoc](https://github.com/protocolbuffers/protobuf/releases) и выполнить `task proto`.

После этого команды grpcurl из README будут работать без флага `-proto`.

---

## Как пользоваться брокером

### 1. Создать топик

Топик — именованный поток сообщений (аналог топика в Kafka).

**gRPC:** `CreateTopic(CreateTopicRequest) → CreateTopicResponse`

- `name` — имя топика (обязательно)
- `retention_messages` — максимальное число сообщений в очереди (опционально, по умолчанию можно задать 10000)

После создания автоматически создаётся очередь с идентификатором `"0"`. Дополнительные очереди можно создать через `CreateQueue`.

**Пример (grpcurl):**

```bash
# Linux/macOS или PowerShell (прямые кавычки)
grpcurl -plaintext -d '{"name": "orders", "retention_messages": 5000}' localhost:50051 broker.Broker/CreateTopic
```

На **Windows (cmd.exe)** кавычки в `-d` часто ломают JSON — появляется ошибка `invalid character 'n' looking for beginning of object key string`. В cmd.exe экранируйте двойные кавычки: `-d "{\"name\": \"orders\", \"retention_messages\": 10000}"`

Ответ:

```json
{"name": "orders", "retention_messages": 5000}
```

---

### 2. Создать очередь (опционально)

Очередь привязана к топику (аналог партиции). По умолчанию у топика уже есть очередь `"0"`. Нужна ещё одна — создайте через `CreateQueue`.

**gRPC:** `CreateQueue(CreateQueueRequest) → CreateQueueResponse`

- `topic_name` — имя топика
- `queue_id` — идентификатор очереди (например `"0"`, `"1"`, `"2"`)

```bash
grpcurl -plaintext -d '{"topic_name": "orders", "queue_id": "1"}' localhost:50051 broker.Broker/CreateQueue
```

---

### 3. Подписаться на топик

Подписка = потребитель (consumer group) + топик + очередь + гарантия доставки.

**gRPC:** `Subscribe(SubscribeRequest) → SubscribeResponse`

- `topic_name` — топик
- `queue_id` — очередь (можно `""` или `"0"` для дефолтной)
- `consumer_group` — имя группы потребителей (уникально в рамках топика)
- `delivery_guarantee` — гарантия доставки:
  - `AT_MOST_ONCE` (1) — сообщение может быть доставлено не более одного раза (без ack)
  - `AT_LEAST_ONCE` (2) — сообщение будет доставлено минимум один раз; нужно вызывать `Ack` после обработки

В ответе приходит `subscription_id` — он нужен для `Consume` и `Ack`.

**Пример — at-most-once:**

```bash
grpcurl -plaintext -d '{
  "topic_name": "orders",
  "queue_id": "0",
  "consumer_group": "my-service",
  "delivery_guarantee": "AT_MOST_ONCE"
}' localhost:50051 broker.Broker/Subscribe
```

**Пример — at-least-once (с подтверждением):**

```bash
grpcurl -plaintext -d '{
  "topic_name": "orders",
  "queue_id": "0",
  "consumer_group": "my-service",
  "delivery_guarantee": "AT_LEAST_ONCE"
}' localhost:50051 broker.Broker/Subscribe
```

Ответ:

```json
{
  "subscription_id": "sub-a1b2c3d4e5f6...",
  "topic_name": "orders",
  "queue_id": "0",
  "consumer_group": "my-service"
}
```

Сохраните `subscription_id` для чтения и ack.

---

### 4. Отправить сообщение (Publish)

**gRPC:** `Publish(PublishRequest) → PublishResponse`

- `topic_name` — топик
- `queue_id` — очередь (`""` или `"0"` — дефолтная)
- `payload` — тело сообщения (bytes; в JSON через grpcurl задаётся base64)
- `key` — опциональный ключ
- `headers` — опциональные заголовки (map string → string)

Ответ: `message_id`, `offset`.

**Пример (текст в base64):**

Строка `Hello, broker` в base64: `SGVsbG8sIGJyb2tlcg==`

```bash
grpcurl -plaintext -d '{
  "topic_name": "orders",
  "queue_id": "0",
  "payload": "SGVsbG8sIGJyb2tlcg==",
  "key": "order-123"
}' localhost:50051 broker.Broker/Publish
```

Ответ:

```json
{"message_id": "a1b2c3d4...", "offset": 0}
```

---

### 5. Получить сообщения (Consume)

**gRPC:** `Consume(ConsumeRequest) → ConsumeResponse`

- `subscription_id` — ID подписки из шага 3
- `max_messages` — максимум сообщений в ответе (например 10)

Ответ: список `messages` с полями `id`, `topic_name`, `queue_id`, `payload` (base64 в JSON), `key`, `headers`, `offset`, `delivery_id`.

**Пример:**

```bash
grpcurl -plaintext -d '{
  "subscription_id": "sub-a1b2c3d4e5f6...",
  "max_messages": 10
}' localhost:50051 broker.Broker/Consume
```

Пример ответа:

```json
{
  "messages": [
    {
      "id": "a1b2c3d4...",
      "topic_name": "orders",
      "queue_id": "0",
      "payload": "SGVsbG8sIGJyb2tlcg==",
      "key": "order-123",
      "offset": 0,
      "delivery_id": "abc-def-..."
    }
  ]
}
```

- Для **at-most-once** поле `delivery_id` может быть пустым; подтверждать не нужно.
- Для **at-least-once** после успешной обработки сообщения нужно вызвать `Ack` с этим `delivery_id`.

---

### 6. Подтвердить доставку (Ack) — только для at-least-once

**gRPC:** `Ack(AckRequest) → AckResponse`

- `subscription_id` — ID подписки
- `delivery_id` — значение из полученного сообщения

Если не вызвать `Ack` до истечения таймаута (по умолчанию 30 сек, см. `config.yaml`), сообщение будет доставлено снова.

```bash
grpcurl -plaintext -d '{
  "subscription_id": "sub-a1b2c3d4e5f6...",
  "delivery_id": "abc-def-..."
}' localhost:50051 broker.Broker/Ack
```

---

## Полный сценарий

```bash
# 1. Создать топик
grpcurl -plaintext -d '{"name": "orders", "retention_messages": 10000}' localhost:50051 broker.Broker/CreateTopic

# 2. Подписаться (at-least-once), сохрани subscription_id из ответа
grpcurl -plaintext -d '{"topic_name": "orders", "queue_id": "0", "consumer_group": "my-app", "delivery_guarantee": "AT_LEAST_ONCE"}' localhost:50051 broker.Broker/Subscribe

# 3. Опубликовать сообщение (payload = base64("Hello"))
grpcurl -plaintext -d '{"topic_name": "orders", "queue_id": "0", "payload": "SGVsbG8="}' localhost:50051 broker.Broker/Publish

# 4. Получить сообщения (подставь свой subscription_id)
grpcurl -plaintext -d '{"subscription_id": "sub-XXXX", "max_messages": 10}' localhost:50051 broker.Broker/Consume

# 5. Подтвердить доставку (подставь subscription_id и delivery_id из сообщения)
grpcurl -plaintext -d '{"subscription_id": "sub-XXXX", "delivery_id": "..."}' localhost:50051 broker.Broker/Ack
```

---

## Дополнительные методы

- **ListTopics** — список топиков:  
  `grpcurl -plaintext -d '{}' localhost:50051 broker.Broker/ListTopics`

- **ListQueues** — список очередей топика:  
  `grpcurl -plaintext -d '{"topic_name": "orders"}' localhost:50051 broker.Broker/ListQueues`

---

## Гарантии доставки

| Гарантия        | Поведение |
|-----------------|-----------|
| **AT_MOST_ONCE** | Сообщение отдаётся потребителю не более одного раза. Подтверждение (Ack) не требуется. |
| **AT_LEAST_ONCE** | Сообщение хранится до вызова Ack. Если Ack не пришёл в течение `ack_timeout_seconds`, оно снова отдаётся в Consume. После обработки обязательно вызывайте Ack. |

---

## Конфигурация (config.yaml)

| Параметр | Описание |
|----------|----------|
| `server.grpc_port` | Порт gRPC (по умолчанию 50051) |
| `server.http_port` | Зарезервировано под HTTP |
| `broker.default_retention_messages` | Лимит сообщений в очереди |
| `broker.ack_timeout_seconds` | Таймаут до повторной доставки при at-least-once (сек) |
| `broker.max_message_size` | Максимальный размер сообщения (байты) |

---

## Тесты и бенчмарки

```bash
# Все тесты
go test ./...

# С выводом
go test ./... -v

# Бенчмарки (usecase + repository)
go test ./internal/usecase/... ./internal/repository/memory/... -bench=. -benchmem -run=^$
```

Тесты покрывают: use case (топики, публикация, подписки, consume, ack), in-memory репозитории, конфиг и gRPC handler. Бенчмарки: Publish, Consume, полный цикл Publish+Consume+Ack, Append/Read в репозитории.

---

## Требования

- **Go 1.21+**
- Для примеров из README: [grpcurl](https://github.com/fullstorydev/grpcurl) (`go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`)
- Чтобы grpcurl видел сервис: один раз выполнить `task proto` или на Windows `task proto:win` (генерация кода из `api/proto/broker.proto`)

Сервер по умолчанию слушает **localhost:50051** без TLS (`-plaintext` в grpcurl).
