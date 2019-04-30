package consumer

import (
	"context"
	"errors"

	"example.com/a/b/c"
)

// BookBatchConsumer is a batch consumer interface for c.Book.
type BookBatchConsumer interface {
	BatchConsume(context.Context, []*c.Book) error
}

// NewBookBatchConsumer creates a new consumer instance.
func NewBookBatchConsumer() BookBatchConsumer {
	return &bookBatchConsumerImpl{}
}

type bookBatchConsumerImpl struct{}

func (c *bookBatchConsumerImpl) BatchConsume(ctx context.Context, msgs []*c.Book) error {
	return errors.New("BatchConsume() has not been implemented yet")
}

