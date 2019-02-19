package subee

// Message is an interface of the subscribed message.
type Message interface {
	Ack()
	Nack()
	Data() []byte
}
