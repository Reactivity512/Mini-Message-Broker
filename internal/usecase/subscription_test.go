package usecase

import (
	"context"
	"testing"

	"queue-service/internal/domain"
	"queue-service/internal/repository/memory"
)

func TestSubscriptionUseCase_Subscribe(t *testing.T) {
	ctx := context.Background()
	subs := memory.NewSubscriptionRepository()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	topicUC := NewTopicUseCase(topics, queues)
	_, _ = topicUC.CreateTopic(ctx, "orders", 10000)
	uc := NewSubscriptionUseCase(subs, topics, queues, 30)

	sub, err := uc.Subscribe(ctx, "orders", "0", "my-group", domain.AtLeastOnce)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if sub.ID == "" || sub.TopicName != "orders" || sub.QueueID != "0" || sub.ConsumerGroup != "my-group" {
		t.Errorf("got sub %+v", sub)
	}
	if sub.DeliveryGuarantee != domain.AtLeastOnce {
		t.Errorf("want AtLeastOnce, got %v", sub.DeliveryGuarantee)
	}
}

func TestSubscriptionUseCase_Subscribe_duplicateReturnsErr(t *testing.T) {
	ctx := context.Background()
	subs := memory.NewSubscriptionRepository()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	topicUC := NewTopicUseCase(topics, queues)
	_, _ = topicUC.CreateTopic(ctx, "orders", 10000)
	uc := NewSubscriptionUseCase(subs, topics, queues, 30)

	_, _ = uc.Subscribe(ctx, "orders", "0", "my-group", domain.AtMostOnce)
	_, err := uc.Subscribe(ctx, "orders", "0", "my-group", domain.AtMostOnce)
	if err != ErrSubscriptionExists {
		t.Errorf("want ErrSubscriptionExists, got %v", err)
	}
}

func TestSubscriptionUseCase_Subscribe_topicNotFound(t *testing.T) {
	ctx := context.Background()
	uc := NewSubscriptionUseCase(
		memory.NewSubscriptionRepository(),
		memory.NewTopicRepository(),
		memory.NewQueueRepository(),
		30,
	)
	_, err := uc.Subscribe(ctx, "nonexistent", "0", "g", domain.AtMostOnce)
	if err != ErrTopicNotFound {
		t.Errorf("want ErrTopicNotFound, got %v", err)
	}
}

func TestSubscriptionUseCase_GetSubscription_ListSubscriptions(t *testing.T) {
	ctx := context.Background()
	subs := memory.NewSubscriptionRepository()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	topicUC := NewTopicUseCase(topics, queues)
	_, _ = topicUC.CreateTopic(ctx, "orders", 10000)
	uc := NewSubscriptionUseCase(subs, topics, queues, 30)

	created, _ := uc.Subscribe(ctx, "orders", "0", "g1", domain.AtMostOnce)
	got, err := uc.GetSubscription(ctx, created.ID)
	if err != nil || got.ID != created.ID {
		t.Fatalf("GetSubscription: err=%v got=%+v", err, got)
	}
	list, _ := uc.ListSubscriptions(ctx, "orders")
	if len(list) != 1 {
		t.Errorf("ListSubscriptions(topic): want 1, got %d", len(list))
	}
	all, _ := uc.ListSubscriptions(ctx, "")
	if len(all) != 1 {
		t.Errorf("ListSubscriptions(all): want 1, got %d", len(all))
	}
}
