package search

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

type matchSender interface {
	Send(protocol.FileMatch)
	SentCount() int
	Remaining() int
	LimitHit() bool
}

type limitedStreamCollector struct {
	mux       sync.Mutex
	collected []protocol.FileMatch
	sentCount int
	remaining int
	limitHit  bool
	cancel    context.CancelFunc
}

func newLimitedStreamCollector(ctx context.Context, limit int) (context.Context, context.CancelFunc, *limitedStreamCollector) {
	ctx, cancel := context.WithCancel(ctx)
	s := &limitedStreamCollector{
		cancel:    cancel,
		remaining: limit,
	}
	return ctx, cancel, s
}

func (m *limitedStreamCollector) Send(match protocol.FileMatch) {
	m.mux.Lock()
	if match.MatchCount <= m.remaining {
		m.collected = append(m.collected, match)
		m.remaining -= match.MatchCount
		m.sentCount += match.MatchCount
		m.mux.Unlock()
		return
	}

	m.limitHit = true
	m.cancel()

	// Can't truncate a path match
	if len(match.LineMatches) == 0 {
		m.mux.Unlock()
		return
	}

	// NOTE: this isn't strictly correct for structural search matches
	// since a single match can be multiple lines. However, by the time we
	// convert a structural search to a protocol.FileMatch, we lose the
	// information required to properly limit. However, multiline matches
	// are also not limited correctly in the frontend, so doing it correctly
	// here won't fix that.
	match.LineMatches = match.LineMatches[:m.remaining]
	match.LimitHit = true
	match.MatchCount = m.remaining
	m.sentCount += m.remaining
	m.remaining = 0
	m.collected = append(m.collected, match)
	m.mux.Unlock()
}

func (m *limitedStreamCollector) SentCount() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return len(m.collected)
}

func (m *limitedStreamCollector) Collected() []protocol.FileMatch {
	return m.collected
}

func (m *limitedStreamCollector) Remaining() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.remaining
}

func (m *limitedStreamCollector) LimitHit() bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.limitHit
}

type limitedStream struct {
	cb        func(protocol.FileMatch)
	mux       sync.Mutex
	sentCount int
	remaining int
	limitHit  bool
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
		remaining: limit,
	}
	return ctx, cancel, s
}

func (m *limitedStream) Send(match protocol.FileMatch) {
	m.mux.Lock()
	if match.MatchCount <= m.remaining {
		m.remaining -= match.MatchCount
		m.sentCount += match.MatchCount
		m.cb(match)
		m.mux.Unlock()
		return
	}

	m.limitHit = true
	m.cancel()

	// Can't truncate a path match
	if len(match.LineMatches) == 0 {
		m.mux.Unlock()
		return
	}

	// NOTE: this isn't strictly correct for structural search matches
	// since a single match can be multiple lines. However, by the time we
	// convert a structural search to a protocol.FileMatch, we lose the
	// information required to properly limit. However, multiline matches
	// are also not limited correctly in the frontend, so doing it correctly
	// here won't fix that.
	match.LineMatches = match.LineMatches[:m.remaining]
	match.LimitHit = true
	match.MatchCount = m.remaining
	m.sentCount += m.remaining
	m.remaining = 0
	m.cb(match)
	m.mux.Unlock()
}

func (m *limitedStream) SentCount() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.sentCount
}

func (m *limitedStream) Remaining() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.remaining
}

func (m *limitedStream) LimitHit() bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.limitHit
}
