package subee_zap

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/wantedly/subee"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SingleMessageConsumerInterceptor returns a new single message consumer interceptor for logging with zap.
func SingleMessageConsumerInterceptor(logger *zap.Logger) subee.SingleMessageConsumerInterceptor {
	return func(consumer subee.SingleMessageConsumer) subee.SingleMessageConsumer {
		return subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
			msgCnt := 1

			startConsume(logger, msgCnt)

			startTime := time.Now()

			err := consumer.Consume(ctx, msg)

			endConsume(logger, time.Since(startTime), msgCnt, err)

			return errors.WithStack(err)
		})
	}
}

// MultiMessagesConsumerInterceptor returns a new multi messages consumer interceptor for logging with zap.
func MultiMessagesConsumerInterceptor(logger *zap.Logger) subee.MultiMessagesConsumerInterceptor {
	return func(consumer subee.MultiMessagesConsumer) subee.MultiMessagesConsumer {
		return subee.MultiMessagesConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
			startConsume(logger, len(msgs))

			startTime := time.Now()

			err := consumer.Consume(ctx, msgs)

			endConsume(logger, time.Since(startTime), len(msgs), err)

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
