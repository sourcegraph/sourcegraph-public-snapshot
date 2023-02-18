package backend

import (
	"sync"

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
)

// collectSender is a sender that will aggregate results. Once sending is
// done, you call Done to return the aggregated result which are ranked.
//
// Note: It aggregates Progress as well, and expects that the
// MaxPendingPriority it receives are monotonically decreasing.
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

	c.sizeBytes += r.SizeBytes()
}

func (c *collectSender) SendOverflow(r *zoekt.SearchResult) {
	if c.overflow == nil {
		c.overflow = []*zoekt.SearchResult{}
	}
	c.overflow = append(c.overflow, r)
	c.sizeBytes += r.SizeBytes()
}

// Done returns the aggregated result. Before returning them the files are
// ranked and truncated according to the input SearchOptions.
//
// If no results are aggregated, ok is false and the result is nil.
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

type FlushCollectSender struct {
	mu                 sync.Mutex
	remainingEndpoints map[string]bool
	maxSizeBytes       int
	collectSender      *collectSender
	sender             zoekt.Sender
}

// newFlushCollectSender creates a sender which will collect and rank results
// until it has received one result from every endpoint. After that it will stream
// each result as it is sent.
func newFlushCollectSender(opts *zoekt.SearchOptions, endpoints []string, maxSizeBytes int, sender zoekt.Sender) *FlushCollectSender {
	remainingEndpoints := map[string]bool{}
	for _, endpoint := range endpoints {
		remainingEndpoints[endpoint] = true
	}

	collectSender := newCollectSender(opts)
	return &FlushCollectSender{remainingEndpoints: remainingEndpoints, maxSizeBytes: maxSizeBytes, collectSender: collectSender, sender: sender}
}

// Send consumes a search event. We transition through 3 states
// 1. collectSender != nil: collect results via collectSender
// 2. we've received one result from every endpoint (or the 'done' signal)
// 3. collectSender == nil: directly use sender
func (f *FlushCollectSender) Send(endpoint string, event *zoekt.SearchResult) {
	f.mu.Lock()
	if f.collectSender != nil {
		firstEvent := f.remainingEndpoints[endpoint]
		if firstEvent {
			f.collectSender.Send(event)
			delete(f.remainingEndpoints, endpoint)
		} else {
			f.collectSender.SendOverflow(event)
		}

		if len(f.remainingEndpoints) == 0 {
			f.stopCollectingAndFlush(zoekt.FlushReasonTimerExpired)
		}

		// Protect against too large aggregates. This should be the exception and only
		// happen for queries yielding an extreme number of results.
		if f.maxSizeBytes >= 0 && f.collectSender.sizeBytes > uint64(f.maxSizeBytes) {
			f.stopCollectingAndFlush(zoekt.FlushReasonMaxSize)
		}
	} else {
		f.sender.Send(event)
	}
	f.mu.Unlock()
}

// SendDone is called to signal that an endpoint is finished streaming results. Some endpoints
// may not return any results, so we must use SendDone to signal their completion.
func (f *FlushCollectSender) SendDone(endpoint string) {
	f.mu.Lock()
	if f.collectSender != nil {
		delete(f.remainingEndpoints, endpoint)
		if len(f.remainingEndpoints) == 0 {
			f.stopCollectingAndFlush(zoekt.FlushReasonTimerExpired)
		}
	}
	f.mu.Unlock()
}

// stopCollectingAndFlush will send what we have collected and all future
// sends will go via sender directly.
func (f *FlushCollectSender) stopCollectingAndFlush(reason zoekt.FlushReason) {
	if f.collectSender == nil {
		return
	}

	if agg, overflow, ok := f.collectSender.Done(); ok {
		metricFinalAggregateSize.WithLabelValues(reason.String()).Observe(float64(len(agg.Files)))
		agg.FlushReason = reason
		f.sender.Send(agg)

		for _, result := range overflow {
			f.sender.Send(result)
		}
	}

	// From now on use sender directly
	f.collectSender = nil
}

func (f *FlushCollectSender) Flush() {
	f.mu.Lock()
	f.stopCollectingAndFlush(zoekt.FlushReasonFinalFlush)
	f.mu.Unlock()
}
