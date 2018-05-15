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
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"time"

	"golang.org/x/net/trace"
	"golang.org/x/sync/semaphore"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

type rankedShard struct {
	zoekt.Searcher
	rank uint16
}

type shardedSearcher struct {
	// Limit the number of parallel queries. Since searching is
	// CPU bound, we can't do better than #CPU queries in
	// parallel.  If we do so, we just create more memory
	// pressure.
	throttle *semaphore.Weighted
	capacity int64

	shards map[string]rankedShard
}

func newShardedSearcher(n int64) *shardedSearcher {
	ss := &shardedSearcher{
		shards:   make(map[string]rankedShard),
		throttle: semaphore.NewWeighted(n),
		capacity: n,
	}
	return ss
}

// NewDirectorySearcher returns a searcher instance that loads all
// shards corresponding to a glob into memory.
func NewDirectorySearcher(dir string) (zoekt.Searcher, error) {
	ss := newShardedSearcher(int64(runtime.NumCPU()))
	tl := &throttledLoader{
		ss:       ss,
		throttle: make(chan struct{}, runtime.NumCPU()),
	}
	_, err := NewDirectoryWatcher(dir, tl)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

// throttledLoader tries to load up to throttle shards in parallel.
type throttledLoader struct {
	ss       *shardedSearcher
	throttle chan struct{}
}

func (tl *throttledLoader) load(key string) {
	tl.throttle <- struct{}{}
	shard, err := loadShard(key)
	<-tl.throttle
	if err != nil {
		log.Printf("reloading: %s, err %v ", key, err)
		return
	}

	tl.ss.replace(key, shard)
}

func (tl *throttledLoader) drop(key string) {
	tl.ss.replace(key, nil)
}

func (ss *shardedSearcher) String() string {
	return "shardedSearcher"
}

// Close closes references to open files. It may be called only once.
func (ss *shardedSearcher) Close() {
	ss.lock(context.Background())
	defer ss.unlock()
	for _, s := range ss.shards {
		s.Close()
	}
}

func (ss *shardedSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (sr *zoekt.SearchResult, err error) {
	tr := trace.New("shardedSearcher.Search", "")
	tr.LazyLog(q, true)
	tr.LazyPrintf("opts: %+v", opts)
	defer func() {
		if sr != nil {
			tr.LazyPrintf("num files: %d", len(sr.Files))
			tr.LazyPrintf("stats: %+v", sr.Stats)
		}
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	start := time.Now()

	aggregate := &zoekt.SearchResult{
		RepoURLs:      map[string]string{},
		LineFragments: map[string]string{},
	}

	// This critical section is large, but we don't want to deal with
	// searches on shards that have just been closed.
	if err := ss.rlock(ctx); err != nil {
		return aggregate, err
	}
	defer ss.runlock()
	tr.LazyPrintf("acquired lock")
	aggregate.Wait = time.Now().Sub(start)
	start = time.Now()

	// TODO - allow for canceling the query.
	shards := ss.getShards()
	all := make(chan shardResult, len(shards))

	var childCtx context.Context
	var cancel context.CancelFunc
	if opts.MaxWallTime == 0 {
		childCtx, cancel = context.WithCancel(ctx)
	} else {
		childCtx, cancel = context.WithTimeout(ctx, opts.MaxWallTime)
	}

	defer cancel()

	// For each query, throttle the number of parallel
	// actions. Since searching is mostly CPU bound, we limit the
	// number of parallel searches. This reduces the peak working
	// set, which hopefully stops https://cs.bazel.build from crashing
	// when looking for the string "com".
	feeder := make(chan zoekt.Searcher, len(shards))
	for _, s := range shards {
		feeder <- s
	}
	close(feeder)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for s := range feeder {
				searchOneShard(childCtx, s, q, opts, all)
			}
		}()
	}

	for range shards {
		r := <-all
		if r.err != nil {
			return nil, r.err
		}
		aggregate.Files = append(aggregate.Files, r.sr.Files...)
		aggregate.Stats.Add(r.sr.Stats)

		if len(r.sr.Files) > 0 {
			for k, v := range r.sr.RepoURLs {
				aggregate.RepoURLs[k] = v
			}
			for k, v := range r.sr.LineFragments {
				aggregate.LineFragments[k] = v
			}
		}

		if cancel != nil && aggregate.Stats.MatchCount > opts.TotalMaxMatchCount {
			cancel()
			cancel = nil
		}
	}

	zoekt.SortFilesByScore(aggregate.Files)
	if max := opts.MaxDocDisplayCount; max > 0 && len(aggregate.Files) > max {
		aggregate.Files = aggregate.Files[:max]
	}
	for i := range aggregate.Files {
		copySlice(&aggregate.Files[i].Content)
		copySlice(&aggregate.Files[i].Checksum)
		for l := range aggregate.Files[i].LineMatches {
			copySlice(&aggregate.Files[i].LineMatches[l].Line)
		}
	}

	aggregate.Duration = time.Now().Sub(start)
	return aggregate, nil
}

func copySlice(src *[]byte) {
	dst := make([]byte, len(*src))
	copy(dst, *src)
	*src = dst
}

type shardResult struct {
	sr  *zoekt.SearchResult
	err error
}

func searchOneShard(ctx context.Context, s zoekt.Searcher, q query.Q, opts *zoekt.SearchOptions, sink chan shardResult) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("crashed shard: %s: %s, %s", s.String(), r, debug.Stack())

			var r zoekt.SearchResult
			r.Stats.Crashes = 1
			sink <- shardResult{&r, nil}
		}
	}()

	ms, err := s.Search(ctx, q, opts)
	sink <- shardResult{ms, err}
}

