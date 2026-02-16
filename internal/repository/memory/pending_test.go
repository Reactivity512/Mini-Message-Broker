package memory

import (
	"context"
	"testing"
	"time"

	"queue-service/internal/domain"
)

func TestPendingDeliveryRepository_Add_Ack_Expired(t *testing.T) {
	ctx := context.Background()
	r := NewPendingDeliveryRepository()

	pd := &domain.PendingDelivery{
		Message:    domain.Message{ID: "m1", TopicName: "t", QueueID: "0", Offset: 0},
		ExpiresAt:  time.Now().Add(10 * time.Second),
		DeliveryID: "del-1",
	}
	err := r.Add(ctx, "sub-1", pd)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	offset, err := r.Ack(ctx, "sub-1", "del-1")
	if err != nil {
		t.Fatalf("Ack: %v", err)
	}
	if offset != 0 {
		t.Errorf("acked offset want 0, got %d", offset)
	}

	_, err = r.Ack(ctx, "sub-1", "del-1")
	if err == nil {
		t.Error("double ack should error")
	}
}

func TestPendingDeliveryRepository_Expired(t *testing.T) {
	ctx := context.Background()
	r := NewPendingDeliveryRepository()
	now := time.Now()
	if err := r.Add(ctx, "sub-1", &domain.PendingDelivery{
		Message:    domain.Message{Offset: 1},
		ExpiresAt:  now.Add(-time.Second),
		DeliveryID: "del-expired",
	}); err != nil {
		t.Error(err)
	}

	if err := r.Add(ctx, "sub-1", &domain.PendingDelivery{
		Message:    domain.Message{Offset: 2},
		ExpiresAt:  now.Add(time.Hour),
		DeliveryID: "del-future",
	}); err != nil {
		t.Error(err)
	}

	expired, _ := r.Expired(ctx, "sub-1", now)
	if len(expired) != 1 {
		t.Fatalf("want 1 expired, got %d", len(expired))
	}
	if expired[0].DeliveryID != "del-expired" {
		t.Errorf("got %s", expired[0].DeliveryID)
	}
}
