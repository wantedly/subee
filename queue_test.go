package subee

import (
	"context"
	"testing"
	"time"
)

func TestCreateBufferedQueue(t *testing.T) {
	type fakeMessage struct{ Message }

	inCh, outCh := createBufferedQueue(
		context.Background,
		3,
		4*time.Millisecond,
	)

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

	for i, n := range []int{2, 3, 1, 2} {
		out := <-outCh

		if got, want := out.Count(), n; got != want {
			t.Errorf("Item[%d] has %d messages, want %d", i, got, want)
		}
	}

	_, ok := <-outCh

	if ok {
		t.Error("out channel should close")
	}
}
