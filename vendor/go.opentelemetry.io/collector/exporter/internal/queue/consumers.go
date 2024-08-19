// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package queue // import "go.opentelemetry.io/collector/exporter/internal/queue"

import (
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
)

type Consumers[T any] struct {
	queue        Queue[T]
	numConsumers int
	consumeFunc  func(context.Context, T) error
	stopWG       sync.WaitGroup
}

func NewQueueConsumers[T any](q Queue[T], numConsumers int, consumeFunc func(context.Context, T) error) *Consumers[T] {
	return &Consumers[T]{
		queue:        q,
		numConsumers: numConsumers,
		consumeFunc:  consumeFunc,
		stopWG:       sync.WaitGroup{},
	}
}

// Start ensures that queue and all consumers are started.
func (qc *Consumers[T]) Start(ctx context.Context, host component.Host) error {
	if err := qc.queue.Start(ctx, host); err != nil {
		return err
	}

	var startWG sync.WaitGroup
	for i := 0; i < qc.numConsumers; i++ {
		qc.stopWG.Add(1)
		startWG.Add(1)
		go func() {
			startWG.Done()
			defer qc.stopWG.Done()
			for {
				if !qc.queue.Consume(qc.consumeFunc) {
					return
				}
			}
		}()
	}
	startWG.Wait()

	return nil
}

// Shutdown ensures that queue and all consumers are stopped.
func (qc *Consumers[T]) Shutdown(ctx context.Context) error {
	if err := qc.queue.Shutdown(ctx); err != nil {
		return err
	}
	qc.stopWG.Wait()
	return nil
}
