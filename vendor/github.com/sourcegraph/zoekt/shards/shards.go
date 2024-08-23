// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shards

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/atomic"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
	"github.com/sourcegraph/zoekt/trace"
)

var (
	metricShardsLoaded = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_shards_loaded",
		Help: "The number of shards currently loaded",
	})
	metricShardsLoadedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_shards_loaded_total",
		Help: "The total number of shards loaded",
	})
	metricShardsLoadFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_shards_load_failed_total",
		Help: "The total number of shard loads that failed",
	})

	metricSearchRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_search_running",
		Help: "The number of concurrent search requests running",
	})
	metricSearchShardRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_search_shard_running",
		Help: "The number of concurrent search requests in a shard running",
	})
	metricSearchFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_failed_total",
		Help: "The total number of search requests that failed",
	})
	metricSearchDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "zoekt_search_duration_seconds",
		Help:    "The duration a search request took in seconds",
		Buckets: prometheus.DefBuckets, // DefBuckets good for service timings
	})

	// A Counter per Stat. Name should match field in zoekt.Stats.
	metricSearchContentBytesLoadedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_content_loaded_bytes_total",
		Help: "Total amount of I/O for reading contents",
	})
	metricSearchIndexBytesLoadedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_index_loaded_bytes_total",
		Help: "Total amount of I/O for reading from index",
	})
	metricSearchCrashesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_crashes_total",
		Help: "Total number of search shards that had a crash",
	})
	metricSearchFileCountTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_file_count_total",
		Help: "Total number of files containing a match",
	})
	metricSearchShardFilesConsideredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_shard_files_considered_total",
		Help: "Total number of files in shards that we considered",
	})
	metricSearchFilesConsideredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_files_considered_total",
		Help: "Total files that we evaluated. Equivalent to files for which all atom matches (including negations) evaluated to true",
	})
	metricSearchFilesLoadedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_files_loaded_total",
		Help: "Total files for which we loaded file content to verify substring matches",
	})
	metricSearchFilesSkippedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_files_skipped_total",
		Help: "Total candidate files whose contents weren't examined because we gathered enough matches",
	})
	metricSearchShardsSkippedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_shards_skipped_total",
		Help: "Total shards that we did not process because a query was canceled",
	})
	metricSearchMatchCountTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_match_count_total",
		Help: "Total number of non-overlapping matches",
	})
	metricSearchNgramMatchesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_ngram_matches_total",
		Help: "Total number of candidate matches as a result of searching ngrams",
	})
	metricSearchNgramLookupsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_ngram_lookups_total",
		Help: "Total number of times we accessed an ngram in the index",
	})
	metricSearchRegexpsConsideredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "zoekt_search_regexps_considered_total",
		Help: "Total number of times regexp was called on files that we evaluated",
	})

	metricListRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_running",
		Help: "The number of concurrent list requests running",
	})
	metricListShardRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_shard_running",
		Help: "The number of concurrent list requests in a shard running",
	})
	metricShardsBatchReplaceDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "zoekt_shards_batch_replace_duration_seconds",
		Help:    "The time it takes to replace a batch of Searchers.",
		Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
	})
	metricListAllRepos = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_repos",
		Help: "The last List(true) value for RepoStats.Repos. Repos is used for aggregrating the number of repositories.",
	})
	metricListAllShards = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_shards",
		Help: "The last List(true) value for RepoStats.Shards. Shards is the total number of search shards.",
	})
	metricListAllDocuments = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_documents",
		Help: "The last List(true) value for RepoStats.Documents. Documents holds the number of documents or files.",
	})
	metricListAllIndexBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_index_bytes",
		Help: "The last List(true) value for RepoStats.IndexBytes. IndexBytes is the amount of RAM used for index overhead.",
	})
	metricListAllContentBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_content_bytes",
		Help: "The last List(true) value for RepoStats.ContentBytes. ContentBytes is the amount of RAM used for raw content.",
	})
	metricListAllNewLinesCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_new_lines_count",
		Help: "The last List(true) value for RepoStats.NewLinesCount.",
	})
	metricListAllDefaultBranchNewLinesCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_default_branch_new_lines_count",
		Help: "The last List(true) value for RepoStats.DefaultBranchNewLinesCount.",
	})
	metricListAllOtherBranchesNewLinesCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "zoekt_list_all_stats_other_branches_new_lines_count",
		Help: "The last List(true) value for RepoStats.OtherBranchesNewLinesCount.",
	})
)

type rankedShard struct {
	zoekt.Searcher

	priority float64 // maximum priority across all repos in the shard

	// We have out of band ranking on compound shards which can change even if
	// the shard file does not. So we compute a rank in getShards. We store
	// repos here to avoid the cost of List in the search request path.
	repos []*zoekt.Repository
}

