package domain

import (
	"context"
	"time"
)

// TopicRepository manages topics.
type TopicRepository interface {
	Create(ctx context.Context, topic *Topic) error
	Get(ctx context.Context, name string) (*Topic, error)
	Delete(ctx context.Context, name string) error
	List(ctx context.Context) ([]*Topic, error)
}

// QueueRepository manages queues (partitions) per topic.
type QueueRepository interface {
	Create(ctx context.Context, queue *Queue) error
	Get(ctx context.Context, topicName, queueID string) (*Queue, error)
	ListByTopic(ctx context.Context, topicName string) ([]*Queue, error)
}

// MessageRepository stores and retrieves messages.
type MessageRepository interface {
	Append(ctx context.Context, msg *Message) error
	Read(ctx context.Context, topicName, queueID string, offset, limit int) ([]*Message, error)
	GetByID(ctx context.Context, topicName, queueID, messageID string) (*Message, error)
}

// SubscriptionRepository manages consumer subscriptions.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *Subscription) error
	Get(ctx context.Context, id string) (*Subscription, error)
	GetByTopicAndGroup(ctx context.Context, topicName, consumerGroup string) (*Subscription, error)
	ListByTopic(ctx context.Context, topicName string) ([]*Subscription, error)
	List(ctx context.Context) ([]*Subscription, error)
	AdvanceOffset(ctx context.Context, id string, offset int64) error
}

// PendingDeliveryRepository tracks unacknowledged deliveries (at-least-once).
type PendingDeliveryRepository interface {
	Add(ctx context.Context, subID string, pd *PendingDelivery) error
	Ack(ctx context.Context, subID, deliveryID string) (ackedOffset int64, err error)
	Expired(ctx context.Context, subID string, before time.Time) ([]*PendingDelivery, error)
	Remove(ctx context.Context, subID, deliveryID string) error
}
