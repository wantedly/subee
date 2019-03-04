package consumer

import (
	"context"
	"errors"

	e "example.com/a/b/d"
)

// AuthorConsumer is a consumer interface for e.Author.
type AuthorConsumer interface {
	Consume(context.Context, []*e.Author) error
}

// NewAuthorConsumer creates a new consumer instance.
func NewAuthorConsumer() AuthorConsumer {
	return &authorConsumerImpl{}
}

type authorConsumerImpl struct{}

func (c *authorConsumerImpl) Consume(ctx context.Context, msgs []*e.Author) error {
	return errors.New("Consume() has not been implemented yet")
}