// loaded stores the state we compute when updating the state of shards from
// disk.
type loaded struct {
	// shards is the currently loaded shards sorted by decreasing rank and
	// should not be mutated.
	shards []*rankedShard

	// ready is true if sharded searcher has finished loading all initial
	// shards on startup.
	ready bool
}

type shardedSearcher struct {
	// Limit the number of parallel queries. Since searching is
	// CPU bound, we can't do better than #CPU queries in
	// parallel.  If we do so, we just create more memory
	// pressure.
	sched scheduler

	mu     sync.Mutex // protects writes to shards
	shards map[string]*rankedShard

	ready  atomic.Bool
	ranked atomic.Value
}

func newShardedSearcher(n int64) *shardedSearcher {
	ss := &shardedSearcher{
		shards: make(map[string]*rankedShard),
		sched:  newScheduler(n),
	}
	return ss
}

// NewDirectorySearcher returns a searcher instance that loads all
// shards corresponding to a glob into memory.
func NewDirectorySearcher(dir string) (zoekt.Streamer, error) {
	return newDirectorySearcher(dir, true)
}

// NewDirectorySearcherFast is like NewDirectorySearcher, but does not block
// on the initial loading of shards.
//
// This exists since in the case of zoekt-webserver we are happy with having
// partial availability since that is better than no availability on large
// instances.
func NewDirectorySearcherFast(dir string) (zoekt.Streamer, error) {
	return newDirectorySearcher(dir, false)
}

func newDirectorySearcher(dir string, waitUntilReady bool) (zoekt.Streamer, error) {
	ss := newShardedSearcher(int64(runtime.GOMAXPROCS(0)))
	tl := &loader{
		ss: ss,
	}
	dw, err := newDirectoryWatcher(dir, tl)
	if err != nil {
		return nil, err
	}

	if waitUntilReady {
		if err := dw.WaitUntilReady(); err != nil {
			return nil, err
		}
	}

	ds := &directorySearcher{
		Streamer:         ss,
		directoryWatcher: dw,
	}

	return &typeRepoSearcher{Streamer: ds}, nil
}

type directorySearcher struct {
	zoekt.Streamer

	directoryWatcher *DirectoryWatcher
}

func (s *directorySearcher) Close() {
	// We need to Stop directoryWatcher first since it calls load/unload on
	// Searcher.
	s.directoryWatcher.Stop()
	s.Streamer.Close()
}

type loader struct {
	ss *shardedSearcher
}

func (tl *loader) load(keys ...string) {
	// This is called with all keys on startup, so once this function has
	// finished running shardedSearcher will be ready.
	defer tl.ss.markReady()

	if len(keys) == 0 {
		// If there's nothing to load, we exit early here, but we want to mark
		// ourselves as ready.
		return
	}

	var (
		mu           sync.Mutex     // synchronizes writes to the shards map
		wg           sync.WaitGroup // used to wait for all shards to load
		sem          = semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
		loadedShards = make(map[string]zoekt.Searcher)
	)

	publishLoaded := func() {
		mu.Lock()
		chunk := loadedShards
		loadedShards = make(map[string]zoekt.Searcher)
		mu.Unlock()
		tl.ss.replace(chunk)
	}

	log.Printf("loading %d shard(s): %s", len(keys), humanTruncateList(keys, 5))

	lastProgress := time.Now()
	for i, key := range keys {
		// If taking a while to start-up occasionally give a progress message
		if time.Since(lastProgress) > 5*time.Second {
			log.Printf("still need to load %d shards...", len(keys)-i)
			lastProgress = time.Now()

			publishLoaded()
		}

		_ = sem.Acquire(context.Background(), 1)
		wg.Add(1)

		go func(key string) {
			defer sem.Release(1)
			defer wg.Done()

			shard, err := loadShard(key)
			if err != nil {
				metricShardsLoadFailedTotal.Inc()
				log.Printf("reloading: %s, err %v ", key, err)
				return
			}
			metricShardsLoadedTotal.Inc()

			mu.Lock()
			loadedShards[key] = shard
			mu.Unlock()
		}(key)
	}

	wg.Wait()

	publishLoaded()
}

func (tl *loader) drop(keys ...string) {
	shards := make(map[string]zoekt.Searcher, len(keys))
	for _, key := range keys {
		shards[key] = nil
	}
	tl.ss.replace(shards)
}

func (ss *shardedSearcher) String() string {
	return "shardedSearcher"
}

