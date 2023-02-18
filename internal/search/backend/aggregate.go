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
	c.sizeBytes = 0

	zoekt.SortFiles(agg.Files)

	if max := c.opts.MaxDocDisplayCount; max > 0 && len(agg.Files) > max {
		agg.Files = agg.Files[:max]
	}

	return agg, true
}

type FlushCollectSender struct {
	mu            sync.Mutex
	maxSizeBytes  int
	collectSender *collectSender
	sender        zoekt.Sender
}

// newFlushCollectSender creates a sender which will collect and rank results
// until a stopping condition. After that it will stream each result as it is
// sent.
func newFlushCollectSender(opts *zoekt.SearchOptions, maxSizeBytes int, sender zoekt.Sender) *FlushCollectSender {
	return &FlushCollectSender{maxSizeBytes: maxSizeBytes, collectSender: newCollectSender(opts), sender: sender}
}

// Send consumes a search event. We transition through 3 states
// 1. collectSender != nil: collect results via collectSender
// 2. stopping condition hit
// 3. collectSender == nil: directly use sender
func (f *FlushCollectSender) Send(event *zoekt.SearchResult) {
	f.mu.Lock()
	if f.collectSender != nil {
		f.collectSender.Send(event)

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

// stopCollectingAndFlush will send what we have collected and all future
// sends will go via sender directly.
func (f *FlushCollectSender) stopCollectingAndFlush(reason zoekt.FlushReason) {
	if f.collectSender == nil {
		return
	}

	if agg, ok := f.collectSender.Done(); ok {
		metricFinalAggregateSize.WithLabelValues(reason.String()).Observe(float64(len(agg.Files)))
		agg.FlushReason = reason
		f.sender.Send(agg)
	}

	// From now on use sender directly
	f.collectSender = nil
}

func (f *FlushCollectSender) Flush() {
	f.mu.Lock()
	f.stopCollectingAndFlush(zoekt.FlushReasonFinalFlush)
	f.mu.Unlock()
}
