package subee_zap

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/wantedly/subee"
	message_testing "github.com/wantedly/subee/testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func dummyLogger(buf *bytes.Buffer) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.InfoLevel)
	return zap.New(core)
}

func dummySince() func(t time.Time) time.Duration {
	return func(t time.Time) time.Duration {
		return 0
	}
}

func TestConsumerInterceptor(t *testing.T) {
	since = dummySince()

	defer func() {
		since = func(t time.Time) time.Duration {
			return time.Since(t)
		}
	}()

	buf := &bytes.Buffer{}
	logger := dummyLogger(buf)

	want := `{"level":"INFO","msg":"Start consume message.","message_count":1}
{"level":"INFO","msg":"called single message consumer func"}
{"level":"INFO","msg":"End consume message.","message_count":1,"time":"0s"}
`

	ConsumerInterceptor(logger)(
		subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			logger.Info("called single message consumer func")
			return nil
		}),
	).Consume(
		context.Background(),
		message_testing.NewFakeMessage(nil, false, false),
	)

	if got := buf.String(); got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestBatchConsumerInterceptor(t *testing.T) {
	since = dummySince()

	defer func() {
		since = func(t time.Time) time.Duration {
			return time.Since(t)
		}
	}()

	buf := &bytes.Buffer{}
	logger := dummyLogger(buf)

	want := `{"level":"INFO","msg":"Start consume message.","message_count":2}
{"level":"INFO","msg":"called single message consumer func"}
{"level":"INFO","msg":"End consume message.","message_count":2,"time":"0s"}
`

	BatchConsumerInterceptor(logger)(
		subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			logger.Info("called single message consumer func")
			return nil
		}),
	).BatchConsume(
		context.Background(),
		[]subee.Message{
			message_testing.NewFakeMessage(nil, false, false),
			message_testing.NewFakeMessage(nil, false, false),
		},
	)

	if got := buf.String(); got != want {
		t.Errorf("\nwant:\n%sgot:\n%s", want, got)
	}
}
