package testing

import "sync/atomic"

// FakeMessage implements Message interface.
type FakeMessage struct {
	data     []byte
	metadata map[string]string
	acked    int32
	nacked   int32
}

// NewFakeMessage creates a new FakeMessage object.
func NewFakeMessage(data []byte, acked, nacked bool) *FakeMessage {
	m := &FakeMessage{data: data}
	if acked {
		m.Ack()
	}
	if nacked {
		m.Nack()
	}
	return m
}

// Data returns the message'm payload.
func (m *FakeMessage) Data() []byte { return m.data }

// Metadata returns the message'm payload.
func (m *FakeMessage) Metadata() map[string]string { return m.metadata }

// Ack changes the mssage state to 'Acked'.
func (m *FakeMessage) Ack() { atomic.StoreInt32(&m.acked, 1) }

// Nack changes the mssage state to 'Nacked'.
func (m *FakeMessage) Nack() { atomic.StoreInt32(&m.nacked, 1) }

// Acked returned true if the message has been acked.
func (m *FakeMessage) Acked() bool { return atomic.LoadInt32(&m.acked) == 1 }

// Nacked returned true if the message has been nacked.
func (m *FakeMessage) Nacked() bool { return atomic.LoadInt32(&m.nacked) == 1 }
