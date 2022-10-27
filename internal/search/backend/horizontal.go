package backend

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
	"github.com/sourcegraph/zoekt/stream"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	metricReorderQueueSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_zoekt_reorder_queue_size",
		Help:    "Maximum size of result reordering buffer for a request.",
		Buckets: prometheus.ExponentialBuckets(4, 2, 10),
	}, nil)
	metricIgnoredError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_zoekt_ignored_error_total",
		Help: "Total number of errors ignored from Zoekt.",
	}, []string{"reason"})
	// temporary metric so we can check if we are encountering non-empty
	// queues once streaming is complete.
	metricFinalQueueSize = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_zoekt_final_queue_size",
		Help: "the size of the results queue once streaming is done.",
	})
	metricMaxMatchCount = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_zoekt_queue_max_match_count",
		Help:    "Maximum number of matches in the queue.",
		Buckets: prometheus.ExponentialBuckets(4, 2, 20),
	}, nil)
	metricMaxSizeBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_zoekt_queue_max_size_bytes",
		Help:    "Maximum number of bytes in the queue.",
		Buckets: prometheus.ExponentialBuckets(1000, 2, 20), // 1kb -> 500mb
	}, nil)
)

// HorizontalSearcher is a Streamer which aggregates searches over
// Map. It manages the connections to Map as the endpoints come and go.
type HorizontalSearcher struct {
	// Map is a subset of EndpointMap only using the Endpoints function. We
	// use this to find the endpoints to dial over time.
	Map interface {
		Endpoints() ([]string, error)
	}
	Dial func(endpoint string) zoekt.Streamer

	mu      sync.RWMutex
	clients map[string]zoekt.Streamer // addr -> client
}

var rankingEnabled, _ = strconv.ParseBool(os.Getenv("ENABLE_EXPERIMENTAL_RANKING"))

// StreamSearch does a search which merges the stream from every endpoint in Map, reordering results to produce a sorted stream.
func (s *HorizontalSearcher) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, streamer zoekt.Sender) error {

	if rankingEnabled {
		return s.streamSearchExperimentalRanking(ctx, q, opts, streamer)
	}

	clients, err := s.searchers()
	if err != nil {
		return err
	}

	endpoints := make([]string, 0, len(clients))
	for endpoint := range clients {
		endpoints = append(endpoints, endpoint)
	}

	siteConfig := newRankingSiteConfig(conf.Get().SiteConfiguration)

	// rq is used to re-order results by priority.
	var mu sync.Mutex
	rq := newResultQueue(siteConfig, endpoints)

	// Flush the queue latest after maxReorderDuration. The longer
	// maxReorderDuration, the more stable the ranking and the more MEM pressure we
	// put on frontend. maxReorderDuration is effectively the budget we give each
	// Zoekt to produce its highest ranking result. It should be large enough to
	// give each Zoekt the chance to search at least 1 maximum size simple shard
	// plus time spent on network.
	//
	// At the same time maxReorderDuration guarantees a minimum response time. It
	// protects us from waiting on slow Zoekts for too long.
	//
	// maxReorderDuration and maxQueueDepth are tightly connected: If the queue is
	// too short we will always flush before reaching maxReorderDuration and if the
	// queue is too long we risk OOMs of frontend for queries with a lot of results.
	//
	// maxQueueDepth should be chosen as large as possible given the available
	// resources.
	if siteConfig.maxReorderDuration > 0 {
		done := make(chan struct{})
		defer close(done)

		// we can race with done being closed and as such call FlushAll after
		// the return of the function. So track if the function has exited.
		searchDone := false
		defer func() {
			mu.Lock()
			searchDone = true
			mu.Unlock()
		}()

		go func() {
			select {
			case <-done:
			case <-time.After(siteConfig.maxReorderDuration):
				mu.Lock()
				defer mu.Unlock()
				if searchDone {
					return
				}
				rq.FlushAll(streamer)
			}
		}()
	}

	// During rebalancing a repository can appear on more than one replica.
	dedupper := dedupper{}

	// GobCache exists so we only pay the cost of marshalling a query once
	// when we aggregate it out over all the replicas. Zoekt's RPC layers
	// unwrap this before passing it on to the Zoekt evaluation layers.
	q = &query.GobCache{Q: q}

	ch := make(chan error, len(clients))
	for endpoint, c := range clients {
		go func(endpoint string, c zoekt.Streamer) {
			err := c.StreamSearch(ctx, q, opts, stream.SenderFunc(func(sr *zoekt.SearchResult) {
				// This shouldn't happen, but skip event if sr is nil.
				if sr == nil {
					return
				}

				mu.Lock()
				defer mu.Unlock()

				sr.Files = dedupper.Dedup(endpoint, sr.Files)

				rq.Enqueue(endpoint, sr)
				rq.FlushReady(streamer)
			}))
			mu.Lock()
			rq.Done(endpoint)
			mu.Unlock()

			if canIgnoreError(ctx, err) {
				err = nil
			}

			ch <- err
		}(endpoint, c)
	}

	var errs errors.MultiError
	for i := 0; i < cap(ch); i++ {
		errs = errors.Append(errs, <-ch)
	}
	if errs != nil {
		return errs
	}

	mu.Lock()
	metricReorderQueueSize.WithLabelValues().Observe(float64(rq.metricMaxLength))
	metricMaxMatchCount.WithLabelValues().Observe(float64(rq.metricMaxMatchCount))
	metricFinalQueueSize.Add(float64(rq.queue.Len()))
	metricMaxSizeBytes.WithLabelValues().Observe(float64(rq.metricMaxSizeBytes))

	rq.FlushAll(streamer)
	mu.Unlock()

	return nil
}

