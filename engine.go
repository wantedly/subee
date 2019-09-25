package subee

import (
	"context"

	"github.com/pkg/errors"
)

// Engine is the framework instance.
type Engine struct {
	*Config
	subscriber Subscriber
}

// New creates a Engine intstance with Consumer.
func New(subscriber Subscriber, consumer Consumer, opts ...Option) *Engine {
	return newEngine(subscriber, nil, consumer, opts...)
}

// NewBatch creates a Engine intstance with BatchConsumer.
func NewBatch(subscriber Subscriber, consumer BatchConsumer, opts ...Option) *Engine {
	return newEngine(subscriber, consumer, nil, opts...)
}

func newEngine(subscriber Subscriber, bConsumer BatchConsumer, consumer Consumer, opts ...Option) *Engine {
	cfg := newDefaultConfig()
	cfg.BatchConsumer = bConsumer
	cfg.Consumer = consumer
	cfg.apply(opts)

	e := &Engine{
		Config:     cfg,
		subscriber: subscriber,
	}

	return e
}

// Start starts Subscriber and Consumer process.
func (e *Engine) Start(ctx context.Context) error {
	e.Logger.Print("Start Pub/Sub worker")
	defer e.Logger.Print("Finish Pub/Sub worker")

	ctx = setLogger(ctx, e.Logger)

	return errors.WithStack(newProcess(e).Start(ctx))
}
