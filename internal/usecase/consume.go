package usecase

import (
	"context"
	"time"

	"queue-service/internal/domain"
)

type ConsumeUseCase struct {
	subs     domain.SubscriptionRepository
	messages domain.MessageRepository
	pending  domain.PendingDeliveryRepository
}

func NewConsumeUseCase(
	subs domain.SubscriptionRepository,
	messages domain.MessageRepository,
	pending domain.PendingDeliveryRepository,
) *ConsumeUseCase {
	return &ConsumeUseCase{
		subs:     subs,
		messages: messages,
		pending:  pending,
	}
}

// Функция Consume получает сообщения для подписки.
// Для случаев, когда сообщение получено хотя бы один раз, отслеживаются ожидающие сообщения, и возвращается delivery_id для подтверждения.
func (u *ConsumeUseCase) Consume(ctx context.Context, subscriptionID string, maxMessages int) ([]*domain.Message, error) {
	if maxMessages <= 0 {
		maxMessages = 1
	}
	sub, err := u.subs.Get(ctx, subscriptionID)
	if err != nil || sub == nil {
		return nil, ErrSubscriptionNotFound
	}

	// Сначала повторно доставить сообщения, срок действия которых истек хотя бы один раз.
	if sub.DeliveryGuarantee == domain.AtLeastOnce {
		expired, _ := u.pending.Expired(ctx, subscriptionID, time.Now())
		if len(expired) > 0 {
			out := make([]*domain.Message, 0, len(expired))
			for _, pd := range expired {
				msg := pd.Message
				msg.DeliveryID = pd.DeliveryID
				out = append(out, &msg)
			}
			return out, nil
		}
	}

	// Чтение из очереди по смещению подписки
	msgs, err := u.messages.Read(ctx, sub.TopicName, sub.QueueID, int(sub.Offset), maxMessages)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, nil
	}

	if sub.DeliveryGuarantee == domain.AtMostOnce {
		// Немедленно выполнить смещение
		lastOffset := msgs[len(msgs)-1].Offset
		_ = u.subs.AdvanceOffset(ctx, subscriptionID, lastOffset+1)
		return msgs, nil
	}

	// AtLeastOnce: добавить в список ожидающих и установить DeliveryID
	deliveryID := genID()
	for i := range msgs {
		msgs[i].DeliveryID = deliveryID + "-" + msgs[i].ID
		pd := &domain.PendingDelivery{
			Message:    *msgs[i],
			ExpiresAt:  time.Now().Add(sub.AckTimeout),
			DeliveryID: msgs[i].DeliveryID,
		}
		_ = u.pending.Add(ctx, sub.ID, pd)
	}
	return msgs, nil
}

func (u *ConsumeUseCase) Ack(ctx context.Context, subscriptionID, deliveryID string) error {
	if _, err := u.subs.Get(ctx, subscriptionID); err != nil {
		return ErrSubscriptionNotFound
	}
	ackedOffset, err := u.pending.Ack(ctx, subscriptionID, deliveryID)
	if err != nil {
		return err
	}
	return u.subs.AdvanceOffset(ctx, subscriptionID, ackedOffset+1)
}
