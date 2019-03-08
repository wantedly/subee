package subee_recovery

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wantedly/subee"
)

// RecoveryHandlerFunc is a function that recovers from the panic `p`
type RecoveryHandlerFunc func(ctx context.Context, p interface{})

// ConsumerInterceptor returns a new consumer interceptor to recovery from panic.
func ConsumerInterceptor(f RecoveryHandlerFunc) subee.ConsumerInterceptor {
	return func(consumer subee.Consumer) subee.Consumer {
		return subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			defer func() {
				if r := recover(); r != nil {
					f(ctx, r)
				}
			}()

			return errors.WithStack(consumer.Consume(ctx, msg))
		})
	}
}

// BatchConsumerInterceptor returns a new batch consumer interceptor to recovery from panic.
func BatchConsumerInterceptor(f RecoveryHandlerFunc) subee.BatchConsumerInterceptor {
	return func(consumer subee.BatchConsumer) subee.BatchConsumer {
		return subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			defer func() {
				if r := recover(); r != nil {
					f(ctx, r)
				}
			}()

			return errors.WithStack(consumer.BatchConsume(ctx, msgs))
		})
	}
}
