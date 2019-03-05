package subee

import "context"

// BatchConsumer represents an interface that consume multiple messages.
type BatchConsumer interface {
	BatchConsume(context.Context, []Message) error
}

// Consumer represents an interface that consume single message.
type Consumer interface {
	Consume(context.Context, Message) error
}
