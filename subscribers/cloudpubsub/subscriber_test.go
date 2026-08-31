package cloudpubsub_test

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/wantedly/subee"
	"github.com/wantedly/subee/subscribers/cloudpubsub"
)

func orDie(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

const testTopicName = "projects/test-proj/topics/test-topic"

type testEnv struct {
	srv          *pstest.Server
	conn         *grpc.ClientConn
	subscription *pubsub.Subscription
}

func newTestEnv(t *testing.T, ctx context.Context) *testEnv {
	t.Helper()

	srv := pstest.NewServer()
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	orDie(t, err)
	t.Cleanup(func() { conn.Close() })

	client, err := pubsub.NewClient(ctx, "test-proj", option.WithGRPCConn(conn))
	orDie(t, err)
	t.Cleanup(func() { client.Close() })

	topic, err := client.CreateTopic(ctx, "test-topic")
	orDie(t, err)
	subscription, err := client.CreateSubscription(ctx, "test-sub", pubsub.SubscriptionConfig{Topic: topic})
	orDie(t, err)

	return &testEnv{srv: srv, conn: conn, subscription: subscription}
}

func TestSubscriber(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t, ctx)

	subscriber, err := cloudpubsub.CreateSubscriber(
		ctx,
		"test-proj",
		env.subscription.ID(),
		cloudpubsub.WithClientOptions(option.WithGRPCConn(env.conn)),
	)
	orDie(t, err)

	type Msg struct {
		Data []byte
		Meta map[string]string
	}

	in := []Msg{
		{Data: []byte("foo")},
		{Data: []byte("qux")},
		{Data: []byte("baz"), Meta: map[string]string{"corge": "12", "id": "aaabbbccc"}},
		{Data: []byte("bar")},
		{Data: []byte("quux")},
	}

	msgCh := make(chan subee.Message)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, m := range in {
			env.srv.Publish(testTopicName, m.Data, m.Meta)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var cnt int32
		err := subscriber.Subscribe(ctx, func(msg subee.Message) {
			msgCh <- msg
			if int(atomic.AddInt32(&cnt, 1)) >= len(in) {
				close(msgCh)
			}
			msg.Ack()
		})
		if err != nil {
			t.Errorf("Subscribe returned an error: %v", err)
		}
	}()

	out := []Msg{}
	for m := range msgCh {
		out = append(out, Msg{Data: m.Data(), Meta: m.Metadata()})
	}

	sorter := cmp.Transformer("Sort", func(in []Msg) []Msg {
		out := append([]Msg(nil), in...)
		sort.Slice(out, func(i, j int) bool { return string(out[i].Data) < string(out[j].Data) })
		return out
	})
	if diff := cmp.Diff(in, out, sorter); diff != "" {
		t.Errorf("Received message differs: (-want +got)\n%s", diff)
	}

	cancel()
	wg.Wait()
}

// MaxOutstandingMessages bounds concurrent deliveries via the client's flow
// controller regardless of NumGoroutines, making this deterministic to observe.
func TestSubscriber_AppliesReceiveSettings(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t, ctx)

	subscriber, err := cloudpubsub.CreateSubscriber(
		ctx,
		"test-proj",
		env.subscription.ID(),
		cloudpubsub.WithClientOptions(option.WithGRPCConn(env.conn)),
		cloudpubsub.WithReceiveSettings(pubsub.ReceiveSettings{MaxOutstandingMessages: 1}),
	)
	orDie(t, err)

	const numMsgs = 10
	for i := 0; i < numMsgs; i++ {
		env.srv.Publish(testTopicName, []byte("x"), nil)
	}

	var inFlight, maxInFlight, processed int32
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- subscriber.Subscribe(ctx, func(msg subee.Message) {
			cur := atomic.AddInt32(&inFlight, 1)
			for {
				m := atomic.LoadInt32(&maxInFlight)
				if cur <= m || atomic.CompareAndSwapInt32(&maxInFlight, m, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&inFlight, -1)
			msg.Ack()
			if atomic.AddInt32(&processed, 1) >= numMsgs {
				cancel()
			}
		})
	}()
	orDie(t, <-done)

	if got := atomic.LoadInt32(&maxInFlight); got > 1 {
		t.Errorf("MaxOutstandingMessages was not applied to the subscription: observed %d concurrent in-flight messages, want at most 1", got)
	}
}
