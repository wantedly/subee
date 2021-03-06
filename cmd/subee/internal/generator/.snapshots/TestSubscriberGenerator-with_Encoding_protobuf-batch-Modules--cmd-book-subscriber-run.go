package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/srvc/appctx"
	"github.com/wantedly/subee"
)

func run() error {
	ctx := appctx.Global() // application-scope context

	subscriber, err := createSubscriber(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	engine := subee.NewBatch(
		subscriber,
		consumer.NewBookBatchConsumerAdapter(
			consumer.NewBookBatchConsumer(),
		),
	)

	return engine.Start(ctx)
}

func createSubscriber(ctx context.Context) (subee.Subscriber, error) {
	// TODO: not yet implemented
	return nil, errors.New("createSubscriber() has not been implemented yet")
}

