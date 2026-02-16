package usecase

import (
	"context"
	"testing"

	"queue-service/internal/domain"
	"queue-service/internal/repository/memory"
)

// Запуск TestFullFlow создают topic -> subscribe -> publish -> consume -> ack (at-least-once).
func TestFullFlow(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	msgs := memory.NewMessageRepository()
	subs := memory.NewSubscriptionRepository()
	pending := memory.NewPendingDeliveryRepository()

	topicUC := NewTopicUseCase(topics, queues)
	pub := NewPublishUseCase(topics, queues, msgs, 1024*1024)
	subUC := NewSubscriptionUseCase(subs, topics, queues, 30)
	consumeUC := NewConsumeUseCase(subs, msgs, pending)

	// Create topic
	_, err := topicUC.CreateTopic(ctx, "orders", 10000)
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	// Subscribe at-least-once
	sub, err := subUC.Subscribe(ctx, "orders", "0", "consumer-1", domain.AtLeastOnce)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	// Publish two messages
	m1, _ := pub.Publish(ctx, "orders", "0", []byte("order-1"), "k1", nil)
	m2, _ := pub.Publish(ctx, "orders", "0", []byte("order-2"), "k2", nil)
	if m1.Offset != 0 || m2.Offset != 1 {
		t.Errorf("offsets: %d %d", m1.Offset, m2.Offset)
	}
	// Consume
	out, err := consumeUC.Consume(ctx, sub.ID, 10)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("want 2 messages, got %d", len(out))
	}
	// Ack first only
	err = consumeUC.Ack(ctx, sub.ID, out[0].DeliveryID)
	if err != nil {
		t.Fatalf("Ack: %v", err)
	}
	// Consume again: should redeliver only the second (or we ack both and get empty)
	err = consumeUC.Ack(ctx, sub.ID, out[1].DeliveryID)
	if err != nil {
		t.Fatalf("Ack second: %v", err)
	}
	out2, _ := consumeUC.Consume(ctx, sub.ID, 10)
	if len(out2) != 0 {
		t.Errorf("after all acks consume should be empty, got %d", len(out2))
	}
}
