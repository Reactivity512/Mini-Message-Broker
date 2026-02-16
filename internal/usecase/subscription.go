package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"queue-service/internal/domain"
)

var ErrSubscriptionExists = errors.New("subscription already exists for topic and consumer group")
var ErrSubscriptionNotFound = errors.New("subscription not found")

type SubscriptionUseCase struct {
	subs       domain.SubscriptionRepository
	topics     domain.TopicRepository
	queues     domain.QueueRepository
	ackTimeout time.Duration
}

func NewSubscriptionUseCase(
	subs domain.SubscriptionRepository,
	topics domain.TopicRepository,
	queues domain.QueueRepository,
	ackTimeoutSeconds int,
) *SubscriptionUseCase {
	return &SubscriptionUseCase{
		subs:       subs,
		topics:     topics,
		queues:     queues,
		ackTimeout: time.Duration(ackTimeoutSeconds) * time.Second,
	}
}

func (u *SubscriptionUseCase) Subscribe(ctx context.Context, topicName, queueID, consumerGroup string, guarantee domain.DeliveryGuarantee) (*domain.Subscription, error) {
	_, err := u.topics.Get(ctx, topicName)
	if err != nil {
		return nil, ErrTopicNotFound
	}
	if queueID == "" {
		queueID = "0"
	}
	_, err = u.queues.Get(ctx, topicName, queueID)
	if err != nil {
		return nil, err
	}
	existing, _ := u.subs.GetByTopicAndGroup(ctx, topicName, consumerGroup)
	if existing != nil {
		return nil, ErrSubscriptionExists
	}
	sub := &domain.Subscription{
		ID:                genSubID(),
		TopicName:         topicName,
		QueueID:           queueID,
		ConsumerGroup:     consumerGroup,
		DeliveryGuarantee: guarantee,
		AckTimeout:        u.ackTimeout,
		CreatedAt:         time.Now(),
	}
	if err := u.subs.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (u *SubscriptionUseCase) GetSubscription(ctx context.Context, id string) (*domain.Subscription, error) {
	return u.subs.Get(ctx, id)
}

func (u *SubscriptionUseCase) ListSubscriptions(ctx context.Context, topicName string) ([]*domain.Subscription, error) {
	if topicName != "" {
		return u.subs.ListByTopic(ctx, topicName)
	}
	return u.subs.List(ctx)
}

func genSubID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "sub-" + hex.EncodeToString(b)
}
