package backend

import (
	"sync"
	"time"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/stream"
)

// collectSender is a sender that will aggregate results. Once sending is
// done, you call Done to return the aggregated result which are ranked.
//
// Note: It aggregates Progress as well, and expects that the
// MaxPendingPriority it receives are monotonically decreasing.
type collectSender struct {
	aggregate          *zoekt.SearchResult
	maxDocDisplayCount int
}

type collectOpts struct {
	maxDocDisplayCount int
	flushWallTime      time.Duration
}

func newCollectSender(opts *collectOpts) *collectSender {
	return &collectSender{
		maxDocDisplayCount: opts.maxDocDisplayCount,
	}
}

func (c *collectSender) Send(r *zoekt.SearchResult) {
	if c.aggregate == nil {
		c.aggregate = &zoekt.SearchResult{
			RepoURLs:      map[string]string{},
			LineFragments: map[string]string{},
		}
	}

	c.aggregate.Stats.Add(r.Stats)

	if len(r.Files) > 0 {
		c.aggregate.Files = append(c.aggregate.Files, r.Files...)

		for k, v := range r.RepoURLs {
			c.aggregate.RepoURLs[k] = v
		}
		for k, v := range r.LineFragments {
			c.aggregate.LineFragments[k] = v
		}
	}

	// The priority of our aggregate is the largest priority we collect.
	if c.aggregate.Priority < r.Priority {
		c.aggregate.Priority = r.Priority
	}

	// We receive monotonically decreasing values, so we update on every call.
	c.aggregate.MaxPendingPriority = r.MaxPendingPriority
}

// Done returns the aggregated result. Before returning them the files are
// ranked and truncated according to the input SearchOptions.
//
// If no results are aggregated, ok is false and the result is nil.
func (c *collectSender) Done() (_ *zoekt.SearchResult, ok bool) {
	if c.aggregate == nil {
		return nil, false
	}

	agg := c.aggregate
	c.aggregate = nil

	zoekt.SortFiles(agg.Files)

	if max := c.maxDocDisplayCount; max > 0 && len(agg.Files) > max {
		agg.Files = agg.Files[:max]
	}

	return agg, true
}

// newFlushCollectSender creates a sender which will collect and rank results
// until opts.flushWallTime. After that it will stream each result as it is
// sent.
func newFlushCollectSender(opts *collectOpts, sender zoekt.Sender) (zoekt.Sender, func()) {
	// We don't need to do any collecting, so just pass back the sender to use
	// directly.
	if opts.flushWallTime == 0 {
		return sender, func() {}
	}

	// We transition through 3 states
	// 1. collectSender != nil: collect results via collectSender
	// 2. timerFired: send collected results and mark collectSender nil
	// 3. collectSender == nil: directly use sender

	var (
		mu            sync.Mutex
		collectSender = newCollectSender(opts)
		timerCancel   = make(chan struct{})
	)

	// stopCollectingAndFlush will send what we have collected and all future
	// sends will go via sender directly.
	stopCollectingAndFlush := func() {
		mu.Lock()
		defer mu.Unlock()

		if collectSender == nil {
			return
		}

		if agg, ok := collectSender.Done(); ok {
			sender.Send(agg)
		}

		// From now on use sender directly
		collectSender = nil

		// Stop timer goroutine if it is still running.
		close(timerCancel)
	}

	// Wait flushWallTime to call stopCollecting.
	go func() {
		timer := time.NewTimer(opts.flushWallTime)
		select {
		case <-timerCancel:
			timer.Stop()
		case <-timer.C:
			stopCollectingAndFlush()
		}
	}()

	return stream.SenderFunc(func(event *zoekt.SearchResult) {
		mu.Lock()
		if collectSender != nil {
			collectSender.Send(event)
		} else {
			sender.Send(event)
		}
		mu.Unlock()
	}), stopCollectingAndFlush
}