// Close closes references to open files. It may be called only once.
func (ss *shardedSearcher) Close() {
	ss.mu.Lock()
	shards := make(map[string]zoekt.Searcher, len(ss.shards))
	for k := range ss.shards {
		shards[k] = nil
	}
	ss.mu.Unlock()

	ss.replace(shards)
}

func selectRepoSet(shards []*rankedShard, q query.Q) ([]*rankedShard, query.Q) {
	and, ok := q.(*query.And)
	if ok {
		return doSelectRepoSet(shards, and)
	}

	// We have queries which look like (reposet ...) and we want to do the same
	// optimizations. To simplify we just always wrap the query in And and then
	// on the return value call Simplify to unwrap. In particular this is
	// important for List calls.
	and = &query.And{Children: []query.Q{q}}
	shards, q = doSelectRepoSet(shards, and)
	return shards, query.Simplify(q)
}

func doSelectRepoSet(shards []*rankedShard, and *query.And) ([]*rankedShard, query.Q) {
	// (and (reposet ...) (q))
	// (and true (q)) with a filtered shards
	// (and false) // noop

	// (and (repobranches ...) (q))
	// (and (repobranches ...) (q))

	// Note: we also support (and (repo ...) (q)) even though sourcegraph does
	// not generate those sorts of queries. This is to support manual testing.

	hasReposForPredicate := func(pred func(repo *zoekt.Repository) bool) func(repos []*zoekt.Repository) (any, all bool) {
		return func(repos []*zoekt.Repository) (any, all bool) {
			any = false
			all = true
			for _, repo := range repos {
				b := pred(repo)
				any = any || b
				all = all && b
			}
			return any, all
		}
	}

	for i, c := range and.Children {
		var setSize int
		var hasRepos func([]*zoekt.Repository) (bool, bool)
		switch setQuery := c.(type) {
		case *query.RepoSet:
			setSize = len(setQuery.Set)
			hasRepos = hasReposForPredicate(func(repo *zoekt.Repository) bool {
				return setQuery.Set[repo.Name]
			})
		case *query.RepoIDs:
			setSize = int(setQuery.Repos.GetCardinality())
			hasRepos = hasReposForPredicate(func(repo *zoekt.Repository) bool {
				return setQuery.Repos.Contains(repo.ID)
			})
		case *query.Repo:
			setSize = 0
			hasRepos = hasReposForPredicate(func(repo *zoekt.Repository) bool {
				return setQuery.Regexp.MatchString(repo.Name)
			})
		case *query.BranchesRepos:
			for _, br := range setQuery.List {
				setSize += int(br.Repos.GetCardinality())
			}

			hasRepos = hasReposForPredicate(func(repo *zoekt.Repository) bool {
				for _, br := range setQuery.List {
					if br.Repos.Contains(repo.ID) {
						return true
					}
				}
				return false
			})
		default:
			continue
		}

		// setSize may be larger than the number of shards we have. The size of
		// filtered is bounded by min(len(set), len(shards))
		if setSize > len(shards) {
			setSize = len(shards)
		}

		filtered := make([]*rankedShard, 0, setSize)
		filteredAll := true

		for _, s := range shards {
			if any, all := hasRepos(s.repos); any {
				filtered = append(filtered, s)
				filteredAll = filteredAll && all
			}
		}

		// We don't need to adjust the query since we are returning an empty set
		// of shards to search.
		if len(filtered) == 0 {
			return filtered, and
		}

		// We can't simplify the query since we are searching shards which contain
		// repos we aren't supposed to search.
		if !filteredAll {
			return filtered, and
		}

		// We don't want to mutate the original and, so we clone it before
		// mutating it.
		and = &query.And{Children: slices.Clone(and.Children)}

		// This optimization allows us to avoid the work done by
		// indexData.simplify for each shard.
		//
		// For example if our query is (and (reposet foo bar) (content baz))
		// then at this point filtered is [foo bar] and q is the same. For each
		// shard indexData.simplify will simplify to (and true (content baz)) ->
		// (content baz). This work can be done now once, rather than per shard.
		switch c := c.(type) {
		case *query.RepoSet, *query.RepoIDs, *query.Repo:
			and.Children[i] = &query.Const{Value: true}
			return filtered, query.Simplify(and)

		case *query.BranchesRepos:
			// We can only replace if all the repos want the same branches. We
			// simplify and just check that we are requesting 1 branch. The common
			// case is just asking for HEAD, so this should be effective.
			if len(c.List) != 1 {
				return filtered, and
			}

			// Every repo wants the same branches, so we can replace RepoBranches
			// with a list of branch queries.
			and.Children[i] = &query.Branch{Pattern: c.List[0].Branch, Exact: true}
			return filtered, query.Simplify(and)
		}

		// Stop after first RepoSet, otherwise we might append duplicate
		// shards to `filtered`
		return filtered, and
	}

	return shards, and
}

