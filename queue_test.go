package subee

import (
	"context"
	"testing"
	"time"
)

type fakeMessage struct {
	Message
}

func queuing(inCh chan<- Message) {
	go func() {
		inCh <- &fakeMessage{}
		inCh <- &fakeMessage{}
		time.Sleep(6 * time.Millisecond)
		inCh <- &fakeMessage{}
		inCh <- &fakeMessage{}
		inCh <- &fakeMessage{}
		inCh <- &fakeMessage{}
		time.Sleep(6 * time.Millisecond)
		inCh <- &fakeMessage{}
		inCh <- &fakeMessage{}
		close(inCh)
	}()
}

func TestQueue(t *testing.T) {
	inCh, outCh := createQueue(
		context.Background,
	)

	queuing(inCh)

	for i, n := range []int{1, 1, 1, 1, 1, 1, 1, 1} {
		out := <-outCh

		if got, want := out.Count(), n; got != want {
			t.Errorf("Item[%d] has %d messages, want %d", i, got, want)
		}

		if m, ok := out.(*singleMessage); !ok {
			t.Errorf("Item[%d] is %T type, want *subee.singleMessage type", i, m)
		}
	}

	_, ok := <-outCh

	if ok {
		t.Error("out channel should close")
	}
}

func TestCreateBufferedQueue(t *testing.T) {
	inCh, outCh := createBufferedQueue(
		context.Background,
		3,
		4*time.Millisecond,
	)

	queuing(inCh)

	for i, n := range []int{2, 3, 1, 2} {
		out := <-outCh

		if got, want := out.Count(), n; got != want {
			t.Errorf("Item[%d] has %d messages, want %d", i, got, want)
		}

		if m, ok := out.(*multiMessages); !ok {
			t.Errorf("Item[%d] is %T type, want *subee.MultiMessagesConsumer type", i, m)
		}
	}

	_, ok := <-outCh

	if ok {
		t.Error("out channel should close")
	}
}
