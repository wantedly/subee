package consumer

import (
	"context"
	"errors"

	"github.com/wantedly/subee"
)

// BookConsumer is a consumer interface.
type BookConsumer subee.SingleMessageConsumer

// NewBookConsumer creates a new consumer instance.
func NewBookConsumer() BookConsumer {
	return &bookConsumerImpl{}
}

type bookConsumerImpl struct{}

func (c *bookConsumerImpl) Consume(ctx context.Context, msg subee.Message) error {
	return errors.New("Consume() has not been implemented yet")
}