func (ss *shardedSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (sr *zoekt.SearchResult, err error) {
	tr, ctx := trace.New(ctx, "shardedSearcher.Search", "")
	defer func() {
		tr.Finish()
	}()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	collectSender := newCollectSender(opts)

	start := time.Now()
	proc, err := ss.sched.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer proc.Release()
	tr.LazyPrintf("acquired process")

	wait := time.Since(start)
	start = time.Now()

	loaded := ss.getLoaded()
	done, err := streamSearch(ctx, proc, q, opts, loaded.shards, collectSender)
	defer done()
	if err != nil {
		return nil, err
	}

	aggregate, ok := collectSender.Done()
	if !ok {
		aggregate = &zoekt.SearchResult{
			RepoURLs:      map[string]string{},
			LineFragments: map[string]string{},
		}
	}

	copyFiles(aggregate)

	if !loaded.ready {
		// We may have missed results due to not being fully loaded.
		aggregate.Stats.Crashes++
	}

	aggregate.Stats.Wait = wait
	aggregate.Stats.Duration = time.Since(start)

	return aggregate, nil
}

func (ss *shardedSearcher) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, sender zoekt.Sender) (err error) {
	tr, ctx := trace.New(ctx, "shardedSearcher.StreamSearch", "")
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError(err)
		}
		tr.Finish()
	}()

	start := time.Now()
	proc, err := ss.sched.Acquire(ctx)
	if err != nil {
		return err
	}
	defer proc.Release()
	tr.LazyPrintf("acquired process")

	loaded := ss.getLoaded()
	shards := loaded.shards

	maxPendingPriority := math.Inf(-1)
	if len(shards) > 0 {
		maxPendingPriority = shards[0].priority
	}

	stillLoadingCrashes := 0
	if !loaded.ready {
		// We may have missed results due to not being fully loaded.
		stillLoadingCrashes++
	}

	sender.Send(&zoekt.SearchResult{
		Stats: zoekt.Stats{
			Crashes: stillLoadingCrashes,
			Wait:    time.Since(start),
		},
		Progress: zoekt.Progress{
			MaxPendingPriority: maxPendingPriority,
		},
	})

	// Matches flow from the shards up the stack in the following order:
	//
	// 1. Search shards
	// 2. flushCollectSender (aggregate)
	// 3. limitSender (limit)
	// 4. copyFileSender (copy)
	//
	// For streaming, the wrapping has to happen in the inverted order.
	sender = copyFileSender(sender)

	if truncator, hasLimits := zoekt.NewDisplayTruncator(opts); hasLimits {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		sender = limitSender(cancel, sender, truncator)
	}

	sender, flush := newFlushCollectSender(opts, sender)

	done, err := streamSearch(ctx, proc, q, opts, shards, sender)

	// Even though streaming is done, we may have results sitting in a buffer we
	// need to flush. So we need to send those before calling done.
	flush()
	done()

	return err
}

