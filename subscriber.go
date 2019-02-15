package subee

import "context"

// Subscriber is the interface to subscribe message.
type Subscriber interface {
	Subscribe(context.Context, chan *BufferedMessage) error
}
