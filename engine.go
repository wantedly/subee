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

	var (
		inCh  chan<- Message
		outCh <-chan queuedMessage
	)

	if e.Consumer != nil {
		inCh, outCh = createQueue(
			queuingContext(e.StatsHandler),
		)
	}

	if e.BatchConsumer != nil {
		inCh, outCh = createBufferedQueue(
			queuingContext(e.StatsHandler),
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
		return e.handleConcurrently(outCh)
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

func (e *Engine) handleConcurrently(msgCh <-chan queuedMessage) error {
	eg := errgroup.Group{}

	for i := 0; i < e.Config.Concurrency; i++ {
		eg.Go(func() error {
			return e.handle(msgCh, i)
		})
	}

	err := eg.Wait()

	return err
}

func (e *Engine) handle(msgCh <-chan queuedMessage, id int) error {
	var (
		consumer      Consumer
		batchConsumer BatchConsumer
	)
	switch {
	case e.Consumer != nil:
		consumer = chainConsumerInterceptors(e.Consumer, e.ConsumerInterceptors...)
	case e.BatchConsumer != nil:
		batchConsumer = chainBatchConsumerInterceptors(e.BatchConsumer, e.BatchConsumerInterceptors...)
	default:
		panic("unreachable")
	}

	e.Logger.Printf("Start consumer %d", id)
	defer e.Logger.Printf("Finish consumer %d", id)

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

		var err error
		switch m := m.(type) {
		case *singleMessage:
			err = errors.WithStack(consumer.Consume(m.Ctx, m.Msg))
		case *multiMessages:
			err = errors.WithStack(batchConsumer.BatchConsume(m.Ctx, m.Msgs))
		default:
			panic("unreachable")
		}

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