// streamSearch is an internal helper since both Search and StreamSearch are
// largely similar.
//
// done must always be called, even if err is non-nil. The SearchResults sent
// via sender contain references to the underlying mmap data that the garbage
// collector can't see. Calling done informs the garbage collector it is free
// to collect those shards. The caller must call copyFiles on any
// SearchResults it returns/streams out before calling done.
func streamSearch(ctx context.Context, proc *process, q query.Q, opts *zoekt.SearchOptions, shards []*rankedShard, sender zoekt.Sender) (done func(), err error) {
	tr, ctx := trace.New(ctx, "shardedSearcher.streamSearch", "")
	overallStart := time.Now()
	metricSearchRunning.Inc()
	defer func() {
		metricSearchRunning.Dec()
		metricSearchDuration.Observe(time.Since(overallStart).Seconds())
		if err != nil {
			metricSearchFailedTotal.Inc()

			tr.LazyPrintf("error: %v", err)
			tr.SetError(err)
		}
		tr.Finish()
	}()

	// Select the subset of shards that we will search over for the given query.
	{
		beforeLen := len(shards)
		beforeQ := q
		shards, q = selectRepoSet(shards, q)
		tr.LazyPrintf("selectRepoSet shards=%d->%d q=%s->%s", beforeLen, len(shards), beforeQ, q)
	}

	if len(shards) == 0 {
		return func() {}, nil
	}

	var cancel context.CancelFunc
	if opts.MaxWallTime == 0 {
		ctx, cancel = context.WithCancel(ctx)
	} else {
		ctx, cancel = context.WithTimeout(ctx, opts.MaxWallTime)
	}

	defer cancel()

	// We set the number of workers to GOMAXPROCS, or the number of shards,
	// whichever is smaller.
	workers := runtime.GOMAXPROCS(0)
	if workers > len(shards) {
		workers = len(shards)
	}

	type result struct {
		priority float64
		*zoekt.SearchResult
		err error
	}

	var (
		// buffered channels to continue searching when sending back results
		// takes a while / blocks. The maximum pending result set is workers * 2.
		results = make(chan *result, workers)
		search  = make(chan *rankedShard, workers)
		wg      sync.WaitGroup
	)

	// Start workers that receive shards from the search channel, search them,
	// and send the results to the results channel. This process is repeated
	// until the search channel is closed.
	//
	// Note: Making "search" a buffered channel has the effect of limiting the number of parallel shard searches.
	// Since searching is mostly CPU bound, limiting parallel shard searches also reduces the peak working set.
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for s := range search {
				sr, err := searchOneShard(ctx, s, q, opts)
				r := &result{priority: s.priority, SearchResult: sr, err: err}
				results <- r
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var (
		pending = make(prioritySlice, 0, workers)
		shard   = 0
		next    = shards[shard]

		// We need a separate nil-able reference to the same channel so we can close(search) for the worker
		// go-routines to finish but also set work to nil in order for the select statement below to ignore
		// that case when we want to stop a search. This is needed because sending on a closed channel panics.
		work = search
	)

	stop := func() {
		if work != nil {
			close(search)
			work = nil
			next = nil
		}
	}

	// tracked so we can stop when we hit TotalMaxMatchCount
	var totalMatchCount int

search:
	for {
		// At the top of each iteration, have the proc associated with this search yield its won "timeslice"
		// to possibly allow other searches to make progress
		_ = proc.Yield(ctx) // Note: we let searchOneShard handle context errors

		select {
		case work <- next: // is there a worker available to search the next shard?
			pending.append(next.priority)

			shard++
			if shard == len(shards) {
				stop()
			} else {
				next = shards[shard]
			}
		case r, ok := <-results: // is there a result to send back?
			if !ok {
				break search
			}

			// delete this result's priority from pending before computing the new max pending priority
			pending.remove(r.priority)

			if r.err != nil {
				// Set final error and stop searching new shards, but consume any pending
				// search results.
				stop()
				err = r.err
				continue
			}

			// Update the match count statistics and stop searching new shards if we've
			// reached the limit set in the options.
			totalMatchCount += r.SearchResult.Stats.MatchCount
			if opts.TotalMaxMatchCount > 0 && totalMatchCount > opts.TotalMaxMatchCount {
				stop()
			}

			observeMetrics(r.SearchResult)

			r.Priority = r.priority
			r.MaxPendingPriority = pending.max()

			sendByRepository(r.SearchResult, opts, sender) // send the result back to the client
		}
	}

	return func() { runtime.KeepAlive(shards) }, err
}

// sendByRepository splits a zoekt.SearchResult by repository and calls
// sender.Send for each batch. Ranking in Sourcegraph expects zoekt.SearchResult
// to contain results with the same zoekt.SearchResult.Priority only.
//
// We split by repository instead of by priority because it is easier to set
// RepoURLs and LineFragments in zoekt.SearchResult.
func sendByRepository(result *zoekt.SearchResult, opts *zoekt.SearchOptions, sender zoekt.Sender) {
	if len(result.RepoURLs) <= 1 || len(result.Files) == 0 {
		zoekt.SortFiles(result.Files)
		sender.Send(result)
		return
	}

	send := func(repoName string, a, b int, stats zoekt.Stats) {
		zoekt.SortFiles(result.Files[a:b])
		sender.Send(&zoekt.SearchResult{
			Stats: stats,
			Progress: zoekt.Progress{
				Priority:           result.Files[a].RepositoryPriority,
				MaxPendingPriority: result.MaxPendingPriority,
			},
			Files:         result.Files[a:b],
			RepoURLs:      map[string]string{repoName: result.RepoURLs[repoName]},
			LineFragments: map[string]string{repoName: result.LineFragments[repoName]},
		})
	}

	var startIndex, endIndex int
	curRepoID := result.Files[0].RepositoryID
	curRepoName := result.Files[0].Repository

	fm := zoekt.FileMatch{}
	for endIndex, fm = range result.Files {
		if curRepoID != fm.RepositoryID {
			// Stats must stay aggregate-able, hence we sent the aggregate stats with the
			// last event.
			send(curRepoName, startIndex, endIndex, zoekt.Stats{})

			startIndex = endIndex
			curRepoID = fm.RepositoryID
			curRepoName = fm.Repository
		}
	}

	send(curRepoName, startIndex, endIndex+1, result.Stats)
}

func observeMetrics(sr *zoekt.SearchResult) {
	metricSearchContentBytesLoadedTotal.Add(float64(sr.Stats.ContentBytesLoaded))
	metricSearchIndexBytesLoadedTotal.Add(float64(sr.Stats.IndexBytesLoaded))
	metricSearchCrashesTotal.Add(float64(sr.Stats.Crashes))
	metricSearchFileCountTotal.Add(float64(sr.Stats.FileCount))
	metricSearchShardFilesConsideredTotal.Add(float64(sr.Stats.ShardFilesConsidered))
	metricSearchFilesConsideredTotal.Add(float64(sr.Stats.FilesConsidered))
	metricSearchFilesLoadedTotal.Add(float64(sr.Stats.FilesLoaded))
	metricSearchFilesSkippedTotal.Add(float64(sr.Stats.FilesSkipped))
	metricSearchShardsSkippedTotal.Add(float64(sr.Stats.ShardsSkipped))
	metricSearchMatchCountTotal.Add(float64(sr.Stats.MatchCount))
	metricSearchNgramMatchesTotal.Add(float64(sr.Stats.NgramMatches))
	metricSearchNgramLookupsTotal.Add(float64(sr.Stats.NgramLookups))
	metricSearchRegexpsConsideredTotal.Add(float64(sr.Stats.RegexpsConsidered))
}

func copySlice(src *[]byte) {
	if *src == nil {
		return
	}
	dst := make([]byte, len(*src))
	copy(dst, *src)
	*src = dst
}

func copyFiles(sr *zoekt.SearchResult) {
	for i := range sr.Files {
		copySlice(&sr.Files[i].Content)
		copySlice(&sr.Files[i].Checksum)
		for l := range sr.Files[i].LineMatches {
			copySlice(&sr.Files[i].LineMatches[l].Line)
			copySlice(&sr.Files[i].LineMatches[l].Before)
			copySlice(&sr.Files[i].LineMatches[l].After)
		}
		for c := range sr.Files[i].ChunkMatches {
			copySlice(&sr.Files[i].ChunkMatches[c].Content)
		}
	}
}

func searchOneShard(ctx context.Context, s zoekt.Searcher, q query.Q, opts *zoekt.SearchOptions) (sr *zoekt.SearchResult, err error) {
	metricSearchShardRunning.Inc()
	defer func() {
		metricSearchShardRunning.Dec()
		if e := recover(); e != nil {
			log.Printf("crashed shard: %s: %#v, %s", s, e, debug.Stack())

			if sr == nil {
				sr = &zoekt.SearchResult{}
			}
			sr.Stats.Crashes = 1
		}
	}()

	return s.Search(ctx, q, opts)
}

type shardListResult struct {
	rl  *zoekt.RepoList
	err error
}

func listOneShard(ctx context.Context, s zoekt.Searcher, q query.Q, opts *zoekt.ListOptions, sink chan shardListResult) {
	metricListShardRunning.Inc()
	defer func() {
		metricListShardRunning.Dec()
		if r := recover(); r != nil {
			log.Printf("crashed shard: %s: %s, %s", s.String(), r, debug.Stack())
			sink <- shardListResult{
				&zoekt.RepoList{Crashes: 1}, nil,
			}
		}
	}()

	ms, err := s.List(ctx, q, opts)
	sink <- shardListResult{ms, err}
}

func (ss *shardedSearcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (rl *zoekt.RepoList, err error) {
	tr, ctx := trace.New(ctx, "shardedSearcher.List", "")
	metricListRunning.Inc()
	defer func() {
		metricListRunning.Dec()
		if rl != nil {
			tr.LazyPrintf("repos.size=%d reposmap.size=%d crashes=%d", len(rl.Repos), len(rl.ReposMap), rl.Crashes)
		}
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError(err)
		}
		tr.Finish()
	}()

	q = query.Simplify(q)
	isAll := false
	if c, ok := q.(*query.Const); ok {
		isAll = c.Value
	}

	proc, err := ss.sched.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer proc.Release()
	tr.LazyPrintf("acquired process")

	loaded := ss.getLoaded()
	shards := loaded.shards

	// Setup what we return now, since we may short circuit if there are no
	// shards to search.
	stillLoadingCrashes := 0
	if !loaded.ready {
		// We may have missed results due to not being fully loaded.
		stillLoadingCrashes++
	}
	agg := zoekt.RepoList{
		Crashes:  stillLoadingCrashes,
		ReposMap: zoekt.ReposMap{},
		Repos:    []*zoekt.RepoListEntry{},
	}

	// PERF: Select the subset of shards that we will search over for the given
	// query. A common List query only asks for a specific repo, so this is an
	// important optimization.
	{
		beforeLen := len(shards)
		beforeQ := q
		shards, q = selectRepoSet(shards, q)
		tr.LazyPrintf("selectRepoSet shards=%d->%d q=%s->%s", beforeLen, len(shards), beforeQ, q)
	}

	if len(shards) == 0 {
		return &agg, nil
	}

	shardCount := len(shards)
	all := make(chan shardListResult, shardCount)
	feeder := make(chan zoekt.Searcher, len(shards))
	for _, s := range shards {
		feeder <- s
	}
	close(feeder)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for s := range feeder {
				listOneShard(ctx, s, q, opts, all)
			}
		}()
	}

	uniq := map[string]*zoekt.RepoListEntry{}

	for range shards {
		r := <-all
		if r.err != nil {
			return nil, r.err
		}

		agg.Crashes += r.rl.Crashes
		agg.Stats.Add(&r.rl.Stats)

		for _, r := range r.rl.Repos {
			prev, ok := uniq[r.Repository.Name]
			if !ok {
				cp := *r // We need to copy because we mutate r.Stats when merging duplicates
				uniq[r.Repository.Name] = &cp
			} else {
				prev.Stats.Add(&r.Stats)
			}
		}

		for id, r := range r.rl.ReposMap {
			_, ok := agg.ReposMap[id]
			if !ok {
				agg.ReposMap[id] = r
			}
		}
	}

	agg.Repos = make([]*zoekt.RepoListEntry, 0, len(uniq))
	for _, r := range uniq {
		agg.Repos = append(agg.Repos, r)
	}

	// Only one of these fields is populated and in all cases the size of that
	// field is the number of Repos.
	//
	// Note: we don't just add individual Stats.Repos since a repository can
	// have multiple shards.
	agg.Stats.Repos = len(uniq) + len(agg.ReposMap)

	if isAll && len(agg.Repos) > 0 {
		reportListAllMetrics(agg.Repos)
	}

	return &agg, nil
}

