package cloudpubsub_test

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/wantedly/subee"
	"github.com/wantedly/subee/subscribers/cloudpubsub"
)

func TestSubscriber(t *testing.T) {
	orDie := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	ctx := context.Background()

	srv := pstest.NewServer()
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	orDie(t, err)
	defer conn.Close()
	client, err := pubsub.NewClient(ctx, "test-proj", option.WithGRPCConn(conn))
	orDie(t, err)
	defer client.Close()
	topic, err := client.CreateTopic(ctx, "test-topic")
	orDie(t, err)
	subscription, err := client.CreateSubscription(ctx, "test-sub", pubsub.SubscriptionConfig{Topic: topic})
	orDie(t, err)

	subscriber, err := cloudpubsub.CreateSubscriber(
		ctx,
		"test-proj",
		subscription.ID(),
		cloudpubsub.WithClientOptions(option.WithGRPCConn(conn)),
	)
	orDie(t, err)

	in := [][]byte{
		[]byte("foo"),
		[]byte("qux"),
		[]byte("baz"),
		[]byte("bar"),
		[]byte("quux"),
	}

	msgCh := make(chan subee.Message)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, data := range in {
			srv.Publish("projects/test-proj/topics/test-topic", data, nil)
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

	out := [][]byte{}
	for m := range msgCh {
		out = append(out, m.Data())
	}

	sorter := cmp.Transformer("Sort", func(in [][]byte) [][]byte {
		out := append([][]byte(nil), in...)
		sort.Slice(out, func(i, j int) bool { return string(out[i]) < string(out[j]) })
		return out
	})
	if diff := cmp.Diff(in, out, sorter); diff != "" {
		t.Errorf("Received message differs: (-want +got)\n%s", diff)
	}

	cancel()
	wg.Wait()
}
