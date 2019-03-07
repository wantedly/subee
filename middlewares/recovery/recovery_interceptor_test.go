package subee_recovery

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/wantedly/subee"
	message_testing "github.com/wantedly/subee/testing"
)

func TestConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "Consumer"

	err := ConsumerInterceptor(func(ctx context.Context, p interface{}) error {
		buf.Write([]byte(fmt.Sprint(p)))
		return nil
	})(
		subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			panic(tag)
		}),
	).Consume(
		context.Background(),
		message_testing.NewFakeMessage(nil, false, false),
	)

	if err != nil {
		t.Errorf("Consume() returned %v, want nil", err)
	}

	if got, want := buf.String(), tag; got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestConsumerInterceptor_WhenError(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "Consumer"

	err := ConsumerInterceptor(func(ctx context.Context, p interface{}) error {
		buf.Write([]byte(fmt.Sprint(p)))
		return errors.New("errors")
	})(
		subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			panic(tag)
		}),
	).Consume(
		context.Background(),
		message_testing.NewFakeMessage(nil, false, false),
	)

	if err == nil {
		t.Error("Consume() returned nil, want an error")
	}

	if got, want := buf.String(), tag; got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestBatchConsumerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "BatchConsumer"

	err := BatchConsumerInterceptor(func(ctx context.Context, p interface{}) error {
		buf.Write([]byte(fmt.Sprint(p)))
		return nil
	})(
		subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			panic(tag)
		}),
	).BatchConsume(
		context.Background(),
		[]subee.Message{message_testing.NewFakeMessage(nil, false, false)},
	)

	if err != nil {
		t.Errorf("BatchConsume() returned %v, want nil", err)
	}

	if got, want := buf.String(), tag; got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestBatchConsumerInterceptor_WhenError(t *testing.T) {
	buf := &bytes.Buffer{}

	tag := "BatchConsumer"

	err := BatchConsumerInterceptor(func(ctx context.Context, p interface{}) error {
		buf.Write([]byte(fmt.Sprint(p)))
		return errors.New("errors")
	})(
		subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			panic(tag)
		}),
	).BatchConsume(
		context.Background(),
		[]subee.Message{message_testing.NewFakeMessage(nil, false, false)},
	)

	if err == nil {
		t.Error("BatchConsume() returned nil, want an error")
	}

	if got, want := buf.String(), tag; got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}
