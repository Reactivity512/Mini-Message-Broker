package usecase

import (
	"context"
	"testing"

	"queue-service/internal/repository/memory"
)

func TestPublishUseCase_Publish(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	msgs := memory.NewMessageRepository()
	topicUC := NewTopicUseCase(topics, queues)
	_, _ = topicUC.CreateTopic(ctx, "orders", 10000)
	pub := NewPublishUseCase(topics, queues, msgs, 1024)

	msg, err := pub.Publish(ctx, "orders", "0", []byte("hello"), "key1", nil)
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if msg.ID == "" || msg.TopicName != "orders" || msg.QueueID != "0" || string(msg.Payload) != "hello" {
		t.Errorf("got msg %+v", msg)
	}
	if msg.Offset != 0 {
		t.Errorf("first message offset want 0, got %d", msg.Offset)
	}

	msg2, _ := pub.Publish(ctx, "orders", "", []byte("world"), "", map[string]string{"x": "y"})
	if msg2.Offset != 1 {
		t.Errorf("second message offset want 1, got %d", msg2.Offset)
	}
}

func TestPublishUseCase_topicNotFound(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	pub := NewPublishUseCase(topics, queues, memory.NewMessageRepository(), 1024)

	_, err := pub.Publish(ctx, "nonexistent", "0", []byte("x"), "", nil)
	if err != ErrTopicNotFound {
		t.Errorf("want ErrTopicNotFound, got %v", err)
	}
}

func TestPublishUseCase_messageTooLarge(t *testing.T) {
	ctx := context.Background()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	topicUC := NewTopicUseCase(topics, queues)
	_, _ = topicUC.CreateTopic(ctx, "orders", 10000)
	pub := NewPublishUseCase(topics, queues, memory.NewMessageRepository(), 5)

	_, err := pub.Publish(ctx, "orders", "0", []byte("hello world"), "", nil)
	if err != ErrMessageTooLarge {
		t.Errorf("want ErrMessageTooLarge, got %v", err)
	}
}