func reportListAllMetrics(repos []*zoekt.RepoListEntry) {
	var stats zoekt.RepoStats
	for _, r := range repos {
		stats.Add(&r.Stats)
	}

	metricListAllRepos.Set(float64(stats.Repos))
	metricListAllIndexBytes.Set(float64(stats.IndexBytes))
	metricListAllContentBytes.Set(float64(stats.ContentBytes))
	metricListAllDocuments.Set(float64(stats.Documents))
	metricListAllShards.Set(float64(stats.Shards))
	metricListAllNewLinesCount.Set(float64(stats.NewLinesCount))
	metricListAllDefaultBranchNewLinesCount.Set(float64(stats.DefaultBranchNewLinesCount))
	metricListAllOtherBranchesNewLinesCount.Set(float64(stats.OtherBranchesNewLinesCount))
}

// getLoaded returns the currently loaded shards. Shared so do not mutate.
func (s *shardedSearcher) getLoaded() loaded {
	// next commit will store the true value of this, for now we keep the
	// backwards compatible behaviour.
	ready := s.ready.Load()
	// ranked is loaded after ready to avoid a race were ready is true but
	// ranked is still not the final set of shards.
	ranked, _ := s.ranked.Load().([]*rankedShard)
	return loaded{
		shards: ranked,
		ready:  ready,
	}
}

