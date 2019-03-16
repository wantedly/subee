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
		msg := &singleMessage{
			Ctx:        p.createConsumingContext(),
			Msg:        in,
			EnqueuedAt: time.Now(),
		}
		p.handleMessage(msg, func(in queuedMessage) error {
			m := in.(*singleMessage)
			return errors.WithStack(consumer.Consume(m.Ctx, m.Msg))
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
			p.handleMessage(m, func(in queuedMessage) error {
				m := in.(*multiMessages)
				return errors.WithStack(batchConsumer.BatchConsume(m.Ctx, m.Msgs))
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
	return ctx
}

func (p *processImpl) handleMessage(m queuedMessage, handle func(queuedMessage) error) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.Logger.Printf("Start consuming %d messages", m.Count())
		defer p.Logger.Printf("Finish consuming %d messages", m.Count())

		if p.AckImmediately {
			m.Ack()
		}

		p.StatsHandler.HandleProcess(m.Context(), &Dequeue{
			BeginTime: m.GetEnqueuedAt(),
			EndTime:   time.Now(),
		})

		beginTime := time.Now()

		m.SetContext(p.StatsHandler.TagProcess(m.Context(), &ConsumeBeginTag{}))

		err := handle(m)

		if !p.AckImmediately {
			if err != nil {
				m.Nack()
			} else {
				m.Ack()
			}
		}

		p.StatsHandler.HandleProcess(m.Context(), &ConsumeEnd{
			BeginTime: beginTime,
			EndTime:   time.Now(),
			Error:     err,
		})

		p.StatsHandler.HandleProcess(m.Context(), &End{
			MsgCount:  m.Count(),
			BeginTime: m.GetEnqueuedAt(),
			EndTime:   time.Now(),
		})
	}()
}