type rankingSiteConfig struct {
	maxQueueDepth      int
	maxMatchCount      int
	maxSizeBytes       int
	maxReorderDuration time.Duration
}

func (s *HorizontalSearcher) streamSearchExperimentalRanking(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, streamer zoekt.Sender) error {
	clients, err := s.searchers()
	if err != nil {
		return err
	}

	endpoints := make([]string, 0, len(clients))
	for endpoint := range clients {
		endpoints = append(endpoints, endpoint)
	}

	siteConfig := newRankingSiteConfig(conf.Get().SiteConfiguration)

	// Hack: 500ms is a better default for this function. The original default of 0
	// disables the flushCollectSender and offers no ranking at all. For now
	// StreamSearch and streamSearchExperimentalRanking share newRankingSiteConfig.
	// Once this function is not behind a feature flag anymore, we should update the
	// default.
	if siteConfig.maxReorderDuration == 0 {
		siteConfig.maxReorderDuration = 500 * time.Millisecond
	}

	streamer, flushAll := newFlushCollectSender(
		&collectOpts{
			maxDocDisplayCount: opts.MaxDocDisplayCount,
			flushWallTime:      siteConfig.maxReorderDuration,
			maxSizeBytes:       siteConfig.maxSizeBytes,
		},
		streamer,
	)
	defer flushAll()

	// During re-balancing a repository can appear on more than one replica.
	var mu sync.Mutex
	dedupper := dedupper{}

	// GobCache exists, so we only pay the cost of marshalling a query once
	// when we aggregate it out over all the replicas. Zoekt's RPC layers
	// unwrap this before passing it on to the Zoekt evaluation layers.
	q = &query.GobCache{Q: q}

	ch := make(chan error, len(clients))
	for endpoint, c := range clients {
		go func(endpoint string, c zoekt.Streamer) {
			err := c.StreamSearch(ctx, q, opts, stream.SenderFunc(func(sr *zoekt.SearchResult) {
				// This shouldn't happen, but skip event if sr is nil.
				if sr == nil {
					return
				}

				mu.Lock()
				sr.Files = dedupper.Dedup(endpoint, sr.Files)
				mu.Unlock()

				streamer.Send(sr)
			}))

			if canIgnoreError(ctx, err) {
				err = nil
			}

			ch <- err
		}(endpoint, c)
	}

	var errs errors.MultiError
	for i := 0; i < cap(ch); i++ {
		errs = errors.Append(errs, <-ch)
	}

	return errs
}

