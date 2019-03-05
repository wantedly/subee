package subee

import "time"

// Option configures Config.
type Option func(*Config)

// WithConsumerInterceptors returns an Option that sets the ConsumerInterceptor implementations(s).
// Interceptors are called in order of addition.
// e.g) interceptor1, interceptor2, interceptor3 => interceptor1 => interceptor2 => interceptor3 => Consumer.Consume
func WithConsumerInterceptors(interceptors ...ConsumerInterceptor) Option {
	return func(c *Config) {
		c.Consumer = chainConsumerInterceptors(c.Consumer, interceptors...)
	}
}

// WithBatchConsumerInterceptors returns an Option that sets the BatchConsumerInterceptor implementations(s).
// Interceptors are called in order of addition.
// e.g) interceptor1, interceptor2, interceptor3 => interceptor1 => interceptor2 => interceptor3 => BatchConsumer.Consume
func WithBatchConsumerInterceptors(interceptors ...BatchConsumerInterceptor) Option {
	return func(c *Config) {
		c.BatchConsumer = chainBatchConsumerInterceptors(c.BatchConsumer, interceptors...)
	}
}

func chainConsumerInterceptors(consumer Consumer, interceptors ...ConsumerInterceptor) Consumer {
	if len(interceptors) == 0 {
		return consumer
	}

	for i := len(interceptors) - 1; i >= 0; i-- {
		consumer = interceptors[i](consumer)
	}
	return consumer
}

func chainBatchConsumerInterceptors(consumer BatchConsumer, interceptors ...BatchConsumerInterceptor) BatchConsumer {
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
