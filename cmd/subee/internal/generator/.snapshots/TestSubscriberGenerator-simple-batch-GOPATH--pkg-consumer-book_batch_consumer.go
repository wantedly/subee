package consumer

import (
	"context"
	"errors"

	"github.com/wantedly/subee"
)

// BookBatchConsumer is a consumer interface.
type BookBatchConsumer subee.BatchConsumer

// NewBookBatchConsumer creates a new consumer instance.
func NewBookBatchConsumer() BookBatchConsumer {
	return &bookBatchConsumerImpl{}
}

type bookBatchConsumerImpl struct{}

func (c *bookBatchConsumerImpl) BatchConsume(ctx context.Context, msgs []subee.Message) error {
	return errors.New("BatchConsume() has not been implemented yet")
}

