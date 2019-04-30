package consumer

import (
	"context"
	"errors"

	e "example.com/a/b/d"
)

// AuthorBatchConsumer is a batch consumer interface for e.Author.
type AuthorBatchConsumer interface {
	BatchConsume(context.Context, []*e.Author) error
}

// NewAuthorBatchConsumer creates a new consumer instance.
func NewAuthorBatchConsumer() AuthorBatchConsumer {
	return &authorBatchConsumerImpl{}
}

type authorBatchConsumerImpl struct{}

func (c *authorBatchConsumerImpl) BatchConsume(ctx context.Context, msgs []*e.Author) error {
	return errors.New("BatchConsume() has not been implemented yet")
}

