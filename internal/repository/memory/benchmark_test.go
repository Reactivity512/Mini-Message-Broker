package memory

import (
	"context"
	"testing"
	"time"

	"queue-service/internal/domain"
)

func BenchmarkMessageRepository_Append(b *testing.B) {
	ctx := context.Background()
	r := NewMessageRepository()
	msg := &domain.Message{
		ID:        "id",
		TopicName: "t",
		QueueID:   "0",
		Payload:   []byte("hello"),
		CreatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Append(ctx, msg)
	}
}

func BenchmarkMessageRepository_AppendParallel(b *testing.B) {
	ctx := context.Background()
	r := NewMessageRepository()
	msg := &domain.Message{
		ID:        "id",
		TopicName: "t",
		QueueID:   "0",
		Payload:   []byte("hello"),
		CreatedAt: time.Now(),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = r.Append(ctx, msg)
		}
	})
}

func BenchmarkMessageRepository_Read(b *testing.B) {
	ctx := context.Background()
	r := NewMessageRepository()
	msg := &domain.Message{ID: "id", TopicName: "t", QueueID: "0", Payload: []byte("x"), CreatedAt: time.Now()}
	for i := 0; i < 10000; i++ {
		_ = r.Append(ctx, msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Read(ctx, "t", "0", 0, 100)
	}
}
