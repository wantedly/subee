package main

import (
	"context"
	"encoding/json"
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

	// SubscriptionID is Pub/Sub subscription id for example
	subscriptionID = "sub-1"
)

type publisher struct {
	*pubsub.Topic
}

// event represents message structor to be published.
type event struct {
	CreatedAt int64
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

func main() {
	ctx := context.Background()

	publisher, err := createPublisher(ctx)
	if err != nil {
		panic(err)
	}

	// producing events in background.
	go publisher.publishEvents(ctx)

	logger, _ := zap.NewProduction()

	engine := subee.NewWithSingleMessageConsumer(
		// Set Subscriber implementation, in this case - Subscriber of Google Cloud Pub/Sub.
		func() subee.Subscriber {
			// Create a subscriber of Google Cloud Pub/Sub with Google Cloud project id and subscribtion id.
			s, err := cloudpubsub.CreateSubscriber(ctx, projectID, subscriptionID)
			if err != nil {
				panic(err)
			}
			return s
		}(),
		// Set SingleMessageConsumer implementation.
		// If engine is created with a constructor for single message consumer type, you have to add SingleMessageConsumer implementation.
		func() subee.SingleMessageConsumer {
			return subee.SingleMessageConsumerFunc(
				func(ctx context.Context, msg subee.Message) error {
					payload := event{}

					err := json.Unmarshal(msg.Data(), &payload)
					if err != nil {
						return errors.Wrap(err, "faild to unmarshal message")
					}

					logger.Info("received event",
						zap.Int64("created_at", payload.CreatedAt),
					)

					return nil
				},
			)
		}(),
		// Set interceptor(s) option.
		subee.WithSingleMessageConsumerInterceptors(
			// Receive instant notification of panics in your Go applications.
			subee_recovery.SingleMessageConsumerInterceptor(
				func(ctx context.Context, p interface{}) {
					log.Printf("panic recovery: %v", p)
				},
			),
			// Log the consume handler with zap logging libray.
			subee_zap.SingleMessageConsumerInterceptor(logger),
		),
		// Set logger option.
		subee.WithLogger(zap.NewStdLog(logger)),
		// Set immediate ack option.
		subee.WithAckImmediately(),
	)

	err = engine.Start(ctx)
	if err != nil {
		panic(err)
	}
}
