package consumer

import (
	"context"
	"log"

	"your-first-cloudpubsub-app/pkg/model"
)

// EventConsumer is a consumer interface for model.Event.
type EventConsumer interface {
	Consume(context.Context, *model.Event) error
}

// NewEventConsumer creates a new consumer instance.
func NewEventConsumer(logger *log.Logger) EventConsumer {
	return &eventConsumerImpl{
		logger: logger,
	}
}

type eventConsumerImpl struct {
	logger *log.Logger
}

func (c *eventConsumerImpl) Consume(ctx context.Context, msg *model.Event) error {
	c.logger.Printf("Received event, created_at: %d", msg.CreatedAt)
	return nil
}
