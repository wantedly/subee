package subscribe

import "cloud.google.com/go/pubsub"

// PubsubMessage is wrapps *pubsub.Message
type PubsubMessage struct {
	*pubsub.Message
}

// func (p *PubsubMessage) Ack()         { p.msg.Ack() }
// func (p *PubsubMessage) Nack()        { p.msg.Nack() }

// Data returns *pubsub.Message.Data
func (p *PubsubMessage) Data() []byte { return p.Message.Data }
