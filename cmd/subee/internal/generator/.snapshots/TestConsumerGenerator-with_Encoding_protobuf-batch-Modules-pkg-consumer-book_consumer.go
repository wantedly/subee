package consumer

import (
	"context"
	"errors"

	"example.com/a/b/c"
)

// BookConsumer is a consumer interface for c.Book.
type BookConsumer interface {
	Consume(context.Context, []*c.Book) error
}

// NewBookConsumer creates a new consumer instance.
func NewBookConsumer() BookConsumer {
	return &bookConsumerImpl{}
}

type bookConsumerImpl struct{}

func (c *bookConsumerImpl) Consume(ctx context.Context, msgs []*c.Book) error {
	return errors.New("Consume() has not been implemented yet")
}

