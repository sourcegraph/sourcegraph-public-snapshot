package search

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

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