func newRankingSiteConfig(siteConfig schema.SiteConfiguration) *rankingSiteConfig {
	// defaults
	c := &rankingSiteConfig{
		maxQueueDepth:      24,
		maxMatchCount:      -1,
		maxReorderDuration: 0,
		maxSizeBytes:       -1,
	}

	if siteConfig.ExperimentalFeatures == nil || siteConfig.ExperimentalFeatures.Ranking == nil {
		return c
	}

	if siteConfig.ExperimentalFeatures.Ranking.MaxReorderQueueSize != nil {
		c.maxQueueDepth = *siteConfig.ExperimentalFeatures.Ranking.MaxReorderQueueSize
	}

	if siteConfig.ExperimentalFeatures.Ranking.MaxQueueMatchCount != nil {
		c.maxMatchCount = *siteConfig.ExperimentalFeatures.Ranking.MaxQueueMatchCount
	}

	if siteConfig.ExperimentalFeatures.Ranking.MaxQueueSizeBytes != nil {
		c.maxSizeBytes = *siteConfig.ExperimentalFeatures.Ranking.MaxQueueSizeBytes
	}

	c.maxReorderDuration = time.Duration(siteConfig.ExperimentalFeatures.Ranking.MaxReorderDurationMS) * time.Millisecond

	return c
}

// The results from each endpoint are mostly sorted by priority, with bounded
// errors described by SearchResult.Stats.MaxPendingPriority. Each backend
// will dispatch searches in parallel against its shards in priority order,
// but the actual return order of those searches is not constrained.
//
// Instead, the backend will report the maximum priority shard that it still
// has pending along with the results that it returns, so we accumulate
// results in a heap and only pop when the top item has a priority greater
// than the maximum of all endpoints' pending results.
type resultQueue struct {
	// maxQueueDepth will flush any items in the queue such that we never
	// exceed maxQueueDepth. This is used to prevent aggregating too many
	// results in memory.
	maxQueueDepth int

	// maxMatchCount will flush any items in the queue such that we never exceed
	// maxMatchCount. This is used to prevent aggregating too many results in
	// memory.
	maxMatchCount int

	// The number of matches currently in the queue. We keep track of the current
	// matchCount separately from the stats, because the stats are reset with
	// every event we sent.
	matchCount          int
	metricMaxMatchCount int

	// The approximate size of the queue's content in memory.
	sizeBytes uint64

	// Set by site-config, which does not support uint64. In practice this should be
	// fine. We flush once we reach the threshold of maxSizeBytes.
	maxSizeBytes       int
	metricMaxSizeBytes uint64

	queue           priorityQueue
	metricMaxLength int // for a prometheus metric

	endpointMaxPendingPriority map[string]float64

	// We aggregate statistics independently of matches. Everytime we
	// encounter matches to send we send the current aggregated stats then
	// reset them. This is because in a lot of cases we only get stats and no
	// matches. By aggregating we avoid spamming the sender and the
	// resultQueue with pure stats events.
	stats zoekt.Stats
}

func newResultQueue(siteConfig *rankingSiteConfig, endpoints []string) *resultQueue {
	// To start, initialize every endpoint's maxPending to +inf since we don't yet know the bounds.
	endpointMaxPendingPriority := map[string]float64{}
	for _, endpoint := range endpoints {
		endpointMaxPendingPriority[endpoint] = math.Inf(1)
	}

	return &resultQueue{
		maxQueueDepth:              siteConfig.maxQueueDepth,
		maxMatchCount:              siteConfig.maxMatchCount,
		maxSizeBytes:               siteConfig.maxSizeBytes,
		endpointMaxPendingPriority: endpointMaxPendingPriority,
	}
}

