package search

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

type MatchSender interface {
	Send(protocol.FileMatch)
	SentCount() int
}

type collectingSender struct {
	mux       sync.Mutex
	collected []protocol.FileMatch
	sentCount int
}

func (m *collectingSender) Send(match protocol.FileMatch) {
	m.mux.Lock()
	m.collected = append(m.collected, match)
	m.sentCount += match.MatchCount
	m.mux.Unlock()
}

func (m *collectingSender) SentCount() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.sentCount
}
