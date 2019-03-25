package subee_zap

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/wantedly/subee"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var since = func(t time.Time) time.Duration {
	return time.Since(t)
}

// ConsumerInterceptor returns a new consumer interceptor for logging with zap.
func ConsumerInterceptor(logger *zap.Logger) subee.ConsumerInterceptor {
	return func(consumer subee.Consumer) subee.Consumer {
		return subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			msgCnt := 1

			startConsume(logger, msgCnt)

			startTime := time.Now()

			err := consumer.Consume(ctx, msg)

			endConsume(logger, since(startTime), msgCnt, err)

			return errors.WithStack(err)
		})
	}
}

// BatchConsumerInterceptor returns a new batch consumer interceptor for logging with zap.
func BatchConsumerInterceptor(logger *zap.Logger) subee.BatchConsumerInterceptor {
	return func(consumer subee.BatchConsumer) subee.BatchConsumer {
		return subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			startConsume(logger, len(msgs))

			startTime := time.Now()

			err := consumer.BatchConsume(ctx, msgs)

			endConsume(logger, since(startTime), len(msgs), err)

			return errors.WithStack(err)
		})
	}
}

func startConsume(logger *zap.Logger, msgCnt int) {
	logger.Info(
		"Start consume message.",
		zap.Int("message_count", msgCnt),
	)
}

func endConsume(logger *zap.Logger, d time.Duration, msgCnt int, err error) {
	logger.Check(level(err), "End consume message.").Write(
		zap.Error(err),
		zap.Int("message_count", msgCnt),
		zap.Duration("time", d),
	)
}

func level(err error) zapcore.Level {
	if err != nil {
		return zap.ErrorLevel
	}
	return zap.InfoLevel
}
