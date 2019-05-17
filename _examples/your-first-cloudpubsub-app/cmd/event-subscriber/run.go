package main

import (
	"context"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/wantedly/subee"
	"github.com/wantedly/subee/subscribers/cloudpubsub"

	"your-first-cloudpubsub-app/pkg/consumer"
	"your-first-cloudpubsub-app/pkg/util"
)

func run() error {
	ctx := context.Background()

	publisher, err := util.CreatePublisher(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer publisher.Close()

	// producing events in background.
	go publisher.PublishEvents(ctx)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// creates subscriber, in this case - Subscriber of Google Cloud Pub/Sub.
	subscriber, err := cloudpubsub.CreateSubscriber(ctx, "your-first-cloudpub-app", "subscription-id")
	if err != nil {
		return errors.WithStack(err)
	}

	engine := subee.New(
		subscriber,
		consumer.NewEventConsumerAdapter(
			consumer.NewEventConsumer(logger),
		),
		subee.WithLogger(logger),
	)

	return engine.Start(ctx)
}
