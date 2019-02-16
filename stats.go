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
	isTag()
}

// Stats is stats information about receive/consume.
type Stats interface {
	isStats()
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

func (*BeginTag) isTag() {}

// End contains stats when an receive/consume process ends.
type End struct {
	MsgCount  int
	BeginTime time.Time
	EndTime   time.Time
}

func (*End) isStats() {}

// ConsumeBeginTag is tag for consumption start.
type ConsumeBeginTag struct{}

func (*ConsumeBeginTag) isTag() {}

// ConsumeEnd contains stats when consume end.
type ConsumeEnd struct {
	BeginTime time.Time
	EndTime   time.Time
	Error     error
}

func (pb *ConsumeEnd) isStats() {}

// EnqueueTag is tag for enqueue in channel.
type EnqueueTag struct{}

func (*EnqueueTag) isTag() {}

// Dequeue contains stats when dequeue in channel.
type Dequeue struct {
	BeginTime time.Time
	EndTime   time.Time
}

func (*Dequeue) isStats() {}
