package nrsubee

import (
	"context"

	newrelic "github.com/newrelic/go-agent"

	"github.com/wantedly/subee"
)

// NewStatsHandler creates a new subee.StatsHandler instance for measuring application performances with New Relic.
func NewStatsHandler(app newrelic.Application, opts ...Option) subee.StatsHandler {
	cfg := DefaultConfig()
	cfg.apply(opts)
	return &statsHandler{
		app: app,
		cfg: cfg,
	}
}

type statsHandler struct {
	app newrelic.Application
	cfg *Config
}

type (
	queueContextKey   struct{}
	consumeContextKey struct{}
	processContextKey struct{}
)

func (sh *statsHandler) TagProcess(ctx context.Context, t subee.Tag) context.Context {
	switch t.(type) {
	case *subee.EnqueueTag:
		txn := sh.app.StartTransaction(sh.cfg.TransactionName, nil, nil)
		ctx := newrelic.NewContext(ctx, txn)
		seg := newrelic.StartSegment(txn, sh.cfg.QueueingSegmentName)
		return context.WithValue(ctx, queueContextKey{}, seg)

	case *subee.ConsumeBeginTag:
		txn := newrelic.FromContext(ctx)
		seg := newrelic.StartSegment(txn, sh.cfg.ConsumeSegmentName)
		return context.WithValue(ctx, consumeContextKey{}, seg)
	}

	return ctx
}

func (sh *statsHandler) HandleProcess(ctx context.Context, s subee.Stats) {
	txn := newrelic.FromContext(ctx)

	switch s := s.(type) {
	case *subee.Dequeue:
		ctx.Value(queueContextKey{}).(*newrelic.Segment).End()

	case *subee.ConsumeEnd:
		ctx.Value(consumeContextKey{}).(*newrelic.Segment).End()
		if s.Error != nil {
			txn.NoticeError(s.Error)
		}

	case *subee.End:
		if s.MsgCount > 0 {
			txn.End()
		}
	}
}
