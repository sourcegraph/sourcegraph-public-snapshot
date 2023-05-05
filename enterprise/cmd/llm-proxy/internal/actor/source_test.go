package actor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockSourceSyncer struct {
	syncCount int
}

var _ SourceSyncer = &mockSourceSyncer{}

func (m *mockSourceSyncer) Name() string { return "mock" }

func (m *mockSourceSyncer) Get(context.Context, string) (*Actor, error) {
	return nil, errors.New("unimplemented")
}

func (m *mockSourceSyncer) Sync(context.Context) error {
	m.syncCount++
	return nil
}

func TestSourcesWorker(t *testing.T) {
	var s mockSourceSyncer
	w := (Sources{&s}).Worker(time.Millisecond)
	stopped := make(chan struct{})

	// Work happens after start
	go func() {
		w.Start()
		stopped <- struct{}{}
	}()
	time.Sleep(9 * time.Millisecond)
	assert.NotZero(t, s.syncCount)

	// No work happens after stop
	w.Stop()
	count := s.syncCount
	time.Sleep(10 * time.Millisecond)
	assert.LessOrEqual(t, count, s.syncCount)

	println("waiting for stop")
	<-stopped
}
