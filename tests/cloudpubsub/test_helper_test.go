package cloudpubsub_test

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"cloud.google.com/go/pubsub"
)

type publisher struct {
	Topic *pubsub.Topic
}

func createPublisher(ctx context.Context, projectID, topicID, subscriptionID string) (*publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create google cloud pub/sub client")
	}

	topic := client.Topic(topicID)
	if ok, _ := topic.Exists(ctx); !ok {
		client.CreateTopic(ctx, topicID)
	}

	subscription := client.Subscription(subscriptionID)
	if ok, _ := subscription.Exists(ctx); ok {
		subscription.Delete(ctx)
	}

	client.CreateSubscription(
		ctx,
		subscriptionID,
		pubsub.SubscriptionConfig{
			Topic: topic,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "faild to create pub/sub subscription")
	}

	return &publisher{
		Topic: topic,
	}, nil
}

func (p *publisher) publish(ctx context.Context, in [][]byte) {
	time.Sleep(3 * time.Millisecond)
	for _, m := range in {
		p.Topic.Publish(ctx, &pubsub.Message{
			Data: m,
		})
		time.Sleep(6 * time.Millisecond)
	}
}

func createInputData(n int) [][]byte {
	in := make([][]byte, n)
	for i := 0; i < len(in); i++ {
		in[i] = []byte{byte(i)}
	}
	return in
}
