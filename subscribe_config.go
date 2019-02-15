package subee

import (
	"time"
)

const (
	// DefaultChunkSize is the default maximum chunked message size per consuming.
	DefaultChunkSize = 4

	// DefaultMaxRetryCount is the default maximum retry count when failed to receive messages.
	DefaultMaxRetryCount = 10

	// DefaultFlushInterval is the default maximum flush interval to receive messages.
	DefaultFlushInterval = 10 * time.Second
)

// SubscribeConfig represents the configuration for subscriber.
type SubscribeConfig struct {
	ChunkSize int

	FlushInterval time.Duration

	MaxRetryCount int
}

// newDefaultSubscribeConfig returns the *SubscribeConfig with default value set.
func newDefaultSubscribeConfig() *SubscribeConfig {
	return &SubscribeConfig{
		ChunkSize:     DefaultChunkSize,
		MaxRetryCount: DefaultMaxRetryCount,
		FlushInterval: DefaultFlushInterval,
	}
}
