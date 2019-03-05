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

func TestConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "Consumer"

	ConsumerInterceptor(writeHandler(buf))(
		subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
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

func TestBatchConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "BatchConsumer"

	BatchConsumerInterceptor(writeHandler(buf))(
		subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			panic(tag)
			return nil
		}),
	).BatchConsume(
		context.Background(),
		[]subee.Message{message_testing.NewFakeMessage(nil, false, false)},
	)

	if got, want := buf.String(), tag; got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}
