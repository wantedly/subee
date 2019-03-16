package subee

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type process interface {
	Start(ctx context.Context) error
}

type processImpl struct {
	*Engine
	wg sync.WaitGroup
}

func newProcess(e *Engine) process {
	return &processImpl{
		Engine: e,
	}
}

func (p *processImpl) Start(ctx context.Context) error {
	p.Logger.Print("Start process")
	defer p.Logger.Print("Finish process")
	defer p.wg.Wait()

	switch {
	case p.Consumer != nil:
		return errors.WithStack(p.startConsumingProcess(ctx))
	case p.BatchConsumer != nil:
		return errors.WithStack(p.startBatchConsumingProcess(ctx))
	default:
		panic("unreachable")
	}
}

func (p *processImpl) startConsumingProcess(ctx context.Context) error {
	consumer := chainConsumerInterceptors(p.Consumer, p.ConsumerInterceptors...)

	err := p.subscribe(ctx, func(in Message) {
		p.handleMessage(p.createConsumingContext(), &singleMessage{Message: in}, func(ctx context.Context) error {
			return errors.WithStack(consumer.Consume(ctx, in))
		})
	})
	return errors.WithStack(err)
}

func (p *processImpl) startBatchConsumingProcess(ctx context.Context) error {
	inCh, outCh := createBufferedQueue(
		p.createConsumingContext,
		p.ChunkSize,
		p.FlushInterval,
	)

	defer close(inCh)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		batchConsumer := chainBatchConsumerInterceptors(p.BatchConsumer, p.BatchConsumerInterceptors...)

		p.Logger.Print("Start batch consuming process")
		defer p.Logger.Print("Finish batch consuming process")

		for m := range outCh {
			msgs := m.Msgs
			p.handleMessage(m.Ctx, m, func(ctx context.Context) error {
				return errors.WithStack(batchConsumer.BatchConsume(ctx, msgs))
			})
		}
	}()

	return errors.WithStack(p.subscribe(ctx, func(msg Message) { inCh <- msg }))
}

func (p *processImpl) subscribe(ctx context.Context, f func(msg Message)) error {
	p.Logger.Print("Start subscribing messages")
	defer p.Logger.Print("Finish subscribing messages")
	return errors.WithStack(p.subscriber.Subscribe(ctx, f))
}

func (p *processImpl) createConsumingContext() context.Context {
	ctx := context.Background()
	ctx = p.StatsHandler.TagProcess(ctx, &BeginTag{})
	ctx = p.StatsHandler.TagProcess(ctx, &EnqueueTag{})
	ctx = setEnqueuedAt(ctx, time.Now().UTC())
	return ctx
}

func (p *processImpl) handleMessage(ctx context.Context, m queuedMessage, handle func(context.Context) error) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.Logger.Printf("Start consuming %d messages", m.Count())
		defer p.Logger.Printf("Finish consuming %d messages", m.Count())

		if p.AckImmediately {
			m.Ack()
		}

		enqueuedAt := getEnqueuedAt(ctx)

		p.StatsHandler.HandleProcess(ctx, &Dequeue{
			BeginTime: enqueuedAt,
			EndTime:   time.Now(),
		})

		beginTime := time.Now()

		ctx = p.StatsHandler.TagProcess(ctx, &ConsumeBeginTag{})

		err := handle(ctx)

		if !p.AckImmediately {
			if err != nil {
				m.Nack()
			} else {
				m.Ack()
			}
		}

		p.StatsHandler.HandleProcess(ctx, &ConsumeEnd{
			BeginTime: beginTime,
			EndTime:   time.Now(),
			Error:     err,
		})

		p.StatsHandler.HandleProcess(ctx, &End{
			MsgCount:  m.Count(),
			BeginTime: enqueuedAt,
			EndTime:   time.Now(),
		})
	}()
}
