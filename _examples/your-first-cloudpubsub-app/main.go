package main

import (
	"context"
	"log"

	"github.com/pkg/errors"

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
	Topic *pubsub.Topic
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
func publishEvents(publisher *publisher) {

}

func main() {
	ctx := context.Background()

	publisher, err := createPublisher(ctx)
	if err != nil {
		panic(err)
	}

	// producing events in background.
	go publishEvents(publisher)

	// Subscriber is created with subscriptionID.
	subscriber, err := cloudpubsub.CreateSubscriber(ctx, projectID, subscriptionID)
	if err != nil {
		panic(err)
	}

	engine := subee.NewWithMultiMessagesConsumer(
		subscriber,
		func() subee.MultiMessagesConsumer {
			return subee.MultiMessagesConsumerFunc(
				func(ctx context.Context, msgs []subee.Message) error {
					return nil
				},
			)
		}(),
		subee.WithMultiMessagesConsumerInterceptors(
			subee_recovery.MultiMessagesConsumerInterceptor(
				func(ctx context.Context, p interface{}) {
					log.Println(p)
				},
			),
			subee_zap.MultiMessagesConsumerInterceptor(
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
