package cloudpubsub

import "cloud.google.com/go/pubsub"

// Message is wrapps *pubsub.Message
type Message struct {
	*pubsub.Message
}

// Data returns *pubsub.Message.Data
func (m *Message) Data() []byte { return m.Message.Data }

// Metadata returns message attributes.
func (m *Message) Metadata() map[string]string { return m.Message.Attributes }
