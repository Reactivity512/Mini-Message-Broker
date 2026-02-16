package usecase

import (
	"context"
	"testing"

	"queue-service/internal/repository/memory"
)

func TestTopicUseCase_CreateTopic(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	uc := NewTopicUseCase(topics, queues)

	t.Run("creates topic and default queue", func(t *testing.T) {
		topic, err := uc.CreateTopic(ctx, "orders", 5000)
		if err != nil {
			t.Fatalf("CreateTopic: %v", err)
		}
		if topic.Name != "orders" || topic.RetentionMessages != 5000 {
			t.Errorf("got topic %+v", topic)
		}
		queuesList, _ := uc.ListQueues(ctx, "orders")
		if len(queuesList) != 1 || queuesList[0].QueueID != "0" {
			t.Errorf("expected default queue 0, got %+v", queuesList)
		}
	})

	t.Run("duplicate topic returns ErrTopicExists", func(t *testing.T) {
		_, err := uc.CreateTopic(ctx, "orders", 1000)
		if err != ErrTopicExists {
			t.Errorf("want ErrTopicExists, got %v", err)
		}
	})
}

func TestTopicUseCase_GetTopic_ListTopics(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	uc := NewTopicUseCase(topics, queues)

	_, _ = uc.CreateTopic(ctx, "orders", 1000)
	_, _ = uc.CreateTopic(ctx, "events", 2000)

	got, err := uc.GetTopic(ctx, "orders")
	if err != nil || got.Name != "orders" {
		t.Fatalf("GetTopic: err=%v got=%+v", err, got)
	}
	list, err := uc.ListTopics(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("ListTopics: err=%v len=%d", err, len(list))
	}
}

func TestTopicUseCase_CreateQueue_ListQueues(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	uc := NewTopicUseCase(topics, queues)

	_, _ = uc.CreateTopic(ctx, "orders", 1000)

	q, err := uc.CreateQueue(ctx, "orders", "1")
	if err != nil {
		t.Fatalf("CreateQueue: %v", err)
	}
	if q.QueueID != "1" || q.TopicName != "orders" {
		t.Errorf("got queue %+v", q)
	}

	list, _ := uc.ListQueues(ctx, "orders")
	if len(list) != 2 {
		t.Errorf("expected 2 queues (0 and 1), got %d", len(list))
	}
}

func TestTopicUseCase_CreateQueue_topicNotFound(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	uc := NewTopicUseCase(topics, queues)

	_, err := uc.CreateQueue(ctx, "nonexistent", "0")
	if err != ErrTopicNotFound {
		t.Errorf("want ErrTopicNotFound, got %v", err)
	}
}
