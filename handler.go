package subee

import (
	"context"
	"time"
)

// StatsHandler is the interface for related stats handling
type StatsHandler interface {
	TagProcess(context.Context, Tag) context.Context
	HandleProcess(context.Context, Stats)
}

// Tag is tag information.
type Tag interface {
	isTag() bool
}

// Stats is stats information about receive/consume.
type Stats interface {
	isStats() bool
}

// NopStatsHandler is no-op StatsHandler
type NopStatsHandler struct{}

// TagProcess returns context without doing anythig
func (*NopStatsHandler) TagProcess(ctx context.Context, t Tag) context.Context {
	return ctx
}

// HandleProcess do nothing
func (*NopStatsHandler) HandleProcess(context.Context, Stats) {}

// BeginTag is  tag for an receive/consume process starts.
type BeginTag struct{}

func (*BeginTag) isTag() bool { return true }

// End contains stats when an receive/consume process ends.
type End struct {
	MsgCount  int
	BeginTime time.Time
	EndTime   time.Time
}

func (*End) isStats() bool { return true }

// ReceiveBeginTag is tag for receive begin.
type ReceiveBeginTag struct{}

func (*ReceiveBeginTag) isTag() bool { return true }

// ReceiveEnd contains stats when an process receive ends.
type ReceiveEnd struct {
	BeginTime time.Time
	EndTime   time.Time
	Error     error
}

func (pb *ReceiveEnd) isStats() bool { return true }

// ConsumeBeginTag is tag for consumption start.
type ConsumeBeginTag struct{}

func (*ConsumeBeginTag) isTag() bool { return true }

// ConsumeEnd contains stats when consume end.
type ConsumeEnd struct {
	BeginTime time.Time
	EndTime   time.Time
	Error     error
}

func (pb *ConsumeEnd) isStats() bool { return true }

// EnqueueTag is tag for enqueue in channel.
type EnqueueTag struct{}

func (*EnqueueTag) isTag() bool { return true }

// Dequeue contains stats when dequeue in channel.
type Dequeue struct {
	BeginTime time.Time
	EndTime   time.Time
}

func (*Dequeue) isStats() bool { return true }
