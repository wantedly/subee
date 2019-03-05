package subee_test

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/wantedly/subee"
	subee_testing "github.com/wantedly/subee/testing"
)

func TestEngine(t *testing.T) {
	in := [][]byte{
		[]byte("foo"),
		[]byte("error!!"),
	}

	type TestCase struct {
		test string
		want *subee_testing.FakeMessage
		err  bool
	}
	type Result struct {
		msg *subee_testing.FakeMessage
		tc  TestCase
	}

	cases := []TestCase{
		{
			test: "acked when no error",
			want: subee_testing.NewFakeMessage([]byte("foo"), true, false),
		},
		{
			test: "nacked when errored",
			want: subee_testing.NewFakeMessage([]byte("error!!"), false, true),
			err:  true,
		},
	}
	resultCh := make(chan *Result, len(cases))

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())

	var cur int
	subscriber := subee_testing.NewFakeSubscriber()
	consumer := subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
		r := &Result{
			tc:  cases[cur],
			msg: msg.(*subee_testing.FakeMessage),
		}

		resultCh <- r
		cur++
		if cur == len(cases) {
			close(resultCh)
			cancel()
		}
		if r.tc.err {
			return errors.New("error")
		}
		return nil
	})

	engine := subee.New(
		subscriber,
		consumer,
		subee.WithLogger(log.New(ioutil.Discard, "", 0)),
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := engine.Start(ctx)
		if err != nil {
			t.Errorf("Start returned an error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Millisecond)
		for _, m := range in {
			subscriber.AddMessage(subee_testing.NewFakeMessage(m, false, false))
			time.Sleep(6 * time.Millisecond)
		}
	}()

	for r := range resultCh {
		r := r
		t.Run(r.tc.test, func(t *testing.T) {
			got, want := r.msg, r.tc.want

			if got, want := got.Data(), want.Data(); !reflect.DeepEqual(got, want) {
				t.Errorf("Message.Data() is %q, want %q", got, want)
			}
			if got, want := got.Acked(), want.Acked(); got != want {
				t.Errorf("Messages.Acked() is %t, want %t", got, want)
			}
			if got, want := got.Nacked(), want.Nacked(); got != want {
				t.Errorf("Messages.Nacked() is %t, want %t", got, want)
			}
		})
	}
}

func TestEngineWithBatchConsumer(t *testing.T) {
	in := [][][]byte{
		{
			[]byte("foo"),
			[]byte("bar"),
			[]byte("baz"),
			[]byte("qux"),
			[]byte("quux"),
		},
		{
			[]byte("error!!"),
			[]byte("error!!!"),
		},
	}

	type TestCase struct {
		test string
		want []*subee_testing.FakeMessage
		err  bool
	}
	type Result struct {
		msgs []*subee_testing.FakeMessage
		tc   TestCase
	}

	cases := []TestCase{
		{
			test: "when max chunk size",
			want: []*subee_testing.FakeMessage{
				subee_testing.NewFakeMessage([]byte("foo"), true, false),
				subee_testing.NewFakeMessage([]byte("bar"), true, false),
				subee_testing.NewFakeMessage([]byte("baz"), true, false),
			},
		},
		{
			test: "when time exceeded",
			want: []*subee_testing.FakeMessage{
				subee_testing.NewFakeMessage([]byte("qux"), true, false),
				subee_testing.NewFakeMessage([]byte("quux"), true, false),
			},
		},
		{
			test: "nacked when errored",
			want: []*subee_testing.FakeMessage{
				subee_testing.NewFakeMessage([]byte("error!!"), false, true),
				subee_testing.NewFakeMessage([]byte("error!!!"), false, true),
			},
			err: true,
		},
	}
	resultCh := make(chan *Result, len(cases))

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())

	var cur int
	subscriber := subee_testing.NewFakeSubscriber()
	consumer := subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
		r := &Result{tc: cases[cur]}
		for _, m := range msgs {
			r.msgs = append(r.msgs, m.(*subee_testing.FakeMessage))
		}
		resultCh <- r
		cur++
		if cur == len(cases) {
			close(resultCh)
			cancel()
		}
		if r.tc.err {
			return errors.New("error")
		}
		return nil
	})

	engine := subee.NewBatch(
		subscriber,
		consumer,
		subee.WithChunkSize(3),
		subee.WithFlushInterval(4*time.Millisecond),
		subee.WithLogger(log.New(ioutil.Discard, "", 0)),
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := engine.Start(ctx)
		if err != nil {
			t.Errorf("Start returned an error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Millisecond)
		for _, msgs := range in {
			for _, m := range msgs {
				subscriber.AddMessage(subee_testing.NewFakeMessage(m, false, false))
			}
			time.Sleep(6 * time.Millisecond)
		}
	}()

	for r := range resultCh {
		r := r
		t.Run(r.tc.test, func(t *testing.T) {
			if got, want := len(r.msgs), len(r.tc.want); got != want {
				t.Errorf("consumed %d messages, want %d", got, want)
			} else {
				for i, got := range r.msgs {
					want := r.tc.want[i]
					if got, want := got.Data(), want.Data(); !reflect.DeepEqual(got, want) {
						t.Errorf("Messages[%d].Data() is %q, want %q", i, got, want)
					}
					if got, want := got.Acked(), want.Acked(); got != want {
						t.Errorf("Messages[%d].Acked() is %t, want %t", i, got, want)
					}
					if got, want := got.Nacked(), want.Nacked(); got != want {
						t.Errorf("Messages[%d].Nacked() is %t, want %t", i, got, want)
					}
				}
			}
		})
	}
}
