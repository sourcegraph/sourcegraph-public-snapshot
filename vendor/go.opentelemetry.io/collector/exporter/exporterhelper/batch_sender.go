// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterbatcher"
)

// batchSender is a component that places requests into batches before passing them to the downstream senders.
// Batches are sent out with any of the following conditions:
// - batch size reaches cfg.MinSizeItems
// - cfg.FlushTimeout is elapsed since the timestamp when the previous batch was sent out.
// - concurrencyLimit is reached.
type batchSender struct {
	baseRequestSender
	cfg            exporterbatcher.Config
	mergeFunc      exporterbatcher.BatchMergeFunc[Request]
	mergeSplitFunc exporterbatcher.BatchMergeSplitFunc[Request]

	// concurrencyLimit is the maximum number of goroutines that can be blocked by the batcher.
	// If this number is reached and all the goroutines are busy, the batch will be sent right away.
	// Populated from the number of queue consumers if queue is enabled.
	concurrencyLimit int64
	activeRequests   atomic.Int64

	mu          sync.Mutex
	activeBatch *batch
	lastFlushed time.Time

	logger *zap.Logger

	shutdownCh         chan struct{}
	shutdownCompleteCh chan struct{}
	stopped            *atomic.Bool
}

// newBatchSender returns a new batch consumer component.
func newBatchSender(cfg exporterbatcher.Config, set exporter.Settings,
	mf exporterbatcher.BatchMergeFunc[Request], msf exporterbatcher.BatchMergeSplitFunc[Request]) *batchSender {
	bs := &batchSender{
		activeBatch:        newEmptyBatch(),
		cfg:                cfg,
		logger:             set.Logger,
		mergeFunc:          mf,
		mergeSplitFunc:     msf,
		shutdownCh:         nil,
		shutdownCompleteCh: make(chan struct{}),
		stopped:            &atomic.Bool{},
	}
	return bs
}

func (bs *batchSender) Start(_ context.Context, _ component.Host) error {
	bs.shutdownCh = make(chan struct{})
	timer := time.NewTimer(bs.cfg.FlushTimeout)
	go func() {
		for {
			select {
			case <-bs.shutdownCh:
				// There is a minimal chance that another request is added after the shutdown signal.
				// This loop will handle that case.
				for bs.activeRequests.Load() > 0 {
					bs.mu.Lock()
					if bs.activeBatch.request != nil {
						bs.exportActiveBatch()
					}
					bs.mu.Unlock()
				}
				if !timer.Stop() {
					<-timer.C
				}
				close(bs.shutdownCompleteCh)
				return
			case <-timer.C:
				bs.mu.Lock()
				nextFlush := bs.cfg.FlushTimeout
				if bs.activeBatch.request != nil {
					sinceLastFlush := time.Since(bs.lastFlushed)
					if sinceLastFlush >= bs.cfg.FlushTimeout {
						bs.exportActiveBatch()
					} else {
						nextFlush = bs.cfg.FlushTimeout - sinceLastFlush
					}
				}
				bs.mu.Unlock()
				timer.Reset(nextFlush)
			}
		}
	}()

	return nil
}

type batch struct {
	ctx     context.Context
	request Request
	done    chan struct{}
	err     error

	// requestsBlocked is the number of requests blocked in this batch
	// that can be immediately released from activeRequests when batch sending completes.
	requestsBlocked int64
}

func newEmptyBatch() *batch {
	return &batch{
		ctx:  context.Background(),
		done: make(chan struct{}),
	}
}

// exportActiveBatch exports the active batch asynchronously and replaces it with a new one.
// Caller must hold the lock.
func (bs *batchSender) exportActiveBatch() {
	go func(b *batch) {
		b.err = bs.nextSender.send(b.ctx, b.request)
		close(b.done)
		bs.activeRequests.Add(-b.requestsBlocked)
	}(bs.activeBatch)
	bs.lastFlushed = time.Now()
	bs.activeBatch = newEmptyBatch()
}

// isActiveBatchReady returns true if the active batch is ready to be exported.
// The batch is ready if it has reached the minimum size or the concurrency limit is reached.
// Caller must hold the lock.
func (bs *batchSender) isActiveBatchReady() bool {
	return bs.activeBatch.request.ItemsCount() >= bs.cfg.MinSizeItems ||
		(bs.concurrencyLimit > 0 && bs.activeRequests.Load() >= bs.concurrencyLimit)
}

func (bs *batchSender) send(ctx context.Context, req Request) error {
	// Stopped batch sender should act as pass-through to allow the queue to be drained.
	if bs.stopped.Load() {
		return bs.nextSender.send(ctx, req)
	}

	if bs.cfg.MaxSizeItems > 0 {
		return bs.sendMergeSplitBatch(ctx, req)
	}
	return bs.sendMergeBatch(ctx, req)
}

// sendMergeSplitBatch sends the request to the batch which may be split into multiple requests.
func (bs *batchSender) sendMergeSplitBatch(ctx context.Context, req Request) error {
	bs.mu.Lock()

	reqs, err := bs.mergeSplitFunc(ctx, bs.cfg.MaxSizeConfig, bs.activeBatch.request, req)
	if err != nil || len(reqs) == 0 {
		bs.mu.Unlock()
		return err
	}

	bs.activeRequests.Add(1)
	if len(reqs) == 1 {
		bs.activeBatch.requestsBlocked++
	} else {
		// if there was a split, we want to make sure that bs.activeRequests is released once all of the parts are sent instead of using batch.requestsBlocked
		defer bs.activeRequests.Add(-1)
	}
	if len(reqs) == 1 || bs.activeBatch.request != nil {
		bs.updateActiveBatch(ctx, reqs[0])
		batch := bs.activeBatch
		if bs.isActiveBatchReady() || len(reqs) > 1 {
			bs.exportActiveBatch()
		}
		bs.mu.Unlock()
		<-batch.done
		if batch.err != nil {
			return batch.err
		}
		reqs = reqs[1:]
	} else {
		bs.mu.Unlock()
	}

	// Intentionally do not put the last request in the active batch to not block it.
	// TODO: Consider including the partial request in the error to avoid double publishing.
	for _, r := range reqs {
		if err := bs.nextSender.send(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// sendMergeBatch sends the request to the batch and waits for the batch to be exported.
func (bs *batchSender) sendMergeBatch(ctx context.Context, req Request) error {
	bs.mu.Lock()

	if bs.activeBatch.request != nil {
		var err error
		req, err = bs.mergeFunc(ctx, bs.activeBatch.request, req)
		if err != nil {
			bs.mu.Unlock()
			return err
		}
	}

	bs.activeRequests.Add(1)
	bs.updateActiveBatch(ctx, req)
	batch := bs.activeBatch
	batch.requestsBlocked++
	if bs.isActiveBatchReady() {
		bs.exportActiveBatch()
	}
	bs.mu.Unlock()
	<-batch.done
	return batch.err
}

// updateActiveBatch update the active batch to the new merged request and context.
// The context is only set once and is not updated after the first call.
// Merging the context would be complex and require an additional goroutine to handle the context cancellation.
// We take the approach of using the context from the first request since it's likely to have the shortest timeout.
func (bs *batchSender) updateActiveBatch(ctx context.Context, req Request) {
	if bs.activeBatch.request == nil {
		bs.activeBatch.ctx = ctx
	}
	bs.activeBatch.request = req
}

func (bs *batchSender) Shutdown(context.Context) error {
	bs.stopped.Store(true)
	if bs.shutdownCh != nil {
		close(bs.shutdownCh)
		<-bs.shutdownCompleteCh
	}
	return nil
}
