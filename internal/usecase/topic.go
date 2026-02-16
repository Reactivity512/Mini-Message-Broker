package usecase

import (
	"context"
	"errors"
	"time"

	"queue-service/internal/domain"
)

var (
	ErrTopicExists   = errors.New("topic already exists")
	ErrTopicNotFound = errors.New("topic not found")
)

type TopicUseCase struct {
	topics domain.TopicRepository
	queues domain.QueueRepository
}

func NewTopicUseCase(topics domain.TopicRepository, queues domain.QueueRepository) *TopicUseCase {
	return &TopicUseCase{topics: topics, queues: queues}
}

func (u *TopicUseCase) CreateTopic(ctx context.Context, name string, retentionMessages int) (*domain.Topic, error) {
	_, err := u.topics.Get(ctx, name)
	if err == nil {
		return nil, ErrTopicExists
	}
	topic := &domain.Topic{
		Name:              name,
		RetentionMessages: retentionMessages,
		CreatedAt:         time.Now(),
	}
	if err := u.topics.Create(ctx, topic); err != nil {
		return nil, err
	}
	// Create default queue "0" for the topic
	queue := &domain.Queue{
		TopicName: name,
		QueueID:   "0",
		CreatedAt: time.Now(),
	}
	if err := u.queues.Create(ctx, queue); err != nil {
		_ = u.topics.Delete(ctx, name)
		return nil, err
	}
	return topic, nil
}

func (u *TopicUseCase) GetTopic(ctx context.Context, name string) (*domain.Topic, error) {
	return u.topics.Get(ctx, name)
}

func (u *TopicUseCase) ListTopics(ctx context.Context) ([]*domain.Topic, error) {
	return u.topics.List(ctx)
}

func (u *TopicUseCase) DeleteTopic(ctx context.Context, name string) error {
	return u.topics.Delete(ctx, name)
}

func (u *TopicUseCase) CreateQueue(ctx context.Context, topicName, queueID string) (*domain.Queue, error) {
	_, err := u.topics.Get(ctx, topicName)
	if err != nil {
		return nil, ErrTopicNotFound
	}
	queue := &domain.Queue{
		TopicName: topicName,
		QueueID:   queueID,
		CreatedAt: time.Now(),
	}
	if err := u.queues.Create(ctx, queue); err != nil {
		return nil, err
	}
	return queue, nil
}

func (u *TopicUseCase) ListQueues(ctx context.Context, topicName string) ([]*domain.Queue, error) {
	return u.queues.ListByTopic(ctx, topicName)
}
