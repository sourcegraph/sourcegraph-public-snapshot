package backend

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

// HorizontalSearcher is a zoekt.Searcher which aggregates searches over
// Map. It manages the connections to Map as the endpoints come and go.
type HorizontalSearcher struct {
	Map  EndpointMap
	Dial func(endpoint string) zoekt.Searcher

	mu      sync.RWMutex
	clients map[string]zoekt.Searcher // addr -> client
}

// Search aggregates search over every endpoint in Map.
func (s *HorizontalSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	start := time.Now()

	clients, err := s.searchers()
	if err != nil {
		return nil, err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	type result struct {
		sr  *zoekt.SearchResult
		err error
	}
	results := make(chan result, len(clients))
	for _, c := range clients {
		go func(c zoekt.Searcher) {
			sr, err := c.Search(ctx, q, opts)
			results <- result{sr: sr, err: err}
		}(c)
	}

	// During rebalancing a repository can appear on more than one replica.
	dedupper := dedupper{}

	var aggregate zoekt.SearchResult
	for range clients {
		r := <-results
		if r.err != nil {
			return nil, r.err
		}

		aggregate.Files = append(aggregate.Files, dedupper.Dedup(r.sr.Files)...)
		aggregate.Stats.Add(r.sr.Stats)
	}

	aggregate.Duration = time.Since(start)

	return &aggregate, nil
}

// List aggregates list over every endpoint in Map.
func (s *HorizontalSearcher) List(ctx context.Context, q query.Q) (*zoekt.RepoList, error) {
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
		go func(c zoekt.Searcher) {
			rl, err := c.List(ctx, q)
			results <- result{rl: rl, err: err}
		}(c)
	}

	// PERF: We don't deduplicate Repos since the only user of List already
	// does deduplication.

	var aggregate zoekt.RepoList
	for range clients {
		r := <-results
		if r.err != nil {
			return nil, r.err
		}

		aggregate.Repos = append(aggregate.Repos, r.rl.Repos...)
		aggregate.Crashes += r.rl.Crashes
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
func (s *HorizontalSearcher) searchers() (map[string]zoekt.Searcher, error) {
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
func (s *HorizontalSearcher) syncSearchers() (map[string]zoekt.Searcher, error) {
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

	// Disconnect first
	for addr, client := range s.clients {
		if _, ok := eps[addr]; !ok {
			client.Close()
		}
	}

	// Use new map to avoid read conflicts
	clients := make(map[string]zoekt.Searcher, len(eps))
	for addr := range eps {
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

func equalKeys(a map[string]zoekt.Searcher, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

type dedupper map[string]struct{}

// Dedup will in-place filter out matches on Repositories we already have
// already seen. A Repository has been seen if a previous call to Dedup had a
// match in it.
func (seenRepo dedupper) Dedup(fms []zoekt.FileMatch) []zoekt.FileMatch {
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
		} else if _, ok := seenRepo[fm.Repository]; ok {
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
			seenRepo[fm.Repository] = struct{}{}
		}
	}

	return dedup
}
