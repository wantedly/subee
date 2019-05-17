package util

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"

	"your-first-cloudpubsub-app/pkg/model"
)

// Publisher represents publisher structor of Google Cloud Pub/Sub.
type Publisher struct {
	topic  *pubsub.Topic
	client *pubsub.Client
}

// CreatePublisher is a helper function which creates Publisher, in this case - Publisher of Google Cloud Pub/Sub.
func CreatePublisher(ctx context.Context) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, "your-first-cloudpub-app")
	if err != nil {
		return nil, errors.Wrap(err, "faild to create pub/sub client")
	}

	topic := client.Topic("pub-sub-topic")
	if ok, _ := topic.Exists(ctx); !ok {
		client.CreateTopic(ctx, "pub-sub-topic")
	}

	subscription := client.Subscription("subscription-id")
	if ok, _ := subscription.Exists(ctx); ok {
		// Since example is for testing, subscription is deleted if it exists, but it is not necessary to delete it in the production environment.
		subscription.Delete(ctx)
	}

	client.CreateSubscription(
		ctx,
		"subscription-id",
		pubsub.SubscriptionConfig{
			Topic: topic,
		},
	)

	return &Publisher{
		topic:  topic,
		client: client,
	}, nil
}

// PublishEvents which will produce some events for consuming.
func (p *Publisher) PublishEvents(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			payload, _ := json.Marshal(
				model.Event{
					CreatedAt: time.Now().UnixNano(),
				},
			)
			p.topic.Publish(ctx, &pubsub.Message{
				Data: payload,
			})
		}
	}
}

func (p *Publisher) Close() {
	p.topic.Stop()
	p.client.Close()
}
