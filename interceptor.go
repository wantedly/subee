package subee

import (
	"context"

	"github.com/pkg/errors"
)

// BatchConsumerFunc type is an adapter to allow the use of ordinary functions as BatchConsumer.
type BatchConsumerFunc func(context.Context, []Message) error

// BatchConsume call f(ctx, msgs)
func (f BatchConsumerFunc) BatchConsume(ctx context.Context, msgs []Message) error {
	return errors.WithStack(f(ctx, msgs))
}

// BatchConsumerInterceptor provides a hook to intercept the execution of a multiple messages consumption.
type BatchConsumerInterceptor func(BatchConsumer) BatchConsumer

// ConsumerFunc type is an adapter to allow the use of ordinary functions as Consumer.
type ConsumerFunc func(context.Context, Message) error

// Consume call f(ctx, msgs)
func (f ConsumerFunc) Consume(ctx context.Context, msg Message) error {
	return errors.WithStack(f(ctx, msg))
}

// ConsumerInterceptor provides a hook to intercept the execution of a message consumption.
type ConsumerInterceptor func(Consumer) Consumer

func chainConsumerInterceptors(consumer Consumer, interceptors ...ConsumerInterceptor) Consumer {
	if len(interceptors) == 0 {
		return consumer
	}

	for i := len(interceptors) - 1; i >= 0; i-- {
		consumer = interceptors[i](consumer)
	}
	return consumer
}

func chainBatchConsumerInterceptors(consumer BatchConsumer, interceptors ...BatchConsumerInterceptor) BatchConsumer {
	if len(interceptors) == 0 {
		return consumer
	}

	for i := len(interceptors) - 1; i >= 0; i-- {
		consumer = interceptors[i](consumer)
	}
	return consumer
}
