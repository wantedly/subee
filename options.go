package subee

import "time"

// Option configures Config.
type Option func(*Config)

// WithSingleMessageConsumerInterceptors returns an Option that sets the SingleMessageConsumerInterceptor implementations(s).
// Interceptors are called in order of addition.
// e.g) interceptor1, interceptor2, interceptor3 => interceptor1 => interceptor2 => interceptor3 => MultiMessageConsumer.Consume
func WithSingleMessageConsumerInterceptors(interceptors ...SingleMessageConsumerInterceptor) Option {
	return func(c *Config) {
		c.SingleMessageConsumer = chainSingleMessageConsumerInterceptors(c.SingleMessageConsumer, interceptors...)
	}
}

// WithMultiMessagesConsumerInterceptors returns an Option that sets the MultiMessageConsumerInterceptor implementations(s).
// Interceptors are called in order of addition.
// e.g) interceptor1, interceptor2, interceptor3 => interceptor1 => interceptor2 => interceptor3 => MultiMessageConsumer.Consume
func WithMultiMessagesConsumerInterceptors(interceptors ...MultiMessagesConsumerInterceptor) Option {
	return func(c *Config) {
		c.MultiMessagesConsumer = chainMultiMessagesConsumerInterceptors(c.MultiMessagesConsumer, interceptors...)
	}
}

func chainSingleMessageConsumerInterceptors(consumer SingleMessageConsumer, interceptors ...SingleMessageConsumerInterceptor) SingleMessageConsumer {
	if len(interceptors) == 0 {
		return consumer
	}

	for i := len(interceptors) - 1; i >= 0; i-- {
		consumer = interceptors[i](consumer)
	}
	return consumer
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
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithStatsHandler returns an Option that sets the StatsHandler implementation.
func WithStatsHandler(sh StatsHandler) Option {
	return func(c *Config) {
		c.StatsHandler = sh
	}
}

// WithChunkSize returns an Option that sets the maximum chunked message size per consuming.
func WithChunkSize(size int) Option {
	return func(c *Config) {
		if size > 0 {
			c.ChunkSize = size
		}
	}
}

// WithFlushInterval returns an Option that sets the maximum flush time interval to receive message.
func WithFlushInterval(inval time.Duration) Option {
	return func(c *Config) {
		if inval > 0 {
			c.FlushInterval = inval
		}
	}
}

// WithAckImmediately returns an Option that make ack messages before consuming.
func WithAckImmediately() Option {
	return func(c *Config) {
		c.AckImmediately = true
	}
}
