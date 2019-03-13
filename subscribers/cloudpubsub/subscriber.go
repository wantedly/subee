package cloudpubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"github.com/wantedly/subee"
	"google.golang.org/api/option"
)

type subscriberImpl struct {
	*Config
	subscription *pubsub.Subscription
}

// CreateSubscriber returns Subscriber implementation.
func CreateSubscriber(ctx context.Context, projectID, subscriptionID string, opts ...Option) (subee.Subscriber, error) {
	cfg := &Config{
		ReceiveSettings: pubsub.ReceiveSettings{Synchronous: true},
	}
	cfg.apply(opts)

	sub := &subscriberImpl{
		Config: cfg,
	}

	c, err := createPubSubClient(ctx, projectID, sub.ClientOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create Google Cloud Pub/Sub client")
	}

	pubsubSub, err := createPubSubSubscription(ctx, subscriptionID, c)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create Google Cloud Pub/Sub subscription")
	}
	sub.subscription = pubsubSub

	return sub, nil
}

func createPubSubClient(ctx context.Context, projectID string, ops ...option.ClientOption) (*pubsub.Client, error) {
	if len(projectID) == 0 {
		return nil, errors.New("missing google project id")
	}

	c, err := pubsub.NewClient(ctx, projectID, ops...)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create pub/sub client")
	}

	return c, nil
}

func createPubSubSubscription(ctx context.Context, subscriptionID string, c *pubsub.Client) (*pubsub.Subscription, error) {
	pubsubSub := c.Subscription(subscriptionID)

	ok, err := pubsubSub.Exists(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "faild to reports whether the subscription exists on the server")
	}

	if !ok {
		return nil, errors.New("failed to get pub/sub subscription. Check subscription presence")
	}

	return pubsubSub, nil
}

func (r *subscriberImpl) Subscribe(ctx context.Context, f func(subee.Message)) error {
	err := r.subscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		f(&Message{m})
	})

	return errors.WithStack(err)
}
