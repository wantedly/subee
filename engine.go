package subee

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
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

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	ctx = setLogger(ctx, e.Logger)

	sigDoneCh := make(chan struct{}, 1)

	eg.Go(func() error {
		return e.watchShutdownSignal(sigDoneCh, cancel)
	})

	eg.Go(func() error {
		defer close(sigDoneCh)
		return errors.WithStack(newProcess(e).Start(ctx))
	})

	err := eg.Wait()

	return errors.WithStack(err)
}

func (e *Engine) watchShutdownSignal(sigstopCh <-chan struct{}, cancel context.CancelFunc) error {
	sigCh := make(chan os.Signal, 1)

	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigstopCh:
			e.Logger.Print("Finish os signal monitoring")
			return nil
		case sig := <-sigCh:
			e.Logger.Printf("Received of signal: %v", sig)
			cancel()
		}
	}
}
