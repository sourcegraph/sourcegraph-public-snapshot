package search

import (
	"context"
	"sync"

	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

type matchSender interface {
	Send(protocol.FileMatch)
	SentCount() int
	Remaining() int
	LimitHit() bool
}

type limitedStream struct {
	cb        func(protocol.FileMatch)
	limit     int
	remaining *atomic.Int64
	limitHit  *atomic.Bool
	cancel    context.CancelFunc
}

// newLimitedStream creates a stream that will limit the number of matches passed through it,
// cancelling the context it returns when that happens. For each match sent to the stream,
// if it hasn't hit the limit, it will call the onMatch callback with that match. The onMatch
// callback will never be called concurrently.
func newLimitedStream(ctx context.Context, limit int, cb func(protocol.FileMatch)) (context.Context, context.CancelFunc, *limitedStream) {
	ctx, cancel := context.WithCancel(ctx)
	s := &limitedStream{
		cb:        cb,
		cancel:    cancel,
		limit:     limit,
		remaining: atomic.NewInt64(int64(limit)),
		limitHit:  atomic.NewBool(false),
	}
	return ctx, cancel, s
}

func (m *limitedStream) Send(match protocol.FileMatch) {
	count := int64(match.MatchCount())

	after := m.remaining.Sub(count)
	before := after + count

	if after > 0 {
		// Remaining was large enough to send the full match
		m.cb(match)
	} else if before <= 0 {
		// We had already hit the limit, so just ignore this event
		return
	} else if after == 0 {
		// We hit the limit exactly.
		m.cancel()
		m.limitHit.Store(true)
		m.cb(match)
	} else {
		// We crossed the limit threshold, so we need to truncate the
		// event before sending.
		m.cancel()
		m.limitHit.Store(true)
		match.Limit(int(before))
		m.cb(match)
	}
}

func (m *limitedStream) SentCount() int {
	remaining := int(m.remaining.Load())
	if remaining < 0 {
		remaining = 0
	}
	return m.limit - remaining
}

func (m *limitedStream) Remaining() int {
	remaining := int(m.remaining.Load())
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func (m *limitedStream) LimitHit() bool {
	return m.limitHit.Load()
}

type limitedStreamCollector struct {
	collected []protocol.FileMatch
	mux       sync.Mutex
	*limitedStream
}

func newLimitedStreamCollector(ctx context.Context, limit int) (context.Context, context.CancelFunc, *limitedStreamCollector) {
	s := &limitedStreamCollector{}
	ctx, cancel, ls := newLimitedStream(ctx, limit, func(fm protocol.FileMatch) {
		s.mux.Lock()
		s.collected = append(s.collected, fm)
		s.mux.Unlock()
	})
	s.limitedStream = ls
	return ctx, cancel, s
}
