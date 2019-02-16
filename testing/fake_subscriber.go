package testing

import (
	"context"

	"github.com/wantedly/subee"
)

// FakeSubscriber implements Subscriber interface for testing.
type FakeSubscriber struct {
	err   error
	Close func()
	msgCh chan subee.Message
}

// NewFakeSubscriber creates a new FakeSubscriber instance.
func NewFakeSubscriber() *FakeSubscriber {
	return &FakeSubscriber{
		msgCh: make(chan subee.Message),
	}
}

// Subscribe implements Subscriber.Subscribe.
func (s *FakeSubscriber) Subscribe(ctx context.Context, f func(subee.Message)) error {
	ctx, s.Close = context.WithCancel(ctx)
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case m := <-s.msgCh:
			f(m)
		}
	}
	return s.err
}

// AddMessage simulates receiving a new message.
func (s *FakeSubscriber) AddMessage(m subee.Message) {
	s.msgCh <- m
}

// CloseWithError closes Subscriber with an error.
func (s *FakeSubscriber) CloseWithError(err error) {
	s.err = err
	s.Close()
}