// Enqueue adds the result to the queue and updates the max pending priority
// for endpoint based on sr.
func (q *resultQueue) Enqueue(endpoint string, sr *zoekt.SearchResult) {
	// Update aggregate stats
	q.stats.Add(sr.Stats)

	q.matchCount += sr.MatchCount
	if q.matchCount > q.metricMaxMatchCount {
		q.metricMaxMatchCount = q.matchCount
	}

	sb := sr.SizeBytes()
	q.sizeBytes += sb
	if q.sizeBytes >= q.metricMaxSizeBytes {
		q.metricMaxSizeBytes = q.sizeBytes
	}

	// Note the endpoint's updated MaxPendingPriority
	q.endpointMaxPendingPriority[endpoint] = sr.Progress.MaxPendingPriority

	// Don't add empty results to the heap.
	if len(sr.Files) != 0 {
		q.queue.add(&queueSearchResult{SearchResult: sr, sizeBytes: sb})
		if q.queue.Len() > q.metricMaxLength {
			q.metricMaxLength = q.queue.Len()
		}
	}
}

// Done must be called once per endpoint once it has finished streaming.
func (q *resultQueue) Done(endpoint string) {
	// Clear pending priority because the endpoint is done sending results--
	// otherwise, an endpoint with 0 results could delay results returning,
	// because it would never set its maxPendingPriority to 0 in the
	// StreamSearch callback.
	delete(q.endpointMaxPendingPriority, endpoint)
}

// FlushReady sends results that are ready to be sent.
func (q *resultQueue) FlushReady(streamer zoekt.Sender) {
	// we can send any results such that priority > maxPending. Need to
	// calculate maxPending.
	maxPending := math.Inf(-1)
	for _, pri := range q.endpointMaxPendingPriority {
		if pri > maxPending {
			maxPending = pri
		}
	}

	for q.hasResultsToSend(maxPending) {
		streamer.Send(q.pop())
	}
}

// FlushAll will send all results in the queue and any aggregate statistics.
func (q *resultQueue) FlushAll(streamer zoekt.Sender) {
	for q.queue.Len() > 0 {
		streamer.Send(q.pop())
	}

	// We may have had no matches but had stats. Send the final stats if there
	// is any.
	if !q.stats.Zero() {
		streamer.Send(&zoekt.SearchResult{
			Stats: q.stats,
		})
		q.stats = zoekt.Stats{}
	}
}

// pop returns 1 search result from q. The search result contains the current
// aggregate stats. After the call to pop() we reset q's aggregate stats
func (q *resultQueue) pop() *zoekt.SearchResult {
	sr := heap.Pop(&q.queue).(*queueSearchResult)
	q.matchCount -= sr.MatchCount
	q.sizeBytes -= sr.sizeBytes

	// We attach the current aggregate stats to the event and then reset them.
	sr.Stats = q.stats
	q.stats = zoekt.Stats{}

	return sr.SearchResult
}

// hasResultsToSend returns true if there are search results in the queue that
// should be sent up the stream. Retrieve search results by calling pop() on
// resultQueue.
func (q *resultQueue) hasResultsToSend(maxPending float64) bool {
	if q.queue.Len() == 0 {
		return false
	}

	if q.maxQueueDepth >= 0 && q.queue.Len() > q.maxQueueDepth {
		return true
	}

	if q.maxMatchCount >= 0 && q.matchCount > q.maxMatchCount {
		return true
	}

	if q.maxSizeBytes >= 0 && q.sizeBytes > uint64(q.maxSizeBytes) {
		return true
	}

	return q.queue.isTopAbove(maxPending)
}

type queueSearchResult struct {
	*zoekt.SearchResult

	// optimization: It can be expensive to calculate sizeBytes, hence we cache it
	// in the queue.
	sizeBytes uint64
}

