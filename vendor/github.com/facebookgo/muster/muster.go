// Package muster provides a framework for writing libraries that internally
// batch operations.
//
// It will be useful to you if you're building an API that benefits from
// performing work in batches for whatever reason. Batching is triggered based
// on a maximum number of items in a batch, and/or based on a timeout for how
// long a batch waits before it is dispatched. For example if you're willing to
// wait for a maximum of a 1m duration, you can just set BatchTimeout and keep
// adding things. Or if you want batches of 50 just set MaxBatchSize and it
// will only fire when the batch is filled. For best results set both.
//
// It would be in your best interest to use this library in a hidden fashion in
// order to avoid unnecessary coupling. You will typically achieve this by
// ensuring your implementation of muster.Batch and the use of muster.Client
// are private.
package muster

import (
	"errors"
	"sync"
	"time"

	"github.com/facebookgo/clock"
	"github.com/facebookgo/limitgroup"
)

var errZeroBoth = errors.New(
	"muster: MaxBatchSize and BatchTimeout can't both be zero",
)

type waitGroup interface {
	Add(delta int)
	Done()
	Wait()
}

// Notifier is used to indicate to the Client when a batch has finished
// processing.
type Notifier interface {
	// Calling Done will indicate the batch has finished processing.
	Done()
}

// Batch collects added items. Fire will be called exactly once. The Batch does
// not need to be safe for concurrent access; synchronization will be handled
// by the Client.
type Batch interface {
	// This should add the given single item to the Batch. This is the "other
	// end" of the Client.Work channel where your application will send items.
	Add(item interface{})

	// Fire off the Batch. It should call Notifier.Done() when it has finished
	// processing the Batch.
	Fire(notifier Notifier)
}

// The Client manages the background process that makes, populates & fires
// Batches.
type Client struct {
	// Maximum number of items in a batch. If this is zero batches will only be
	// dispatched upon hitting the BatchTimeout. It is an error for both this and
	// the BatchTimeout to be zero.
	MaxBatchSize uint

	// Duration after which to send a pending batch. If this is zero batches will
	// only be dispatched upon hitting the MaxBatchSize. It is an error for both
	// this and the MaxBatchSize to be zero.
	BatchTimeout time.Duration

	// MaxConcurrentBatches determines how many parallel batches we'll allow to
	// be "in flight" concurrently. Once these many batches are in flight, the
	// PendingWorkCapacity determines when sending to the Work channel will start
	// blocking. In other words, once MaxConcurrentBatches hits, the system
	// starts blocking. This allows for tighter control over memory utilization.
	// If not set, the number of parallel batches in-flight will not be limited.
	MaxConcurrentBatches uint

	// Capacity of work channel. If this is zero, the Work channel will be
	// blocking.
	PendingWorkCapacity uint

	// This function should create a new empty Batch on each invocation.
	BatchMaker func() Batch

	// Once this Client has been started, send work items here to add to batch.
	Work chan interface{}

	klock     clock.Clock
	workGroup waitGroup
}

func (c *Client) clock() clock.Clock {
	if c.klock == nil {
		return clock.New()
	}
	return c.klock
}

// Start the background worker goroutines and get ready for accepting requests.
func (c *Client) Start() error {
	if int64(c.BatchTimeout) == 0 && c.MaxBatchSize == 0 {
		return errZeroBoth
	}

	if c.MaxConcurrentBatches == 0 {
		c.workGroup = &sync.WaitGroup{}
	} else {
		c.workGroup = limitgroup.NewLimitGroup(c.MaxConcurrentBatches + 1)
	}

	c.Work = make(chan interface{}, c.PendingWorkCapacity)
	c.workGroup.Add(1) // this is the worker itself
	go c.worker()
	return nil
}

// Stop gracefully and return once all processing has finished.
func (c *Client) Stop() error {
	close(c.Work)
	c.workGroup.Wait()
	return nil
}

// Background process.
func (c *Client) worker() {
	defer c.workGroup.Done()
	var batch = c.BatchMaker()
	var count uint
	var batchTimer *clock.Timer
	var batchTimeout <-chan time.Time
	send := func() {
		c.workGroup.Add(1)
		go batch.Fire(c.workGroup)
		batch = c.BatchMaker()
		count = 0
		if batchTimer != nil {
			batchTimer.Stop()
		}
	}
	recv := func(item interface{}, open bool) bool {
		if !open {
			if count != 0 {
				send()
			}
			return true
		}
		batch.Add(item)
		count++
		if c.MaxBatchSize != 0 && count >= c.MaxBatchSize {
			send()
		} else if int64(c.BatchTimeout) != 0 && count == 1 {
			batchTimer = c.clock().Timer(c.BatchTimeout)
			batchTimeout = batchTimer.C
		}
		return false
	}
	for {
		// We use two selects in order to first prefer draining the work queue.
		select {
		case item, open := <-c.Work:
			if recv(item, open) {
				return
			}
		default:
			select {
			case item, open := <-c.Work:
				if recv(item, open) {
					return
				}
			case <-batchTimeout:
				send()
			}
		}
	}
}