func mkRankedShard(s zoekt.Searcher) *rankedShard {
	q := query.Const{Value: true}
	result, err := s.List(context.Background(), &q, nil)
	if err != nil {
		return &rankedShard{Searcher: s}
	}
	if len(result.Repos) == 0 {
		return &rankedShard{Searcher: s}
	}

	var (
		maxPriority float64
		repos       = make([]*zoekt.Repository, 0, len(result.Repos))
	)
	for i := range result.Repos {
		repo := &result.Repos[i].Repository
		repos = append(repos, repo)
		if repo.RawConfig != nil {
			priority, _ := strconv.ParseFloat(repo.RawConfig["priority"], 64)
			if priority > maxPriority {
				maxPriority = priority
			}
		}
	}

	return &rankedShard{
		Searcher: s,
		repos:    repos,
		priority: maxPriority,
	}
}

// markReady should be called once all shards have been passed into replace on
// startup. Once s is marked as ready it stops reporting a Crash in the
// response Stats.
func (s *shardedSearcher) markReady() {
	s.ready.CompareAndSwap(false, true)
}

func (s *shardedSearcher) replace(shards map[string]zoekt.Searcher) {
	if len(shards) == 0 {
		return
	}

	defer func(began time.Time) {
		metricShardsBatchReplaceDurationSeconds.Observe(time.Since(began).Seconds())
	}(time.Now())

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, shard := range shards {
		var r *rankedShard
		if shard != nil {
			r = mkRankedShard(shard)
		}

		old := s.shards[key]
		if shard == nil {
			delete(s.shards, key)
		} else {
			s.shards[key] = r
		}

		if old != nil && old.Searcher != nil {
			//                 _ ___                /^^\ /^\  /^^\_
			//     _          _@)@) \            ,,/ '` ~ `'~~ ', `\.
			//   _/o\_ _ _ _/~`.`...'~\        ./~~..,'`','',.,' '  ~:
			//  / `,'.~,~.~  .   , . , ~|,   ,/ .,' , ,. .. ,,.   `,  ~\_
			// ( ' _' _ '_` _  '  .    , `\_/ .' ..' '  `  `   `..  `,   \_
			//  ~V~ V~ V~ V~ ~\ `   ' .  '    , ' .,.,''`.,.''`.,.``. ',   \_
			//   _/\ /\ /\ /\_/, . ' ,   `_/~\_ .' .,. ,, , _/~\_ `. `. '.,  \_
			//  < ~ ~ '~`'~'`, .,  .   `_: ::: \_ '      `_/ ::: \_ `.,' . ',  \_
			//   \ ' `_  '`_    _    ',/ _::_::_ \ _    _/ _::_::_ \   `.,'.,`., \-,-,-,_,_,
			//    `'~~ `'~~ `'~~ `'~~  \(_)(_)(_)/  `~~' \(_)(_)(_)/ ~'`\_.._,._,'_;_;_;_;_;
			//
			// We can't just call Close now, because there may be ongoing searches
			// which have old in the shards list. Previously we used an exclusive
			// lock to guarantee there were no concurrent searches. However, that
			// led to blocking on the read path.
			//
			// We could introduce granular locking per rankedShard to know when
			// there are no more references. However, this becomes tricky in
			// practice. Instead we rely on the garbage collector noticing old is no
			// longer used. We take care in our searchers to runtime.KeepAlive until
			// we have stopped referencing the underling mmap data.
			runtime.SetFinalizer(old, func(r *rankedShard) {
				r.Close()
			})
		}
	}

	ranked := make([]*rankedShard, 0, len(s.shards))
	for _, r := range s.shards {
		ranked = append(ranked, r)
	}

	sort.Slice(ranked, func(i, j int) bool {
		priorityDiff := ranked[i].priority - ranked[j].priority
		if priorityDiff != 0 {
			return priorityDiff > 0
		}
		if len(ranked[i].repos) == 0 || len(ranked[j].repos) == 0 {
			// Protect against empty names which can happen if we fail to List or
			// the shard is full of tombstones. Prefer the shard which has names.
			return len(ranked[i].repos) >= len(ranked[j].repos)
		}
		return ranked[i].repos[0].Name < ranked[j].repos[0].Name
	})

	s.ranked.Store(ranked)

	metricShardsLoaded.Set(float64(len(ranked)))
}

