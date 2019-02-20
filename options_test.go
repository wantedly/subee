package subee

import (
	"bytes"
	"context"
	"testing"

	"github.com/pkg/errors"
)

type chainContextTestKey struct{}

func writeBuffer(ctx context.Context, tag string) {
	ctx.Value(chainContextTestKey{}).(*bytes.Buffer).Write([]byte(tag))
}

func tagMultiMessagesConsumerInterceptor(tag string) MultiMessagesConsumerInterceptor {
	return func(consumer MultiMessagesConsumer) MultiMessagesConsumer {
		return MultiMessagesConsumerFunc(func(ctx context.Context, msgs []Message) error {
			writeBuffer(ctx, tag)
			return errors.WithStack(consumer.Consume(ctx, msgs))
		})
	}
}

func tagMultiMessagesConsumer(tag string) MultiMessagesConsumer {
	return MultiMessagesConsumerFunc(func(ctx context.Context, msgs []Message) error {
		writeBuffer(ctx, tag)
		return nil
	})
}

func tagSingleMessageConsumerInterceptor(tag string) SingleMessageConsumerInterceptor {
	return func(consumer SingleMessageConsumer) SingleMessageConsumer {
		return SingleMessageConsumerFunc(func(ctx context.Context, msg Message) error {
			writeBuffer(ctx, tag)
			return errors.WithStack(consumer.Consume(ctx, msg))
		})
	}
}

func tagSingleMessageConsumer(tag string) SingleMessageConsumer {
	return SingleMessageConsumerFunc(func(ctx context.Context, msg Message) error {
		writeBuffer(ctx, tag)
		return nil
	})
}

func TestChainSingleMessageConsumerInterceptors(t *testing.T) {
	tests := []struct {
		consumer     SingleMessageConsumer
		interceptors []SingleMessageConsumerInterceptor
		want         string
	}{
		{
			consumer: tagSingleMessageConsumer("consumer-A\n"),
			interceptors: []SingleMessageConsumerInterceptor{
				tagSingleMessageConsumerInterceptor("intr-A\n"),
				tagSingleMessageConsumerInterceptor("intr-B\n"),
				tagSingleMessageConsumerInterceptor("intr-C\n"),
			},
			want: "intr-A\nintr-B\nintr-C\nconsumer-A\n",
		},
		{
			consumer:     tagSingleMessageConsumer("consumer-A\n"),
			interceptors: []SingleMessageConsumerInterceptor{},
			want:         "consumer-A\n",
		},
	}

	for _, test := range tests {
		chain := chainSingleMessageConsumerInterceptors(test.consumer, test.interceptors...)

		ctx := context.WithValue(context.Background(), chainContextTestKey{}, new(bytes.Buffer))
		chain.Consume(ctx, nil)

		got := ctx.Value(chainContextTestKey{}).(*bytes.Buffer).String()

		if test.want != got {
			t.Errorf("chainSingleMessageConsumerInterceptors output is wrong. want: %v, got: %v", test.want, got)
		}
	}
}

func TestChainMultiMessagesConsumerInterceptors(t *testing.T) {
	tests := []struct {
		consumer     MultiMessagesConsumer
		interceptors []MultiMessagesConsumerInterceptor
		want         string
	}{
		{
			consumer: tagMultiMessagesConsumer("consumer-A\n"),
			interceptors: []MultiMessagesConsumerInterceptor{
				tagMultiMessagesConsumerInterceptor("intr-A\n"),
				tagMultiMessagesConsumerInterceptor("intr-B\n"),
				tagMultiMessagesConsumerInterceptor("intr-C\n"),
			},
			want: "intr-A\nintr-B\nintr-C\nconsumer-A\n",
		},
		{
			consumer:     tagMultiMessagesConsumer("consumer-A\n"),
			interceptors: []MultiMessagesConsumerInterceptor{},
			want:         "consumer-A\n",
		},
	}

	for _, test := range tests {
		chain := chainMultiMessagesConsumerInterceptors(test.consumer, test.interceptors...)

		ctx := context.WithValue(context.Background(), chainContextTestKey{}, new(bytes.Buffer))
		chain.Consume(ctx, nil)

		got := ctx.Value(chainContextTestKey{}).(*bytes.Buffer).String()

		if test.want != got {
			t.Errorf("chainMultiMessagesConsumerInterceptors output is wrong. want: %v, got: %v", test.want, got)
		}
	}
}
