package nrsubee

type Config struct {
	TransactionName     string
	ConsumeSegmentName  string
	QueueingSegmentName string
}

func DefaultConfig() *Config {
	return &Config{
		TransactionName:     "Subee",
		ConsumeSegmentName:  "Consume",
		QueueingSegmentName: "Message Queueing",
	}
}

func (c *Config) apply(opts []Option) {
	for _, f := range opts {
		f(c)
	}
}

type Option func(*Config)

func WithTransactionName(name string) Option {
	return func(c *Config) {
		c.TransactionName = name
	}
}