func (ss *shardedSearcher) List(ctx context.Context, r query.Q) (rl *zoekt.RepoList, err error) {
	tr := trace.New("shardedSearcher.List", "")
	tr.LazyLog(r, true)
	defer func() {
		if rl != nil {
			tr.LazyPrintf("repos size: %d", len(rl.Repos))
			tr.LazyPrintf("crashes: %d", rl.Crashes)
		}
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	type res struct {
		rl  *zoekt.RepoList
		err error
	}

	if err := ss.rlock(ctx); err != nil {
		return nil, err
	}
	defer ss.runlock()
	tr.LazyPrintf("acquired lock")

	shards := ss.getShards()
	shardCount := len(shards)
	all := make(chan res, shardCount)
	tr.LazyPrintf("shardCount: %d", len(shards))

	for _, s := range shards {
		go func(s zoekt.Searcher) {
			defer func() {
				if r := recover(); r != nil {
					all <- res{
						&zoekt.RepoList{Crashes: 1}, nil,
					}
				}
			}()
			ms, err := s.List(ctx, r)
			all <- res{ms, err}
		}(s.Searcher)
	}

	crashes := 0
	uniq := map[string]*zoekt.RepoListEntry{}

	for i := 0; i < shardCount; i++ {
		r := <-all
		if r.err != nil {
			return nil, r.err
		}
		crashes += r.rl.Crashes
		for _, r := range r.rl.Repos {
			prev, ok := uniq[r.Repository.Name]
			if !ok {
				cp := *r
				uniq[r.Repository.Name] = &cp
			} else {
				prev.Stats.Add(&r.Stats)
			}
		}
	}

	aggregate := make([]*zoekt.RepoListEntry, 0, len(uniq))
	for _, v := range uniq {
		aggregate = append(aggregate, v)
	}
	return &zoekt.RepoList{
		Repos:   aggregate,
		Crashes: crashes,
	}, nil
}

func (s *shardedSearcher) rlock(ctx context.Context) error {
	return s.throttle.Acquire(ctx, 1)
}

// getShards returns the currently loaded shards. The shards must be
// accessed under a rlock call. The shards are sorted by decreasing
// rank.
func (s *shardedSearcher) getShards() []rankedShard {
	var res []rankedShard
	for _, sh := range s.shards {
		res = append(res, sh)
	}
	// TODO: precompute this.
	sort.Slice(res, func(i, j int) bool {
		return res[i].rank > res[j].rank
	})
	return res
}

func (s *shardedSearcher) runlock() {
	s.throttle.Release(1)
}

func (s *shardedSearcher) lock(ctx context.Context) error {
	return s.throttle.Acquire(ctx, s.capacity)
}

func (s *shardedSearcher) unlock() {
	s.throttle.Release(s.capacity)
}

func shardRank(s zoekt.Searcher) uint16 {
	q := query.Repo{}
	result, err := s.List(context.Background(), &q)
	if err != nil {
		return 0
	}
	if len(result.Repos) == 0 {
		return 0
	}
	return result.Repos[0].Repository.Rank
}

func (s *shardedSearcher) replace(key string, shard zoekt.Searcher) {
	s.lock(context.Background())
	defer s.unlock()
	old := s.shards[key]
	if old.Searcher != nil {
		old.Close()
	}

	if shard == nil {
		delete(s.shards, key)
	} else {

		s.shards[key] = rankedShard{
			rank:     shardRank(shard),
			Searcher: shard,
		}
	}
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
