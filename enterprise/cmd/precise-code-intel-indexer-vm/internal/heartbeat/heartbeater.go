package heartbeat

import (
	"context"
	"sync"
	"time"

	"github.com/efritz/glock"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	queue "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
)

type Heartbeater struct {
	queueClient  queue.Client
	indexManager *indexmanager.Manager
	options      HeartbeaterOptions
	clock        glock.Clock
	ctx          context.Context
	cancel       func()
	wg           sync.WaitGroup
	finished     chan struct{}
}

type HeartbeaterOptions struct {
	Interval time.Duration
}

func NewHeartbeater(ctx context.Context, queueClient queue.Client, indexManager *indexmanager.Manager, options HeartbeaterOptions) *Heartbeater {
	return newHeartbeater(ctx, queueClient, indexManager, options, glock.NewRealClock())
}

func newHeartbeater(ctx context.Context, queueClient queue.Client, indexManager *indexmanager.Manager, options HeartbeaterOptions, clock glock.Clock) *Heartbeater {
	ctx, cancel := context.WithCancel(ctx)

	return &Heartbeater{
		queueClient:  queueClient,
		indexManager: indexManager,
		options:      options,
		clock:        clock,
		ctx:          ctx,
		cancel:       cancel,
		finished:     make(chan struct{}),
	}
}

func (w *Heartbeater) Start() {
	defer close(w.finished)

loop:
	for {
		if err := w.queueClient.Heartbeat(w.ctx, w.indexManager.GetIDs()); err != nil {
			for ex := err; ex != nil; ex = errors.Unwrap(ex) {
				if err == w.ctx.Err() {
					break loop
				}
			}

			log15.Error("Failed to perform heartbeat", "err", err)
		}

		select {
		case <-w.clock.After(w.options.Interval):
		case <-w.ctx.Done():
			break loop
		}
	}

	w.wg.Wait()
}

func (w *Heartbeater) Stop() {
	w.cancel()
	<-w.finished
}
