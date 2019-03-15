package subee

import (
	"log"
	"os"
	"time"
)

const (
	// DefaultChunkSize is the default maximum chunked message size per consuming.
	DefaultChunkSize = 4

	// DefaultFlushInterval is the default maximum flush interval to receive messages.
	DefaultFlushInterval = 10 * time.Second

	// DefaultConcurrency is the default size of goroutines which consume messages concurrently.
	DefaultConcurrency = 1
)

// Config represents the configuration for subee.
type Config struct {
	ChunkSize int

	AckImmediately bool

	FlushInterval time.Duration

	Logger Logger

	StatsHandler StatsHandler

	BatchConsumer             BatchConsumer
	BatchConsumerInterceptors []BatchConsumerInterceptor

	Consumer             Consumer
	ConsumerInterceptors []ConsumerInterceptor

	Concurrency int
}

func (c *Config) apply(opts []Option) {
	for _, f := range opts {
		f(c)
	}
}

// newDefaultConfig returns the *Config with default value set.
func newDefaultConfig() *Config {
	return &Config{
		ChunkSize:     DefaultChunkSize,
		FlushInterval: DefaultFlushInterval,
		Logger:        log.New(os.Stdout, "", log.LstdFlags),
		StatsHandler:  new(NopStatsHandler),
		Concurrency:   DefaultConcurrency,
	}
}
