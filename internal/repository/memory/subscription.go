package memory

import (
	"context"
	"sync"

	"queue-service/internal/domain"
)

type subscriptionRepo struct {
	mu   sync.RWMutex
	byID map[string]*domain.Subscription
	byTg map[string]*domain.Subscription
}

func NewSubscriptionRepository() domain.SubscriptionRepository {
	return &subscriptionRepo{
		byID: make(map[string]*domain.Subscription),
		byTg: make(map[string]*domain.Subscription),
	}
}

func tgKey(topic, group string) string { return topic + "|" + group }

func (r *subscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := *sub
	r.byID[sub.ID] = &s
	r.byTg[tgKey(sub.TopicName, sub.ConsumerGroup)] = &s
	return nil
}

func (r *subscriptionRepo) Get(ctx context.Context, id string) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	s2 := *s
	return &s2, nil
}

func (r *subscriptionRepo) GetByTopicAndGroup(ctx context.Context, topicName, consumerGroup string) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byTg[tgKey(topicName, consumerGroup)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	s2 := *s
	return &s2, nil
}

func (r *subscriptionRepo) ListByTopic(ctx context.Context, topicName string) ([]*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Subscription
	for _, s := range r.byID {
		if s.TopicName == topicName {
			s2 := *s
			out = append(out, &s2)
		}
	}
	return out, nil
}

func (r *subscriptionRepo) List(ctx context.Context) ([]*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Subscription, 0, len(r.byID))
	for _, s := range r.byID {
		s2 := *s
		out = append(out, &s2)
	}
	return out, nil
}

func (r *subscriptionRepo) AdvanceOffset(ctx context.Context, id string, offset int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.byID[id]; ok {
		s.Offset = offset
	}
	return nil
}
