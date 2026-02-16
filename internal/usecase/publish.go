package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"queue-service/internal/domain"
)

var ErrMessageTooLarge = errors.New("message exceeds max size")

type PublishUseCase struct {
	topics   domain.TopicRepository
	queues   domain.QueueRepository
	messages domain.MessageRepository
	maxSize  int
}

func NewPublishUseCase(
	topics domain.TopicRepository,
	queues domain.QueueRepository,
	messages domain.MessageRepository,
	maxMessageSize int,
) *PublishUseCase {
	return &PublishUseCase{
		topics:   topics,
		queues:   queues,
		messages: messages,
		maxSize:  maxMessageSize,
	}
}

func (u *PublishUseCase) Publish(ctx context.Context, topicName, queueID string, payload []byte, key string, headers map[string]string) (*domain.Message, error) {
	if len(payload) > u.maxSize {
		return nil, ErrMessageTooLarge
	}
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
	msg := &domain.Message{
		ID:        genID(),
		TopicName: topicName,
		QueueID:   queueID,
		Payload:   payload,
		Key:       key,
		Headers:   headers,
		CreatedAt: time.Now(),
	}
	if err := u.messages.Append(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func genID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
