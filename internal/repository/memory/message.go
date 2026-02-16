package memory

import (
	"context"
	"sync"

	"queue-service/internal/domain"
)

type messageRepo struct {
	mu       sync.RWMutex
	messages map[string][]*domain.Message
}

func NewMessageRepository() domain.MessageRepository {
	return &messageRepo{messages: make(map[string][]*domain.Message)}
}

func msgKey(topic, queueID string) string { return topic + "|" + queueID }

func (r *messageRepo) Append(ctx context.Context, msg *domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := msgKey(msg.TopicName, msg.QueueID)
	slice := r.messages[k]
	if slice == nil {
		slice = make([]*domain.Message, 0)
	}
	msg.Offset = int64(len(slice))
	m := *msg
	slice = append(slice, &m)
	r.messages[k] = slice
	return nil
}

func (r *messageRepo) Read(ctx context.Context, topicName, queueID string, offset, limit int) ([]*domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	k := msgKey(topicName, queueID)
	slice := r.messages[k]
	if slice == nil || offset >= len(slice) {
		return nil, nil
	}
	end := offset + limit
	if end > len(slice) {
		end = len(slice)
	}
	out := make([]*domain.Message, 0, end-offset)
	for i := offset; i < end; i++ {
		m := *slice[i]
		out = append(out, &m)
	}
	return out, nil
}

func (r *messageRepo) GetByID(ctx context.Context, topicName, queueID, messageID string) (*domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	slice := r.messages[msgKey(topicName, queueID)]
	for _, m := range slice {
		if m.ID == messageID {
			m2 := *m
			return &m2, nil
		}
	}
	return nil, domain.ErrNotFound
}
