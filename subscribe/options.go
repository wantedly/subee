package subscribe

import "google.golang.org/api/option"

// Option is subscriber Option
type Option func(s *subscriber)

// WithRetry returns an Option that set Retry implementation.
func WithRetry(retry Retry) Option {
	return func(s *subscriber) {
		s.retry = retry
	}
}

// WithClientOptions returns an Option that set option.ClientOption implementation(s).
func WithClientOptions(ops ...option.ClientOption) Option {
	return func(s *subscriber) {
		s.pubsubClientOption = ops
	}
}
