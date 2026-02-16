package usecase

import (
	"context"
	"testing"

	"queue-service/internal/domain"
	"queue-service/internal/repository/memory"
)

func setupBench(b *testing.B) (
	context.Context,
	*TopicUseCase,
	*PublishUseCase,
	*SubscriptionUseCase,
	*ConsumeUseCase,
	string,
) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	msgs := memory.NewMessageRepository()
	subs := memory.NewSubscriptionRepository()
	pending := memory.NewPendingDeliveryRepository()

	topicUC := NewTopicUseCase(topics, queues)
	pub := NewPublishUseCase(topics, queues, msgs, 10*1024*1024)
	subUC := NewSubscriptionUseCase(subs, topics, queues, 30)
	consumeUC := NewConsumeUseCase(subs, msgs, pending)

	_, _ = topicUC.CreateTopic(ctx, "bench", 1000000)
	sub, _ := subUC.Subscribe(ctx, "bench", "0", "g1", domain.AtLeastOnce)
	return ctx, topicUC, pub, subUC, consumeUC, sub.ID
}

func BenchmarkPublish(b *testing.B) {
	ctx, _, pub, _, _, _ := setupBench(b)
	payload := []byte("hello world")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pub.Publish(ctx, "bench", "0", payload, "", nil)
	}
}

func BenchmarkPublishParallel(b *testing.B) {
	ctx, _, pub, _, _, _ := setupBench(b)
	payload := []byte("hello world")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = pub.Publish(ctx, "bench", "0", payload, "", nil)
		}
	})
}

func BenchmarkConsume_atMostOnce(b *testing.B) {
	ctx, _, pub, subUC, consumeUC, _ := setupBench(b)
	sub, _ := subUC.Subscribe(ctx, "bench", "0", "g2", domain.AtMostOnce)
	payload := []byte("x")
	for i := 0; i < 1000; i++ {
		_, _ = pub.Publish(ctx, "bench", "0", payload, "", nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = consumeUC.Consume(ctx, sub.ID, 100)
	}
}

func BenchmarkConsume_atLeastOnce_withAck(b *testing.B) {
	ctx, _, pub, _, consumeUC, subID := setupBench(b)
	payload := []byte("x")
	for i := 0; i < 10000; i++ {
		_, _ = pub.Publish(ctx, "bench", "0", payload, "", nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgs, _ := consumeUC.Consume(ctx, subID, 10)
		for _, m := range msgs {
			_ = consumeUC.Ack(ctx, subID, m.DeliveryID)
		}
	}
}

func BenchmarkFullFlow_PublishConsumeAck(b *testing.B) {
	ctx, _, pub, _, consumeUC, subID := setupBench(b)
	payload := []byte("payload")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pub.Publish(ctx, "bench", "0", payload, "", nil)
		msgs, _ := consumeUC.Consume(ctx, subID, 1)
		if len(msgs) > 0 {
			_ = consumeUC.Ack(ctx, subID, msgs[0].DeliveryID)
		}
	}
}
