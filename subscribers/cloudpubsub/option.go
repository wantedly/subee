package cloudpubsub

import (
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type Config struct {
	ClientOpts      []option.ClientOption
	ReceiveSettings pubsub.ReceiveSettings
}

func (c *Config) apply(opts []Option) {
	for _, f := range opts {
		f(c)
	}
}

// Option is subscriber Option
type Option func(*Config)

// WithClientOptions returns an Option that set option.ClientOption implementation(s).
func WithClientOptions(opts ...option.ClientOption) Option {
	return func(c *Config) {
		c.ClientOpts = opts
	}
}

// WithReceiveSettings returns an Option that set pubsub.ReceiveSettings to the pubsub.Subscription.
func WithReceiveSettings(cfg pubsub.ReceiveSettings) Option {
	return func(c *Config) {
		c.ReceiveSettings = cfg
	}
}
