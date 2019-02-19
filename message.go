package subee

// Message is an interface of the subscribed message.
type Message interface {
	Acknowledger
	Data() []byte
}

// Acknowledger is an interface to send ack or nack.
type Acknowledger interface {
	Ack()
	Nack()
}
