package memory

import (
	"context"
	"sync"

	"queue-service/internal/domain"
)

type queueRepo struct {
	mu     sync.RWMutex
	queues map[string]map[string]*domain.Queue
}

func NewQueueRepository() domain.QueueRepository {
	return &queueRepo{queues: make(map[string]map[string]*domain.Queue)}
}

func (r *queueRepo) Create(ctx context.Context, queue *domain.Queue) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.queues[queue.TopicName] == nil {
		r.queues[queue.TopicName] = make(map[string]*domain.Queue)
	}
	if _, exists := r.queues[queue.TopicName][queue.QueueID]; exists {
		return domain.ErrExists
	}
	q := *queue
	r.queues[queue.TopicName][queue.QueueID] = &q
	return nil
}

func (r *queueRepo) Get(ctx context.Context, topicName, queueID string) (*domain.Queue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.queues[topicName]; ok {
		if q, ok := m[queueID]; ok {
			q2 := *q
			return &q2, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *queueRepo) ListByTopic(ctx context.Context, topicName string) ([]*domain.Queue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.queues[topicName]
	if !ok {
		return nil, nil
	}
	out := make([]*domain.Queue, 0, len(m))
	for _, q := range m {
		q2 := *q
		out = append(out, &q2)
	}
	return out, nil
}
