package subee

import (
	"context"
)

type loggerContextKey struct{}

// GetLogger return Logger implementation set in the context.
func GetLogger(ctx context.Context) Logger {
	return ctx.Value(loggerContextKey{}).(Logger)
}

func setLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, l)
}

func queuingContext(sh StatsHandler) func() context.Context {
	return func() context.Context {
		ctx := context.Background()
		ctx = sh.TagProcess(ctx, &BeginTag{})
		ctx = sh.TagProcess(ctx, &EnqueueTag{})
		return ctx
	}
}
