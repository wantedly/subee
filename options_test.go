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

func tagBatchConsumerInterceptor(tag string) BatchConsumerInterceptor {
	return func(consumer BatchConsumer) BatchConsumer {
		return BatchConsumerFunc(func(ctx context.Context, msgs []Message) error {
			writeBuffer(ctx, tag)
			return errors.WithStack(consumer.BatchConsume(ctx, msgs))
		})
	}
}

func tagBatchConsumer(tag string) BatchConsumer {
	return BatchConsumerFunc(func(ctx context.Context, msgs []Message) error {
		writeBuffer(ctx, tag)
		return nil
	})
}

func tagConsumerInterceptor(tag string) ConsumerInterceptor {
	return func(consumer Consumer) Consumer {
		return ConsumerFunc(func(ctx context.Context, msg Message) error {
			writeBuffer(ctx, tag)
			return errors.WithStack(consumer.Consume(ctx, msg))
		})
	}
}

func tagConsumer(tag string) Consumer {
	return ConsumerFunc(func(ctx context.Context, msg Message) error {
		writeBuffer(ctx, tag)
		return nil
	})
}

func TestChainConsumerInterceptors(t *testing.T) {
	tests := []struct {
		consumer     Consumer
		interceptors []ConsumerInterceptor
		want         string
	}{
		{
			consumer: tagConsumer("consumer-A\n"),
			interceptors: []ConsumerInterceptor{
				tagConsumerInterceptor("intr-A\n"),
				tagConsumerInterceptor("intr-B\n"),
				tagConsumerInterceptor("intr-C\n"),
			},
			want: "intr-A\nintr-B\nintr-C\nconsumer-A\n",
		},
		{
			consumer:     tagConsumer("consumer-A\n"),
			interceptors: []ConsumerInterceptor{},
			want:         "consumer-A\n",
		},
	}

	for _, test := range tests {
		chain := chainConsumerInterceptors(test.consumer, test.interceptors...)

		ctx := context.WithValue(context.Background(), chainContextTestKey{}, new(bytes.Buffer))
		chain.Consume(ctx, nil)

		got := ctx.Value(chainContextTestKey{}).(*bytes.Buffer).String()

		if test.want != got {
			t.Errorf("chainConsumerInterceptors output is wrong. want: %v, got: %v", test.want, got)
		}
	}
}

func TestChainBatchConsumerInterceptors(t *testing.T) {
	tests := []struct {
		consumer     BatchConsumer
		interceptors []BatchConsumerInterceptor
		want         string
	}{
		{
			consumer: tagBatchConsumer("consumer-A\n"),
			interceptors: []BatchConsumerInterceptor{
				tagBatchConsumerInterceptor("intr-A\n"),
				tagBatchConsumerInterceptor("intr-B\n"),
				tagBatchConsumerInterceptor("intr-C\n"),
			},
			want: "intr-A\nintr-B\nintr-C\nconsumer-A\n",
		},
		{
			consumer:     tagBatchConsumer("consumer-A\n"),
			interceptors: []BatchConsumerInterceptor{},
			want:         "consumer-A\n",
		},
	}

	for _, test := range tests {
		chain := chainBatchConsumerInterceptors(test.consumer, test.interceptors...)

		ctx := context.WithValue(context.Background(), chainContextTestKey{}, new(bytes.Buffer))
		chain.BatchConsume(ctx, nil)

		got := ctx.Value(chainContextTestKey{}).(*bytes.Buffer).String()

		if test.want != got {
			t.Errorf("chainBatchConsumerInterceptors output is wrong. want: %v, got: %v", test.want, got)
		}
	}
}
