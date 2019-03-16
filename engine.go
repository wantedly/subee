package subee

import (
	"context"
	"os"
	"os/signal"
	"sync"
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

	eg.Go(func() error {
		defer close(sigDoneCh)
		switch {
		case e.Consumer != nil:
			return errors.WithStack(e.startConsumingProcess(ctx))
		case e.BatchConsumer != nil:
			return errors.WithStack(e.startBatchConsumingProcess(ctx))
		default:
			panic("unreachable")
		}
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

func (e *Engine) startConsumingProcess(ctx context.Context) error {
	consumer := chainConsumerInterceptors(e.Consumer, e.ConsumerInterceptors...)

	err := e.startSubscribing(ctx, func(in Message) {
		msg := &singleMessage{
			Ctx:        e.createConsumingContext(),
			Msg:        in,
			EnqueuedAt: time.Now(),
		}
		e.handleMessage(msg, func(in queuedMessage) error {
			m := in.(*singleMessage)
			return errors.WithStack(consumer.Consume(m.Ctx, m.Msg))
		})
	})
	return errors.WithStack(err)
}

func (e *Engine) startBatchConsumingProcess(ctx context.Context) error {
	var (
		err error
		wg  sync.WaitGroup
	)

	inCh, outCh := createBufferedQueue(
		e.createConsumingContext,
		e.ChunkSize,
		e.FlushInterval,
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(inCh)

		err = errors.WithStack(e.startSubscribing(ctx, func(msg Message) { inCh <- msg }))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		batchConsumer := chainBatchConsumerInterceptors(e.BatchConsumer, e.BatchConsumerInterceptors...)

		e.Logger.Print("Start consuming process")
		defer e.Logger.Print("Finish consuming process")

		for m := range outCh {
			e.handleMessage(m, func(in queuedMessage) error {
				m := in.(*multiMessages)
				return errors.WithStack(batchConsumer.BatchConsume(m.Ctx, m.Msgs))
			})
		}
	}()

	wg.Wait()

	return err
}

func (e *Engine) startSubscribing(ctx context.Context, f func(msg Message)) error {
	e.Logger.Print("Start subscribing process")
	defer e.Logger.Print("Finish subscribing process")
	return errors.WithStack(e.subscriber.Subscribe(ctx, f))
}

func (e *Engine) createConsumingContext() context.Context {
	ctx := context.Background()
	ctx = e.StatsHandler.TagProcess(ctx, &BeginTag{})
	ctx = e.StatsHandler.TagProcess(ctx, &EnqueueTag{})
	return ctx
}

func (e *Engine) handleMessage(m queuedMessage, handle func(queuedMessage) error) {
	if e.AckImmediately {
		m.Ack()
	}

	e.StatsHandler.HandleProcess(m.Context(), &Dequeue{
		BeginTime: m.GetEnqueuedAt(),
		EndTime:   time.Now(),
	})

	beginTime := time.Now()

	m.SetContext(e.StatsHandler.TagProcess(m.Context(), &ConsumeBeginTag{}))

	err := handle(m)

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
