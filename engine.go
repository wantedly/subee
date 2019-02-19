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

	var (
		inCh  chan<- Message
		outCh <-chan queuedMessage
	)

	if e.SingleMessageConsumer != nil {
		inCh, outCh = createQueue(e.createStatsHandlerCtx)
	}

	if e.MultiMessagesConsumer != nil {
		inCh, outCh = createBufferedQueue(
			e.createStatsHandlerCtx,
			e.ChunkSize,
			e.FlushInterval,
		)
	}

	eg.Go(func() error {
		defer close(inCh)
		err := e.subscriber.Subscribe(ctx, func(msg Message) { inCh <- msg })
		return errors.WithStack(err)
	})

	eg.Go(func() error {
		defer close(sigDoneCh)
		return e.handle(outCh)
	})

	err := eg.Wait()

	return errors.WithStack(err)
}

func (e *Engine) createStatsHandlerCtx() context.Context {
	ctx := context.Background()
	ctx = e.StatsHandler.TagProcess(ctx, &BeginTag{})
	ctx = e.StatsHandler.TagProcess(ctx, &EnqueueTag{})
	return ctx
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

func (e *Engine) consume(qm queuedMessage) error {
	switch m := qm.(type) {
	case *singleMessage:
		// TODO (@hlts2): consume single message.
	case *multiMessages:
		return e.MultiMessagesConsumer.Consume(m.Ctx, m.Msgs)
	}

	// TODO(@hlts2): return error.
	return nil
}

func (e *Engine) handle(msgCh <-chan queuedMessage) error {
	e.Logger.Print("Start consume process")
	defer e.Logger.Print("Finish consume process")

	for m := range msgCh {
		if e.AckImmediately {
			m.Ack()
		}

		e.StatsHandler.HandleProcess(m.Context(), &Dequeue{
			BeginTime: m.GetEnqueuedAt(),
			EndTime:   time.Now(),
		})

		beginTime := time.Now()

		m.SetContext(e.StatsHandler.TagProcess(m.Context(), &ConsumeBeginTag{}))

		// discard the error, because error can be handled with interceptor.
		err := e.consume(m)
		if !e.AckImmediately {
			if err != nil {
				m.Nack()
			} else {
				m.Ack()
			}
		}

		e.StatsHandler.HandleProcess(m.Context(), &ConsumeEnd{
			BeginTime: beginTime,
			EndTime:   time.Now(),
			Error:     err,
		})

		e.StatsHandler.HandleProcess(m.Context(), &End{
			MsgCount:  m.Count(),
			BeginTime: m.GetEnqueuedAt(),
			EndTime:   time.Now(),
		})
	}

	return nil
}
