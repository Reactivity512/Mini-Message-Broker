package grpc

import (
	"context"

	"queue-service/internal/delivery/grpc/pb"
	"queue-service/internal/domain"
	"queue-service/internal/usecase"
)

type BrokerHandler struct {
	pb.UnimplementedBrokerServer
	topics    *usecase.TopicUseCase
	publish   *usecase.PublishUseCase
	subscribe *usecase.SubscriptionUseCase
	consume   *usecase.ConsumeUseCase
}

func NewBrokerHandler(
	topics *usecase.TopicUseCase,
	publish *usecase.PublishUseCase,
	subscribe *usecase.SubscriptionUseCase,
	consume *usecase.ConsumeUseCase,
) *BrokerHandler {
	return &BrokerHandler{
		topics:    topics,
		publish:   publish,
		subscribe: subscribe,
		consume:   consume,
	}
}

func (h *BrokerHandler) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	retention := int(req.RetentionMessages)
	if retention <= 0 {
		retention = 10000
	}
	t, err := h.topics.CreateTopic(ctx, req.Name, retention)
	if err != nil {
		if err == usecase.ErrTopicExists {
			return nil, errAlreadyExists("topic", req.Name)
		}
		return nil, errInternal(err)
	}
	return &pb.CreateTopicResponse{Name: t.Name, RetentionMessages: int32(t.RetentionMessages)}, nil
}

func (h *BrokerHandler) CreateQueue(ctx context.Context, req *pb.CreateQueueRequest) (*pb.CreateQueueResponse, error) {
	queueID := req.QueueId
	if queueID == "" {
		queueID = "0"
	}
	q, err := h.topics.CreateQueue(ctx, req.TopicName, queueID)
	if err != nil {
		if err == usecase.ErrTopicNotFound {
			return nil, errNotFound("topic", req.TopicName)
		}
		return nil, errInternal(err)
	}
	return &pb.CreateQueueResponse{TopicName: q.TopicName, QueueId: q.QueueID}, nil
}

func (h *BrokerHandler) ListTopics(ctx context.Context, _ *pb.ListTopicsRequest) (*pb.ListTopicsResponse, error) {
	list, err := h.topics.ListTopics(ctx)
	if err != nil {
		return nil, errInternal(err)
	}
	topics := make([]*pb.TopicInfo, 0, len(list))
	for _, t := range list {
		topics = append(topics, &pb.TopicInfo{Name: t.Name, RetentionMessages: int32(t.RetentionMessages)})
	}
	return &pb.ListTopicsResponse{Topics: topics}, nil
}

func (h *BrokerHandler) ListQueues(ctx context.Context, req *pb.ListQueuesRequest) (*pb.ListQueuesResponse, error) {
	list, err := h.topics.ListQueues(ctx, req.TopicName)
	if err != nil {
		return nil, errInternal(err)
	}
	queues := make([]*pb.QueueInfo, 0, len(list))
	for _, q := range list {
		queues = append(queues, &pb.QueueInfo{TopicName: q.TopicName, QueueId: q.QueueID})
	}
	return &pb.ListQueuesResponse{Queues: queues}, nil
}

func (h *BrokerHandler) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	msg, err := h.publish.Publish(ctx, req.TopicName, req.QueueId, req.Payload, req.Key, req.Headers)
	if err != nil {
		if err == usecase.ErrTopicNotFound {
			return nil, errNotFound("topic", req.TopicName)
		}
		if err == usecase.ErrMessageTooLarge {
			return nil, errInvalidArg("message too large")
		}
		return nil, errInternal(err)
	}
	return &pb.PublishResponse{MessageId: msg.ID, Offset: msg.Offset}, nil
}

func (h *BrokerHandler) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	guarantee := domain.AtMostOnce
	switch req.DeliveryGuarantee {
	case pb.DeliveryGuarantee_AT_LEAST_ONCE:
		guarantee = domain.AtLeastOnce
	case pb.DeliveryGuarantee_AT_MOST_ONCE:
		guarantee = domain.AtMostOnce
	}
	sub, err := h.subscribe.Subscribe(ctx, req.TopicName, req.QueueId, req.ConsumerGroup, guarantee)
	if err != nil {
		if err == usecase.ErrTopicNotFound {
			return nil, errNotFound("topic", req.TopicName)
		}
		if err == usecase.ErrSubscriptionExists {
			return nil, errAlreadyExists("subscription", req.TopicName+"/"+req.ConsumerGroup)
		}
		return nil, errInternal(err)
	}
	return &pb.SubscribeResponse{
		SubscriptionId: sub.ID,
		TopicName:      sub.TopicName,
		QueueId:        sub.QueueID,
		ConsumerGroup:  sub.ConsumerGroup,
	}, nil
}

func (h *BrokerHandler) Consume(ctx context.Context, req *pb.ConsumeRequest) (*pb.ConsumeResponse, error) {
	max := int(req.MaxMessages)
	if max <= 0 {
		max = 10
	}
	msgs, err := h.consume.Consume(ctx, req.SubscriptionId, max)
	if err != nil {
		if err == usecase.ErrSubscriptionNotFound {
			return nil, errNotFound("subscription", req.SubscriptionId)
		}
		return nil, errInternal(err)
	}
	out := make([]*pb.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, &pb.Message{
			Id:         m.ID,
			TopicName:  m.TopicName,
			QueueId:    m.QueueID,
			Payload:    m.Payload,
			Key:        m.Key,
			Headers:    m.Headers,
			Offset:     m.Offset,
			DeliveryId: m.DeliveryID,
		})
	}
	return &pb.ConsumeResponse{Messages: out}, nil
}

func (h *BrokerHandler) Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error) {
	if err := h.consume.Ack(ctx, req.SubscriptionId, req.DeliveryId); err != nil {
		if err == usecase.ErrSubscriptionNotFound {
			return nil, errNotFound("subscription", req.SubscriptionId)
		}
		return nil, errInternal(err)
	}
	return &pb.AckResponse{}, nil
}
