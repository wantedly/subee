package subee

import (
	"context"
	"time"
)

type (
	loggerContextKey     struct{}
	enqueuedAtContextKey struct{}
)

// GetLogger return Logger implementation set in the context.
func GetLogger(ctx context.Context) Logger {
	return ctx.Value(loggerContextKey{}).(Logger)
}

func setLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, l)
}

func getEnqueuedAt(ctx context.Context) time.Time {
	return ctx.Value(enqueuedAtContextKey{}).(time.Time)
}

func setEnqueuedAt(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, enqueuedAtContextKey{}, t)
}
