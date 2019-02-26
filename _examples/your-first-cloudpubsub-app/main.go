package main

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/wantedly/subee"
	"github.com/wantedly/subee/subscribers/cloudpubsub"
)

func main() {
	ctx := context.Background()

	publisher, err := createPublisher(ctx)
	if err != nil {
		panic(err)
	}

	// producing events in background.
	go publisher.publishEvents(ctx)

	// creates subscriber, in this case - Subscriber of Google Cloud Pub/Sub.
	subscriber, err := cloudpubsub.CreateSubscriber(ctx, "your-first-cloudpub-app", "subscription-id")
	if err != nil {
		panic(err)
	}

	logger, _ := zap.NewProduction()

	engine := subee.NewWithSingleMessageConsumer(
		// Set Subscriber implementation.
		subscriber,
		// Set SingleMessageConsumer implementation.
		// If engine is created with a constructor for single message consumer type, you have to add SingleMessageConsumer implementation.
		subee.SingleMessageConsumerFunc(
			func(ctx context.Context, msg subee.Message) error {
				payload := event{}

				json.Unmarshal(msg.Data(), &payload)

				logger.Info("Received event",
					zap.Int64("created_at", payload.CreatedAt),
				)

				return nil
			},
		),
	)

	err = engine.Start(ctx)
	if err != nil {
		panic(err)
	}
}

// publisher represents publisher structor of Google Cloud Pub/Sub.
type publisher struct {
	*pubsub.Topic
}

// event represents message structor to be published.
type event struct {
	CreatedAt int64
}

// createPublisher is a helper function which creates Publisher, in this case - Publisher of Google Cloud Pub/Sub.
func createPublisher(ctx context.Context) (*publisher, error) {
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

	return &publisher{
		Topic: topic,
	}, nil
}

// publishEvents which will produce some events for consuming.
func (p *publisher) publishEvents(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			payload, _ := json.Marshal(
				event{
					CreatedAt: time.Now().UnixNano(),
				},
			)
			p.Topic.Publish(ctx, &pubsub.Message{
				Data: payload,
			})
		}
	}
}
