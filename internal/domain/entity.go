package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
)

// Гарантия доставки определяет семантику доставки сообщений.
type DeliveryGuarantee int

const (
	AtMostOnce  DeliveryGuarantee = 0 // Fire-and-forget, no ack
	AtLeastOnce DeliveryGuarantee = 1 // Хранить до подтверждения, повторная доставка по истечении времени ожидания.
)

// Topic represents a named stream of messages (like Kafka topic).
type Topic struct {
	Name              string
	RetentionMessages int
	CreatedAt         time.Time
}

// Queue is a FIFO queue bound to a topic (partition/queue abstraction).
type Queue struct {
	TopicName string
	QueueID   string // partition or queue identifier
	CreatedAt time.Time
}

// Subscription represents a consumer subscription (consumer group + queue).
type Subscription struct {
	ID                string
	TopicName         string
	QueueID           string
	ConsumerGroup     string
	DeliveryGuarantee DeliveryGuarantee
	AckTimeout        time.Duration
	Offset            int64 // next offset to read for this consumer
	CreatedAt         time.Time
}

type Message struct {
	ID         string
	TopicName  string
	QueueID    string
	Payload    []byte
	Key        string
	Headers    map[string]string
	Offset     int64
	CreatedAt  time.Time
	DeliveryID string // for at-least-once ack
}

// PendingDelivery — это сообщение, которое не было подтверждено как минимум один раз
type PendingDelivery struct {
	Message    Message
	ExpiresAt  time.Time
	DeliveryID string
}
