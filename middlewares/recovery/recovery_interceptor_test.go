package subee_recovery

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/wantedly/subee"
	message_testing "github.com/wantedly/subee/testing"
)

func writeHandler(buf *bytes.Buffer) RecoveryHandlerFunc {
	return func(ctx context.Context, p interface{}) {
		buf.Write([]byte(fmt.Sprint(p)))
	}
}

func TestSingleMessageConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "SingleMessageConsumer"

	SingleMessageConsumerInterceptor(writeHandler(buf))(
		subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			panic(tag)
			return nil
		}),
	).Consume(
		context.Background(),
		message_testing.NewFakeMessage(nil, false, false),
	)

	if got, want := buf.String(), tag; got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestMultiMessagesConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "MultiMessagesConsumer"

	SingleMessageConsumerInterceptor(writeHandler(buf))(
		subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			panic(tag)
			return nil
		}),
	).Consume(
		context.Background(),
		message_testing.NewFakeMessage(nil, false, false),
	)

	if got, want := buf.String(), tag; got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}
