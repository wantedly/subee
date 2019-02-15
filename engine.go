package subee

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Engine is the framework instance.
type Engine struct {
	msgCh      chan *BufferedMessage
	subscriber Subscriber

	logger Logger
	sh     StatsHandler

	mmConsumer MultiMessagesConsumer
	scfg       *SubscribeConfig
}

// New creates a Engine intstance.
func New(subscriber Subscriber, consumer MultiMessagesConsumer, options ...Option) *Engine {
	e := &Engine{
		subscriber: subscriber,
		mmConsumer: consumer,
		msgCh:      make(chan *BufferedMessage, 1),
		logger:     log.New(os.Stdout, "", log.LstdFlags),
		sh:         new(NopStatsHandler),
		scfg:       newDefaultSubscribeConfig(),
	}

	for _, option := range options {
		option(e)
	}

	return e
}

// Start starts Subscriber and Consumer process.
func (e *Engine) Start(ctx context.Context) error {
	e.logger.Print("Start Pub/Sub worker")
	defer e.logger.Print("Finish Pub/Sub worker")

	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)

	sigDoneCh := make(chan struct{}, 1)

	eg.Go(func() error {
		return e.watchShutdownSignal(sigDoneCh, cancel)
	})

	eg.Go(func() error {
		ctx = e.setSubscriberSettings(ctx)
		return e.subscriber.Subscribe(ctx, e.msgCh)
	})

	eg.Go(func() error {
		defer close(sigDoneCh)
		return e.consume(e.msgCh)
	})

	err := eg.Wait()

	return errors.WithStack(err)
}

// setSubscriberSettings sets common settings for Subscriber using context object.
func (e *Engine) setSubscriberSettings(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, loggerContextKey{}, e.logger)
	ctx = context.WithValue(ctx, statsHandlerContextKey{}, e.sh)
	return context.WithValue(ctx, subscribeConfigContextkey{}, e.scfg)
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
			e.logger.Print("Finish os signal monitoring")
			return nil
		case sig := <-sigCh:
			e.logger.Printf("Received of signal: %v", sig)
			cancel()
		}
	}
}

func (e *Engine) consume(msgCh <-chan *BufferedMessage) error {
	e.logger.Print("Start consume process")
	defer e.logger.Print("Finish consume process")

	for m := range msgCh {
		e.sh.HandleProcess(m.Ctx, &Dequeue{
			BeginTime: getTimeByContextKey(m.Ctx, EnqueueTimeContextKey{}),
			EndTime:   time.Now(),
		})

		beginTime := time.Now()

		m.Ctx = e.sh.TagProcess(m.Ctx, &ConsumeBeginTag{})

		// // discard the error, because error can be handled with interceptor.
		err := e.mmConsumer.Consume(m.Ctx, m.Msgs)

		e.sh.HandleProcess(m.Ctx, &ConsumeEnd{
			BeginTime: beginTime,
			EndTime:   time.Now(),
			Error:     err,
		})

		e.sh.HandleProcess(m.Ctx, &End{
			MsgCount:  len(m.Msgs),
			BeginTime: getTimeByContextKey(m.Ctx, BeginTimeContextKey{}),
			EndTime:   time.Now(),
		})
	}

	return nil
}
