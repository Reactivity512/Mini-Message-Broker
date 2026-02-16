package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"queue-service/internal/domain"
)

var errDeliveryNotFound = errors.New("delivery not found")

type pendingRepo struct {
	mu    sync.RWMutex
	bySub map[string]map[string]*domain.PendingDelivery
}

func NewPendingDeliveryRepository() domain.PendingDeliveryRepository {
	return &pendingRepo{bySub: make(map[string]map[string]*domain.PendingDelivery)}
}

func (r *pendingRepo) Add(ctx context.Context, subID string, pd *domain.PendingDelivery) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.bySub[subID] == nil {
		r.bySub[subID] = make(map[string]*domain.PendingDelivery)
	}
	p := *pd
	r.bySub[subID][pd.DeliveryID] = &p
	return nil
}

func (r *pendingRepo) Ack(ctx context.Context, subID, deliveryID string) (ackedOffset int64, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.bySub[subID]
	if !ok {
		return 0, errDeliveryNotFound
	}
	pd, ok := m[deliveryID]
	if !ok {
		return 0, errDeliveryNotFound
	}
	offset := pd.Message.Offset
	delete(m, deliveryID)
	if len(m) == 0 {
		delete(r.bySub, subID)
	}
	return offset, nil
}

func (r *pendingRepo) Expired(ctx context.Context, subID string, before time.Time) ([]*domain.PendingDelivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.bySub[subID]
	if !ok {
		return nil, nil
	}
	var out []*domain.PendingDelivery
	for _, pd := range m {
		if pd.ExpiresAt.Before(before) {
			p2 := *pd
			out = append(out, &p2)
		}
	}
	return out, nil
}

func (r *pendingRepo) Remove(ctx context.Context, subID, deliveryID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.bySub[subID]; ok {
		delete(m, deliveryID)
		if len(m) == 0 {
			delete(r.bySub, subID)
		}
	}
	return nil
}
