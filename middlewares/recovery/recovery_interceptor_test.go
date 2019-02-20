package subee_recovery

import (
	"bytes"
	"context"
	"testing"

	"github.com/wantedly/subee"
	message_testing "github.com/wantedly/subee/testing"
)

func writeHandler(buf *bytes.Buffer, tag string) RecoveryHandlerFunc {
	return func(ctx context.Context, p interface{}) {
		buf.Write([]byte(tag))
	}
}

func TestSingleMessageConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "SingleMessageConsumer"

	SingleMessageConsumerInterceptor(writeHandler(buf, tag))(
		subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			panic("occurs panic")
			return nil
		}),
	).Consume(
		context.Background(),
		message_testing.NewFakeMessage(nil, false, false),
	)

	if got, want := buf.String(), tag; got != want {
		t.Errorf("want:%v, but: %v", want, got)
	}
}

func TestMultiMessagesConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "MultiMessagesConsumer"

	SingleMessageConsumerInterceptor(writeHandler(buf, tag))(
		subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			panic("occurs panic")
			return nil
		}),
	).Consume(
		context.Background(),
		message_testing.NewFakeMessage(nil, false, false),
	)

	if got, want := buf.String(), tag; got != want {
		t.Errorf("want:%v, but: %v", want, got)
	}
}