// priorityQueue modified from https://golang.org/pkg/container/heap/
// A priorityQueue implements heap.Interface and holds Items.
// All Exported methods are part of the container.heap interface, and
// unexported methods are local helpers.
type priorityQueue []*queueSearchResult

func (pq *priorityQueue) add(sr *queueSearchResult) {
	heap.Push(pq, sr)
}

func (pq *priorityQueue) isTopAbove(limit float64) bool {
	return len(*pq) > 0 && (*pq)[0].Progress.Priority >= limit
}

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Progress.Priority > pq[j].Progress.Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x any) {
	*pq = append(*pq, x.(*queueSearchResult))
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// Search aggregates search over every endpoint in Map.
func (s *HorizontalSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	return AggregateStreamSearch(ctx, s.StreamSearch, q, opts)
}

// AggregateStreamSearch aggregates the stream events into a single batch
// result.
func AggregateStreamSearch(ctx context.Context, streamSearch func(context.Context, query.Q, *zoekt.SearchOptions, zoekt.Sender) error, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	start := time.Now()

	var mu sync.Mutex
	aggregate := &zoekt.SearchResult{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := streamSearch(ctx, q, opts, ZoektStreamFunc(func(event *zoekt.SearchResult) {
		mu.Lock()
		defer mu.Unlock()
		aggregate.Files = append(aggregate.Files, event.Files...)
		aggregate.Stats.Add(event.Stats)
	}))
	if err != nil {
		return nil, err
	}

	aggregate.Duration = time.Since(start)

	return aggregate, nil
}

// List aggregates list over every endpoint in Map.
func (s *HorizontalSearcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	clients, err := s.searchers()
	if err != nil {
		return nil, err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	type result struct {
		rl  *zoekt.RepoList
		err error
	}
	results := make(chan result, len(clients))
	for _, c := range clients {
		go func(c zoekt.Streamer) {
			rl, err := c.List(ctx, q, opts)
			results <- result{rl: rl, err: err}
		}(c)
	}

	// PERF: We don't deduplicate Repos since the only user of List already
	// does deduplication.

	aggregate := zoekt.RepoList{
		Minimal: make(map[uint32]*zoekt.MinimalRepoListEntry),
	}
	for range clients {
		r := <-results
		if r.err != nil {
			if canIgnoreError(ctx, r.err) {
				continue
			}

			return nil, r.err
		}

		aggregate.Repos = append(aggregate.Repos, r.rl.Repos...)
		aggregate.Crashes += r.rl.Crashes
		aggregate.Stats.Add(&r.rl.Stats)

		for k, v := range r.rl.Minimal {
			aggregate.Minimal[k] = v
		}
	}

	return &aggregate, nil
}

// Close will close all connections in Map.
func (s *HorizontalSearcher) Close() {
	s.mu.Lock()
	clients := s.clients
	s.clients = nil
	s.mu.Unlock()
	for _, c := range clients {
		c.Close()
	}
}

func (s *HorizontalSearcher) String() string {
	s.mu.RLock()
	clients := s.clients
	s.mu.RUnlock()
	addrs := make([]string, 0, len(clients))
	for addr := range clients {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)
	return fmt.Sprintf("HorizontalSearcher{%v}", addrs)
}

// searchers returns the list of clients to aggregate over.
func (s *HorizontalSearcher) searchers() (map[string]zoekt.Streamer, error) {
	eps, err := s.Map.Endpoints()
	if err != nil {
		return nil, err
	}

	// Fast-path, check if Endpoints matches addrs. If it does we can use
	// s.clients.
	//
	// We structure our state to optimize for the fast-path.
	s.mu.RLock()
	clients := s.clients
	s.mu.RUnlock()
	if equalKeys(clients, eps) {
		return clients, nil
	}

	// Slow-path, need to remove/connect.
	return s.syncSearchers()
}

// syncSearchers syncs the set of clients with the set of endpoints. It is the
// slow-path of "searchers" since it obtains an write lock on the state before
// proceeding.
func (s *HorizontalSearcher) syncSearchers() (map[string]zoekt.Streamer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check someone didn't beat us to the update
	eps, err := s.Map.Endpoints()
	if err != nil {
		return nil, err
	}

	if equalKeys(s.clients, eps) {
		return s.clients, nil
	}

	set := make(map[string]struct{}, len(eps))
	for _, ep := range eps {
		set[ep] = struct{}{}
	}

	// Disconnect first
	for addr, client := range s.clients {
		if _, ok := set[addr]; !ok {
			client.Close()
		}
	}

	// Use new map to avoid read conflicts
	clients := make(map[string]zoekt.Streamer, len(eps))
	for _, addr := range eps {
		// Try re-use
		client, ok := s.clients[addr]
		if !ok {
			client = s.Dial(addr)
		}
		clients[addr] = client
	}
	s.clients = clients

	return s.clients, nil
}

func equalKeys(a map[string]zoekt.Streamer, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, k := range b {
		if _, ok := a[k]; !ok {
			return false
		}
	}
	return true
}

type dedupper map[string]string // repoName -> endpoint

// Dedup will in-place filter out matches on Repositories we have already
// seen. A Repository has been seen if a previous call to Dedup had a match in
// it with a different endpoint.
func (repoEndpoint dedupper) Dedup(endpoint string, fms []zoekt.FileMatch) []zoekt.FileMatch {
	if len(fms) == 0 { // handles fms being nil
		return fms
	}

	// PERF: Normally fms is sorted by Repository. So we can avoid the map
	// lookup if we just did it for the previous entry.
	lastRepo := ""
	lastSeen := false

	// Remove entries for repos we have already seen.
	dedup := fms[:0]
	for _, fm := range fms {
		if lastRepo == fm.Repository {
			if lastSeen {
				continue
			}
		} else if ep, ok := repoEndpoint[fm.Repository]; ok && ep != endpoint {
			lastRepo = fm.Repository
			lastSeen = true
			continue
		}

		lastRepo = fm.Repository
		lastSeen = false
		dedup = append(dedup, fm)
	}

	// Update seenRepo now, so the next call of dedup will contain the
	// repos.
	lastRepo = ""
	for _, fm := range dedup {
		if lastRepo != fm.Repository {
			lastRepo = fm.Repository
			repoEndpoint[fm.Repository] = endpoint
		}
	}

	return dedup
}

// canIgnoreError returns true if the error we received from zoekt can be
// ignored.
//
// Note: ctx is passed in so we can log to the trace when we ignore an
// error. This is a convenience over logging at the call sites.
//
// Currently the only error we ignore is DNS lookup failures. This is since
// during rollouts of Zoekt, we may still have endpoints of zoekt which are
// not available in our endpoint map. In particular, this happens when using
// Kubernetes and the (default) stateful set watcher.
func canIgnoreError(ctx context.Context, err error) bool {
	reason := canIgnoreErrorReason(err)
	if reason == "" {
		return false
	}

	metricIgnoredError.WithLabelValues(reason).Inc()
	if span := trace.TraceFromContext(ctx); span != nil {
		span.LogFields(
			otlog.String("ignored.reason", reason),
			otlog.String("ignored.error", err.Error()))
	}

	return true
}

func canIgnoreErrorReason(err error) string {
	// Please only add very specific error checks here. An error can be added
	// here if we see it correlated with rollouts on sourcegraph.com.
	// Additionally you should be able to justify why it is related to races
	// between service discovery and us trying to dial.

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return "dns-not-found"
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		if opErr.Timeout() {
			return "dial-timeout"
		}
		// ugly to do this, but is the most robust way. go's net tests do the
		// same check. example:
		//
		//   dial tcp 10.164.51.47:6070: connect: connection refused
		if strings.Contains(opErr.Err.Error(), "connection refused") {
			return "dial-refused"
		}
	}

	return ""
}
