package usecase

import (
	"context"
	"testing"

	"queue-service/internal/domain"
	"queue-service/internal/repository/memory"
)

func TestConsumeUseCase_Consume_atMostOnce(t *testing.T) {
	ctx := context.Background()
	subs := memory.NewSubscriptionRepository()
	msgs := memory.NewMessageRepository()
	pending := memory.NewPendingDeliveryRepository()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()

	topicUC := NewTopicUseCase(topics, queues)
	pub := NewPublishUseCase(topics, queues, msgs, 1024)
	subUC := NewSubscriptionUseCase(subs, topics, queues, 30)
	consumeUC := NewConsumeUseCase(subs, msgs, pending)

	_, _ = topicUC.CreateTopic(ctx, "orders", 10000)
	_, _ = pub.Publish(ctx, "orders", "0", []byte("m1"), "", nil)
	_, _ = pub.Publish(ctx, "orders", "0", []byte("m2"), "", nil)
	sub, _ := subUC.Subscribe(ctx, "orders", "0", "g1", domain.AtMostOnce)

	out, err := consumeUC.Consume(ctx, sub.ID, 10)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("want 2 messages, got %d", len(out))
	}
	if string(out[0].Payload) != "m1" || string(out[1].Payload) != "m2" {
		t.Errorf("wrong payloads: %q %q", out[0].Payload, out[1].Payload)
	}
	// at-most-once: offset advanced, second consume returns empty
	out2, _ := consumeUC.Consume(ctx, sub.ID, 10)
	if len(out2) != 0 {
		t.Errorf("second consume should return empty, got %d", len(out2))
	}
}

func TestConsumeUseCase_Consume_atLeastOnce_and_Ack(t *testing.T) {
	ctx := context.Background()
	subs := memory.NewSubscriptionRepository()
	msgs := memory.NewMessageRepository()
	pending := memory.NewPendingDeliveryRepository()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()

	topicUC := NewTopicUseCase(topics, queues)
	pub := NewPublishUseCase(topics, queues, msgs, 1024)
	subUC := NewSubscriptionUseCase(subs, topics, queues, 30)
	consumeUC := NewConsumeUseCase(subs, msgs, pending)

	_, _ = topicUC.CreateTopic(ctx, "orders", 10000)
	_, _ = pub.Publish(ctx, "orders", "0", []byte("m1"), "", nil)
	sub, _ := subUC.Subscribe(ctx, "orders", "0", "g1", domain.AtLeastOnce)

	out, err := consumeUC.Consume(ctx, sub.ID, 10)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("want 1 message, got %d", len(out))
	}
	if out[0].DeliveryID == "" {
		t.Error("at-least-once must set DeliveryID")
	}
	deliveryID := out[0].DeliveryID

	err = consumeUC.Ack(ctx, sub.ID, deliveryID)
	if err != nil {
		t.Fatalf("Ack: %v", err)
	}
	// after ack, consume again should return empty (no new messages)
	out2, _ := consumeUC.Consume(ctx, sub.ID, 10)
	if len(out2) != 0 {
		t.Errorf("after ack consume should return empty, got %d", len(out2))
	}
}

func TestConsumeUseCase_Consume_subscriptionNotFound(t *testing.T) {
	ctx := context.Background()
	consumeUC := NewConsumeUseCase(
		memory.NewSubscriptionRepository(),
		memory.NewMessageRepository(),
		memory.NewPendingDeliveryRepository(),
	)
	_, err := consumeUC.Consume(ctx, "sub-nonexistent", 10)
	if err != ErrSubscriptionNotFound {
		t.Errorf("want ErrSubscriptionNotFound, got %v", err)
	}
}

func TestConsumeUseCase_Ack_subscriptionNotFound(t *testing.T) {
	ctx := context.Background()
	consumeUC := NewConsumeUseCase(
		memory.NewSubscriptionRepository(),
		memory.NewMessageRepository(),
		memory.NewPendingDeliveryRepository(),
	)
	err := consumeUC.Ack(ctx, "sub-nonexistent", "delivery-1")
	if err != ErrSubscriptionNotFound {
		t.Errorf("want ErrSubscriptionNotFound, got %v", err)
	}
}
