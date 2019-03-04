package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/wantedly/subee"
)

func run() error {
	ctx := context.Background()

	subscriber, err := createSubscriber(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	engine := subee.NewWithSingleMessageConsumer(
		subscriber,
		consumer.NewBookConsumerAdapter(
			consumer.NewBookConsumer(),
		),
	)

	return engine.Start(ctx)
}

func createSubscriber(ctx context.Context) (subee.Subscriber, error) {
	// TODO: not yet implemented
	return nil, errors.New("createSubscriber() has not been implemented yet")
}