func loadShard(fn string) (zoekt.Searcher, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	iFile, err := zoekt.NewIndexFile(f)
	if err != nil {
		return nil, err
	}
	s, err := zoekt.NewSearcher(iFile)
	if err != nil {
		iFile.Close()
		return nil, fmt.Errorf("NewSearcher(%s): %v", fn, err)
	}

	return s, nil
}

// prioritySlice is a trivial implementation of an array that provides three
// things: appending a value, removing a value, and getting the array's max.
// Operations take O(n) time, which is acceptable because N is restricted to
// GOMAXPROCS (i.e., number of cpu cores) by the shardedSearcher interface.
type prioritySlice []float64

func (p *prioritySlice) append(pri float64) {
	*p = append(*p, pri)
}

func (p *prioritySlice) remove(pri float64) {
	for i, opri := range *p {
		if opri == pri {
			if i != len(*p)-1 {
				// swap to make this element the tail
				(*p)[i] = (*p)[len(*p)-1]
			}
			// pop the end off
			*p = (*p)[:len(*p)-1]
			break
		}
	}
}

func (p *prioritySlice) max() float64 {
	// remove() and max() could be combined, but this is easier to read and
	// the expected performance difference from the extra lock and loop is
	// almost certainly irrelevant.
	maxPri := math.Inf(-1)
	for _, pri := range *p {
		if pri > maxPri {
			maxPri = pri
		}
	}
	return maxPri
}
