package subee

import (
	"context"

	"github.com/pkg/errors"
)

// MultiMessagesConsumer represents an interface that consume multiple messages.
type MultiMessagesConsumer interface {
	Consume(context.Context, []Message) error
}

// MultiMessagesConsumerFunc type is an adapter to allow the use of ordinary functions as MultiMessagesConsumer.
type MultiMessagesConsumerFunc func(context.Context, []Message) error

// Consume call f(ctx, msgs)
func (f MultiMessagesConsumerFunc) Consume(ctx context.Context, msgs []Message) error {
	return errors.WithStack(f(ctx, msgs))
}

// MultiMessagesConsumerInterceptor provides a hook to intercept the execution of a multiple message consumption.
type MultiMessagesConsumerInterceptor func(MultiMessagesConsumer) MultiMessagesConsumer

// SingleMessageConsumer represents an interface that consume single message.
type SingleMessageConsumer interface {
	Consume(context.Context, Message) error
}

// SingleMessageConsumerFunc type is an adapter to allow the use of ordinary functions as SingleMessageConsumer.
type SingleMessageConsumerFunc func(context.Context, Message) error

// Consume call f(ctx, msgs)
func (f SingleMessageConsumerFunc) Consume(ctx context.Context, msg Message) error {
	return errors.WithStack(f(ctx, msg))
}

// SingleMessageConsumerInterceptor provides a hook to intercept the execution of a multiple message consumption.
type SingleMessageConsumerInterceptor func(SingleMessageConsumer) SingleMessageConsumer
