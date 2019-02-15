package nrsubee

import (
	"context"

	newrelic "github.com/newrelic/go-agent"

	"github.com/wantedly/subee"
)

const (
	// TransactionName is const value for newrelic transaction.
	TransactionName = "Pub/Sub"

	// ConsumeSegmentName is const value for segment of consumption.
	ConsumeSegmentName = "consume"

	// QueueSegmentName is const value for segment of enqueue/dequeue.
	QueueSegmentName = "queue"
)

// NewStatsHandler creates a new subee.StatsHandler instance for measuring application performances with New Relic.
func NewStatsHandler(app newrelic.Application) subee.StatsHandler {
	return &statsHandler{
		app: app,
	}
}

type statsHandler struct {
	app newrelic.Application
}

type (
	queueContextKey   struct{}
	consumeContextKey struct{}
	processContextKey struct{}
)

func (sh *statsHandler) TagProcess(ctx context.Context, t subee.Tag) context.Context {
	switch t.(type) {
	case *subee.EnqueueTag:
		txn := sh.app.StartTransaction(TransactionName, nil, nil)
		ctx := newrelic.NewContext(ctx, txn)
		seg := newrelic.StartSegment(txn, QueueSegmentName)
		return context.WithValue(ctx, queueContextKey{}, seg)

	case *subee.ConsumeBeginTag:
		txn := newrelic.FromContext(ctx)
		seg := newrelic.StartSegment(txn, ConsumeSegmentName)
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
