package subee

import "context"

// Message is an interface of the subscribed message.
type Message interface {
	// Ack()
	// Nack()
	Data() []byte
}

// BufferedMessage wrapps context.Context and []Message.
type BufferedMessage struct {
	Ctx  context.Context
	Msgs []Message
}
