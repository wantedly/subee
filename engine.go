package subee

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Engine is the framework instance.
type Engine struct {
	*Config
	subscriber Subscriber
}

// New creates a Engine intstance.
func New(subscriber Subscriber, consumer MultiMessagesConsumer, opts ...Option) *Engine {
	cfg := newDefaultConfig()
	cfg.MultiMessagesConsumer = consumer
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

	inCh, outCh := createBufferedQueue(
		func() context.Context {
			ctx := context.Background()
			ctx = e.StatsHandler.TagProcess(ctx, &BeginTag{})
			ctx = e.StatsHandler.TagProcess(ctx, &EnqueueTag{})
			return ctx
		},
		e.ChunkSize,
		e.FlushInterval,
	)

	eg.Go(func() error {
		defer close(inCh)
		err := e.subscriber.Subscribe(ctx, func(msg Message) { inCh <- msg })
		return errors.WithStack(err)
	})

	eg.Go(func() error {
		defer close(sigDoneCh)
		return e.consume(outCh)
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

func (e *Engine) consume(msgCh <-chan *multiMessages) error {
	e.Logger.Print("Start consume process")
	defer e.Logger.Print("Finish consume process")

	for m := range msgCh {
		if e.AckImmediately {
			m.Ack()
		}

		e.StatsHandler.HandleProcess(m.Ctx, &Dequeue{
			BeginTime: m.EnqueuedAt,
			EndTime:   time.Now(),
		})

		beginTime := time.Now()

		m.Ctx = e.StatsHandler.TagProcess(m.Ctx, &ConsumeBeginTag{})

		// // discard the error, because error can be handled with interceptor.
		err := e.MultiMessagesConsumer.Consume(m.Ctx, m.Msgs)
		if !e.AckImmediately {
			if err != nil {
				m.Nack()
			} else {
				m.Ack()
			}
		}

		e.StatsHandler.HandleProcess(m.Ctx, &ConsumeEnd{
			BeginTime: beginTime,
			EndTime:   time.Now(),
			Error:     err,
		})

		e.StatsHandler.HandleProcess(m.Ctx, &End{
			MsgCount:  len(m.Msgs),
			BeginTime: m.EnqueuedAt,
			EndTime:   time.Now(),
		})
	}

	return nil
}
