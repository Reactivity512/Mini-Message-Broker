package grpc

import (
	"context"
	"testing"

	"queue-service/internal/delivery/grpc/pb"
	"queue-service/internal/repository/memory"
	"queue-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler(t *testing.T) *BrokerHandler {
	t.Helper()
	topics := memory.NewTopicRepository()
	queues := memory.NewQueueRepository()
	msgs := memory.NewMessageRepository()
	subs := memory.NewSubscriptionRepository()
	pending := memory.NewPendingDeliveryRepository()
	topicUC := usecase.NewTopicUseCase(topics, queues)
	pub := usecase.NewPublishUseCase(topics, queues, msgs, 1024*1024)
	subUC := usecase.NewSubscriptionUseCase(subs, topics, queues, 30)
	consumeUC := usecase.NewConsumeUseCase(subs, msgs, pending)
	return NewBrokerHandler(topicUC, pub, subUC, consumeUC)
}

func TestBrokerHandler_CreateTopic(t *testing.T) {
	ctx := context.Background()
	h := newTestHandler(t)

	resp, err := h.CreateTopic(ctx, &pb.CreateTopicRequest{Name: "orders", RetentionMessages: 5000})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	if resp.Name != "orders" || resp.RetentionMessages != 5000 {
		t.Errorf("got %+v", resp)
	}

	// duplicate
	_, err = h.CreateTopic(ctx, &pb.CreateTopicRequest{Name: "orders", RetentionMessages: 1000})
	if err == nil {
		t.Fatal("expected error for duplicate topic")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.AlreadyExists {
		t.Errorf("want AlreadyExists, got %v", err)
	}
}

func TestBrokerHandler_Publish_Consume_Ack(t *testing.T) {
	ctx := context.Background()
	h := newTestHandler(t)

	_, _ = h.CreateTopic(ctx, &pb.CreateTopicRequest{Name: "orders", RetentionMessages: 10000})
	sub, err := h.Subscribe(ctx, &pb.SubscribeRequest{
		TopicName:         "orders",
		QueueId:           "0",
		ConsumerGroup:     "g1",
		DeliveryGuarantee: pb.DeliveryGuarantee_AT_LEAST_ONCE,
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	_, err = h.Publish(ctx, &pb.PublishRequest{
		TopicName: "orders",
		QueueId:   "0",
		Payload:   []byte("hello"),
		Key:       "k1",
	})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	consumeResp, err := h.Consume(ctx, &pb.ConsumeRequest{SubscriptionId: sub.SubscriptionId, MaxMessages: 10})
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if len(consumeResp.Messages) != 1 {
		t.Fatalf("want 1 message, got %d", len(consumeResp.Messages))
	}
	if string(consumeResp.Messages[0].Payload) != "hello" || consumeResp.Messages[0].DeliveryId == "" {
		t.Errorf("got message %+v", consumeResp.Messages[0])
	}

	_, err = h.Ack(ctx, &pb.AckRequest{SubscriptionId: sub.SubscriptionId, DeliveryId: consumeResp.Messages[0].DeliveryId})
	if err != nil {
		t.Fatalf("Ack: %v", err)
	}
}

func TestBrokerHandler_ListTopics_ListQueues(t *testing.T) {
	ctx := context.Background()
	h := newTestHandler(t)
	_, _ = h.CreateTopic(ctx, &pb.CreateTopicRequest{Name: "orders", RetentionMessages: 10000})
	_, _ = h.CreateQueue(ctx, &pb.CreateQueueRequest{TopicName: "orders", QueueId: "1"})

	listTopics, err := h.ListTopics(ctx, &pb.ListTopicsRequest{})
	if err != nil {
		t.Fatalf("ListTopics: %v", err)
	}
	if len(listTopics.Topics) != 1 || listTopics.Topics[0].Name != "orders" {
		t.Errorf("ListTopics: %+v", listTopics.Topics)
	}

	listQueues, err := h.ListQueues(ctx, &pb.ListQueuesRequest{TopicName: "orders"})
	if err != nil {
		t.Fatalf("ListQueues: %v", err)
	}
	if len(listQueues.Queues) != 2 {
		t.Errorf("want 2 queues (0 and 1), got %d", len(listQueues.Queues))
	}
}

func TestBrokerHandler_Publish_topicNotFound(t *testing.T) {
	ctx := context.Background()
	h := newTestHandler(t)
	_, err := h.Publish(ctx, &pb.PublishRequest{TopicName: "nonexistent", QueueId: "0", Payload: []byte("x")})
	if err == nil {
		t.Fatal("expected error")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("want NotFound, got %v", st.Code())
	}
}

func TestBrokerHandler_Subscribe_deliveryGuarantee(t *testing.T) {
	ctx := context.Background()
	h := newTestHandler(t)
	_, _ = h.CreateTopic(ctx, &pb.CreateTopicRequest{Name: "t1", RetentionMessages: 10000})

	sub, _ := h.Subscribe(ctx, &pb.SubscribeRequest{
		TopicName:         "t1",
		QueueId:           "0",
		ConsumerGroup:     "g1",
		DeliveryGuarantee: pb.DeliveryGuarantee_AT_MOST_ONCE,
	})
	if sub.SubscriptionId == "" {
		t.Error("expected subscription id")
	}
}
