package subee_recovery

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wantedly/subee"
)

// RecoveryHandlerFunc is a function that recovers from the panic `p`
type RecoveryHandlerFunc func(ctx context.Context, p interface{})

// SingleMessageConsumerInterceptor returns a new single message interceptor to recovery from panic.
func SingleMessageConsumerInterceptor(f RecoveryHandlerFunc) subee.SingleMessageConsumerInterceptor {
	return func(consumer subee.SingleMessageConsumer) subee.SingleMessageConsumer {
		return subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			defer func() {
				if r := recover(); r != nil {
					f(ctx, r)
				}
			}()

			return errors.WithStack(consumer.Consume(ctx, msg))
		})
	}
}

// MultiMessagesConsumerInterceptor returns a new multi messages interceptor to recovery from panic.
func MultiMessagesConsumerInterceptor(f RecoveryHandlerFunc) subee.MultiMessagesConsumerInterceptor {
	return func(consumer subee.MultiMessagesConsumer) subee.MultiMessagesConsumer {
		return subee.MultiMessagesConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			defer func() {
				if r := recover(); r != nil {
					f(ctx, r)
				}
			}()

			return errors.WithStack(consumer.Consume(ctx, msgs))
		})
	}
}
