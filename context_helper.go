package subee

import (
	"context"
	"time"
)

type statsHandlerContextKey struct{}
type loggerContextKey struct{}
type subscribeConfigContextkey struct{}

// GetLogger return Logger implementation set in the context.
func GetLogger(ctx context.Context) Logger {
	return ctx.Value(loggerContextKey{}).(Logger)
}

// GetStatsHandler return StatsHandler implementation set in the context.
func GetStatsHandler(ctx context.Context) StatsHandler {
	return ctx.Value(statsHandlerContextKey{}).(StatsHandler)
}

// GetSubscribeConfig returns SubscribeConfig instance set in the context.
func GetSubscribeConfig(ctx context.Context) *SubscribeConfig {
	return ctx.Value(subscribeConfigContextkey{}).(*SubscribeConfig)
}

// EnqueueTimeContextKey is context key for storing the enqueued time in the context.
type EnqueueTimeContextKey struct{}

// BeginTimeContextKey is context key for storing the time to start receiving messages in the context
type BeginTimeContextKey struct{}

func getTimeByContextKey(ctx context.Context, key interface{}) time.Time {
	if t, ok := ctx.Value(key).(time.Time); ok {
		return t
	}
	return time.Now()
}
