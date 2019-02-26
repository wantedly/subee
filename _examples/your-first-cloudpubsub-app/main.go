package main

import (
	"context"
	"log"

	"github.com/pkg/errors"

	"time"

	"cloud.google.com/go/pubsub"
	"github.com/wantedly/subee"
	"github.com/wantedly/subee/middlewares/logging/zap"
	"github.com/wantedly/subee/middlewares/recovery"
	"github.com/wantedly/subee/subscribers/cloudpubsub"
	"go.uber.org/zap"
)

const (
	// ProjectID is Google Cloud project id for example.
	projectID = "your-first-cloudpub-app"

	// topicID is Pub/Sub topic id for example.
	topicID = "pub-sub-test"

	subscriptionID = "sub-1"
)

type publisher struct {
	*pubsub.Topic
}

// createPublisher is a helper function which creates Publisher, in this case - Publisher of Google Cloud Pub/Sub.
func createPublisher(ctx context.Context) (*publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create pub/sub client")
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

	return &publisher{
		Topic: topic,
	}, nil
}

// publishEvents which will produce some events with async for consuming.
func publishEvents(ctx context.Context, publisher *publisher) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			publisher.Topic.Publish(ctx, &pubsub.Message{})
		}
	}
}

func main() {
	ctx := context.Background()

	publisher, err := createPublisher(ctx)
	if err != nil {
		panic(err)
	}

	// producing events in background.
	go publishEvents(ctx, publisher)

	engine := subee.NewWithSingleMessageConsumer(
		// Subscriber is created with subscriptionID.
		func() subee.Subscriber {
			s, err := cloudpubsub.CreateSubscriber(ctx, projectID, subscriptionID)
			if err != nil {
				panic(err)
			}
			return s
		}(),
		func() subee.SingleMessageConsumer {
			return subee.SingleMessageConsumerFunc(
				func(ctx context.Context, msg subee.Message) error {
					return nil
				},
			)
		}(),
		subee.WithSingleMessageConsumerInterceptors(
			subee_recovery.SingleMessageConsumerInterceptor(
				func(ctx context.Context, p interface{}) {
					log.Println(p)
				},
			),
			subee_zap.SingleMessageConsumerInterceptor(
				func() *zap.Logger {
					logger, err := zap.NewProduction()
					if err != nil {
						panic(err)
					}
					return logger
				}(),
			),
		),
		subee.WithLogger(nil),
		subee.WithAckImmediately(),
	)

	err = engine.Start(ctx)
	if err != nil {
		panic(err)
	}
}
