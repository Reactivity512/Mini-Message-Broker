package memory

import (
	"context"
	"testing"
	"time"

	"queue-service/internal/domain"
)

func TestMessageRepository_Append_Read(t *testing.T) {
	ctx := context.Background()
	r := NewMessageRepository()

	msg := &domain.Message{
		ID:        "id1",
		TopicName: "t1",
		QueueID:   "0",
		Payload:   []byte("hello"),
		Key:       "k1",
		Headers:   map[string]string{"h": "v"},
		CreatedAt: time.Now(),
	}
	err := r.Append(ctx, msg)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if msg.Offset != 0 {
		t.Errorf("first message offset want 0, got %d", msg.Offset)
	}

	msgs, err := r.Read(ctx, "t1", "0", 0, 10)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("want 1 message, got %d", len(msgs))
	}
	if msgs[0].ID != "id1" || string(msgs[0].Payload) != "hello" || msgs[0].Offset != 0 {
		t.Errorf("got %+v", msgs[0])
	}
}

func TestMessageRepository_Read_offset_limit(t *testing.T) {
	ctx := context.Background()
	r := NewMessageRepository()
	for i := 0; i < 5; i++ {
		_ = r.Append(ctx, &domain.Message{ID: "id", TopicName: "t", QueueID: "0", Payload: []byte("x"), CreatedAt: time.Now()})
	}

	msgs, _ := r.Read(ctx, "t", "0", 2, 2)
	if len(msgs) != 2 {
		t.Fatalf("want 2, got %d", len(msgs))
	}
	if msgs[0].Offset != 2 || msgs[1].Offset != 3 {
		t.Errorf("offsets: %d %d", msgs[0].Offset, msgs[1].Offset)
	}

	empty, _ := r.Read(ctx, "t", "0", 10, 5)
	if len(empty) != 0 {
		t.Errorf("offset past end should return empty, got %d", len(empty))
	}
}

func TestMessageRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	r := NewMessageRepository()
	_ = r.Append(ctx, &domain.Message{ID: "id-a", TopicName: "t", QueueID: "0", Payload: []byte("a"), CreatedAt: time.Now()})

	got, err := r.GetByID(ctx, "t", "0", "id-a")
	if err != nil || got.ID != "id-a" {
		t.Fatalf("GetByID: err=%v got=%+v", err, got)
	}
	_, err = r.GetByID(ctx, "t", "0", "nonexistent")
	if err != domain.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
