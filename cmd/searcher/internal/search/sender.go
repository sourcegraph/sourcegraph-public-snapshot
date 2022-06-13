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
	matchCount := match.MatchCount()
	if matchCount <= m.remaining {
		m.remaining -= matchCount
		m.sentCount += matchCount
		m.cb(match)
		m.mux.Unlock()
		return
	}

	m.limitHit = true
	m.cancel()

	// Can't truncate a path match
	if len(match.ChunkMatches) == 0 {
		m.mux.Unlock()
		return
	}

	for i, cm := range match.ChunkMatches {
		if l := len(cm.Ranges); l >= m.remaining {
			match.ChunkMatches[i].Ranges = cm.Ranges[:m.remaining]
			match.ChunkMatches = match.ChunkMatches[:i+1]
			break
		} else {
			m.remaining -= l
		}
	}
	match.LimitHit = true
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

type limitedStreamCollector struct {
	collected []protocol.FileMatch
	*limitedStream
}

func newLimitedStreamCollector(ctx context.Context, limit int) (context.Context, context.CancelFunc, *limitedStreamCollector) {
	s := &limitedStreamCollector{}
	ctx, cancel, ls := newLimitedStream(ctx, limit, func(fm protocol.FileMatch) {
		s.collected = append(s.collected, fm)
	})
	s.limitedStream = ls
	return ctx, cancel, s
}
