package subee

import (
	"context"
	"time"
)

type multiMessages struct {
	Ctx        context.Context
	Msgs       []Message
	EnqueuedAt time.Time
}

func (m *multiMessages) Ack() {
	for _, msg := range m.Msgs {
		msg.Ack()
	}
}

func (m *multiMessages) Nack() {
	for _, msg := range m.Msgs {
		msg.Nack()
	}
}

func createBufferedQueue(
	createCtx func() context.Context,
	chunkSize int,
	flushInterval time.Duration,
) (
	chan<- Message,
	<-chan *multiMessages,
) {
	inCh := make(chan Message, chunkSize)
	outCh := make(chan *multiMessages)

	go func() {
		defer close(outCh)

		for {
			msgs, opened := buffering(inCh, chunkSize, flushInterval)

			if len(msgs) > 0 {
				outCh <- &multiMessages{
					Ctx:        createCtx(),
					Msgs:       msgs,
					EnqueuedAt: time.Now(),
				}
			}

			if !opened {
				break
			}
		}
	}()

	return inCh, outCh
}

func buffering(
	msgCh <-chan Message,
	chunkSize int,
	flushInterval time.Duration,
) (msgs []Message, opened bool) {
	msgs = make([]Message, 0, chunkSize)
	opened = true

	ctx, cancel := context.WithTimeout(context.Background(), flushInterval)
	defer cancel()

	for {
		select {
		case msg, ok := <-msgCh:
			if !ok {
				opened = false
				return
			}
			msgs = append(msgs, msg)
			if len(msgs) >= chunkSize {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
