package memory

import (
	"context"
	"sync"

	"queue-service/internal/domain"
)

type topicRepo struct {
	mu     sync.RWMutex
	topics map[string]*domain.Topic
}

func NewTopicRepository() domain.TopicRepository {
	return &topicRepo{topics: make(map[string]*domain.Topic)}
}

func (r *topicRepo) Create(ctx context.Context, topic *domain.Topic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.topics[topic.Name]; exists {
		return domain.ErrExists
	}
	t := *topic
	r.topics[topic.Name] = &t
	return nil
}

func (r *topicRepo) Get(ctx context.Context, name string) (*domain.Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.topics[name]
	if !ok {
		return nil, domain.ErrNotFound
	}
	t2 := *t
	return &t2, nil
}

func (r *topicRepo) Delete(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.topics, name)
	return nil
}

func (r *topicRepo) List(ctx context.Context) ([]*domain.Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Topic, 0, len(r.topics))
	for _, t := range r.topics {
		t2 := *t
		out = append(out, &t2)
	}
	return out, nil
}
