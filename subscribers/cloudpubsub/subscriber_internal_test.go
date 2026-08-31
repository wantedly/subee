package cloudpubsub

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

func TestCreateSubscriber_AppliesReceiveSettings(t *testing.T) {
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

	want := pubsub.ReceiveSettings{MaxOutstandingMessages: 1}
	subscriber, err := CreateSubscriber(
		ctx,
		"test-proj",
		subscription.ID(),
		WithClientOptions(option.WithGRPCConn(conn)),
		WithReceiveSettings(want),
	)
	orDie(t, err)

	got := subscriber.(*subscriberImpl).subscription.ReceiveSettings
	if got != want {
		t.Errorf("ReceiveSettings was not applied to the subscription: got %+v, want %+v", got, want)
	}
}
