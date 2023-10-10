package backend

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/zoekt"
)

var (
	metricFinalAggregateSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_zoekt_final_aggregate_size",
		Help:    "The number of file matches we aggregated before flushing",
		Buckets: prometheus.ExponentialBuckets(1, 2, 20),
	}, []string{"reason"})
	metricFinalOverflowSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_zoekt_final_overflow_size",
		Help:    "The number of overflow results we collected before flushing",
		Buckets: prometheus.ExponentialBuckets(1, 2, 20),
	}, []string{"reason"})
)

// collectSender is a sender that will aggregate results. Once sending is
// done, you call Done to return the aggregated result which are ranked.
//
// Note: It aggregates Progress as well, and expects that the
// MaxPendingPriority it receives are monotonically decreasing.
//
// Note: it ignores the top-level fields RepoURLs and LineFragments since we
// do not read those values in Sourcegraph.
type collectSender struct {
	aggregate *zoekt.SearchResult
	overflow  []*zoekt.SearchResult
	opts      *zoekt.SearchOptions
	sizeBytes uint64
}

func newCollectSender(opts *zoekt.SearchOptions) *collectSender {
	return &collectSender{
		opts: opts,
	}
}

func (c *collectSender) Send(r *zoekt.SearchResult) {
	if c.aggregate == nil {
		c.aggregate = &zoekt.SearchResult{}
	}

	c.aggregate.Stats.Add(r.Stats)

	if len(r.Files) > 0 {
		c.aggregate.Files = append(c.aggregate.Files, r.Files...)
	}

	c.sizeBytes += r.SizeBytes()
}

func (c *collectSender) SendOverflow(r *zoekt.SearchResult) {
	if c.overflow == nil {
		c.overflow = []*zoekt.SearchResult{}
	}
	c.overflow = append(c.overflow, r)
	c.sizeBytes += r.SizeBytes()
}

// Done returns the aggregated result. Before returning, the files are
// ranked and truncated according to the input SearchOptions. If an
// endpoint sent any more results its initial ranked result, then these
// are returned as 'overflow' results.
//
// If no results are aggregated, ok is false and both result values are nil.
func (c *collectSender) Done() (_ *zoekt.SearchResult, _ []*zoekt.SearchResult, ok bool) {
	if c.aggregate == nil {
		return nil, nil, false
	}

	agg := c.aggregate
	c.aggregate = nil

	zoekt.SortFiles(agg.Files)
	if max := c.opts.MaxDocDisplayCount; max > 0 && len(agg.Files) > max {
		agg.Files = agg.Files[:max]
	}

	overflow := c.overflow
	c.overflow = nil
	c.sizeBytes = 0

	return agg, overflow, true
}

type FlushSender interface {
	Flush()
	Send(endpoint string, event *zoekt.SearchResult)
	SendDone(endpoint string)
}

type nopFlushCollector struct {
	sender zoekt.Sender
}

func (n *nopFlushCollector) Send(_ string, event *zoekt.SearchResult) {
	n.sender.Send(event)
}

func (n *nopFlushCollector) SendDone(_ string) {}
func (n *nopFlushCollector) Flush()            {}

type flushCollectSender struct {
	mu            sync.Mutex
	collectSender *collectSender
	sender        zoekt.Sender
	// Map of endpoints to boolean, indicating whether we've received their first set of non-empty search results
	firstResults map[string]bool
	maxSizeBytes int
	timerCancel  chan struct{}
}

// newFlushCollectSender creates a sender which will collect and rank results
// until it has received one result from every endpoint. After it flushes that
// ranked result, it will stream out each result as it is received.
//
// If it has not heard back from every endpoint by a certain timeout, then it will
// flush as a 'fallback plan' to avoid delaying the search too much.
func newFlushCollectSender(opts *zoekt.SearchOptions, endpoints []string, maxSizeBytes int, sender zoekt.Sender) FlushSender {
	// Nil options are permitted by Zoekt's "Streamer" interface. There are a few
	// callers within Sourcegraph that call search with nil options (tests,
	// insights), so we have to handle this case.
	if opts == nil {
		return &nopFlushCollector{sender}
	}

	firstResults := map[string]bool{}
	for _, endpoint := range endpoints {
		firstResults[endpoint] = true
	}

	collectSender := newCollectSender(opts)
	timerCancel := make(chan struct{})

	flushSender := &flushCollectSender{collectSender: collectSender,
		sender:       sender,
		firstResults: firstResults,
		maxSizeBytes: maxSizeBytes,
		timerCancel:  timerCancel}

	// As an escape hatch, stop collecting after twice the FlushWallTime. This protects against
	// cases where an endpoint stops being responsive so we never receive its results.
	go func() {
		timer := time.NewTimer(2 * opts.FlushWallTime)
		select {
		case <-timerCancel:
			timer.Stop()
		case <-timer.C:
			flushSender.mu.Lock()
			flushSender.stopCollectingAndFlush(zoekt.FlushReasonTimerExpired)
			flushSender.mu.Unlock()
		}
	}()
	return flushSender
}

// Send consumes a search event. We transition through 3 states
// 1. collectSender != nil: collect results via collectSender
// 2. len(firstResults) == 0: we've received one non-empty result from every endpoint (or the 'done' signal)
// 3. collectSender == nil: directly use sender
func (f *flushCollectSender) Send(endpoint string, event *zoekt.SearchResult) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.collectSender != nil {
		if f.firstResults[endpoint] {
			f.collectSender.Send(event)
			// Ignore first events with no files, like stats-only events
			if len(event.Files) > 0 {
				delete(f.firstResults, endpoint)
			}
		} else {
			f.collectSender.SendOverflow(event)
		}

		if len(f.firstResults) == 0 {
			f.stopCollectingAndFlush(zoekt.FlushReasonTimerExpired)
		} else if f.maxSizeBytes >= 0 && f.collectSender.sizeBytes > uint64(f.maxSizeBytes) {
			// Protect against too large aggregates. This should be the exception and only
			// happen for queries yielding an extreme number of results.
			f.stopCollectingAndFlush(zoekt.FlushReasonMaxSize)
		}
	} else {
		f.sender.Send(event)
	}
}

// SendDone is called to signal that an endpoint is finished streaming results. Some endpoints
// may not return any results, so we must use SendDone to signal their completion.
func (f *flushCollectSender) SendDone(endpoint string) {
	f.mu.Lock()
	delete(f.firstResults, endpoint)
	if len(f.firstResults) == 0 {
		f.stopCollectingAndFlush(zoekt.FlushReasonTimerExpired)
	}
	f.mu.Unlock()
}

// stopCollectingAndFlush will send what we have collected and all future
// sends will go via sender directly.
func (f *flushCollectSender) stopCollectingAndFlush(reason zoekt.FlushReason) {
	if f.collectSender == nil {
		return
	}

	if agg, overflow, ok := f.collectSender.Done(); ok {
		metricFinalAggregateSize.WithLabelValues(reason.String()).Observe(float64(len(agg.Files)))
		metricFinalOverflowSize.WithLabelValues(reason.String()).Observe(float64(len(overflow)))

		agg.FlushReason = reason
		f.sender.Send(agg)

		for _, result := range overflow {
			result.FlushReason = reason
			f.sender.Send(result)
		}
	}

	// From now on use sender directly
	f.collectSender = nil

	// Stop timer goroutine if it is still running.
	close(f.timerCancel)
}

func (f *flushCollectSender) Flush() {
	f.mu.Lock()
	f.stopCollectingAndFlush(zoekt.FlushReasonFinalFlush)
	f.mu.Unlock()
}
