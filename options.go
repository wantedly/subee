package subee

import "time"

// Option configures Engine.
type Option func(*Engine)

// WithMultiMessagesConsumerInterceptors returns an Option that sets the MultiMessageConsumerInterceptor implementations(s).
// Interceptors are called in order of addition.
// e.g) interceptor1, interceptor2, interceptor3 => interceptor1 => interceptor2 => interceptor3 => MultiMessageConsumer.Consume
func WithMultiMessagesConsumerInterceptors(interceptors ...MultiMessagesConsumerInterceptor) Option {
	return func(e *Engine) {
		e.mmConsumer = chainMultiMessagesConsumerInterceptors(e.mmConsumer, interceptors...)
	}
}

func chainMultiMessagesConsumerInterceptors(consumer MultiMessagesConsumer, interceptors ...MultiMessagesConsumerInterceptor) MultiMessagesConsumer {
	if len(interceptors) == 0 {
		return consumer
	}

	for i := len(interceptors) - 1; i >= 0; i-- {
		consumer = interceptors[i](consumer)
	}
	return consumer
}

// WithLogger returns an Option that sets the Logger implementation.
func WithLogger(logger Logger) Option {
	return func(e *Engine) {
		e.logger = logger
	}
}

// WithStatsHandler returns an Option that sets the StatsHandler implementation.
func WithStatsHandler(sh StatsHandler) Option {
	return func(e *Engine) {
		e.sh = sh
	}
}

// WithChunkSize returns an Option that sets the maximum chunked message size per consuming.
func WithChunkSize(size int) Option {
	return func(e *Engine) {
		if size > 0 {
			e.scfg.ChunkSize = size
		}
	}
}

// WithRetryCount returns an Option that sets the maximum retry count when failed to receive messages.
func WithRetryCount(count int) Option {
	return func(e *Engine) {
		if count > 0 {
			e.scfg.MaxRetryCount = count
		}
	}
}

// WithFlushInterval returns an Option that sets the maximum flush time interval to receive message.
func WithFlushInterval(inval time.Duration) Option {
	return func(e *Engine) {
		if inval > 0 {
			e.scfg.FlushInterval = inval
		}
	}
}
