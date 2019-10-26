package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

// AggregateSearcher is a zoekt.Searcher which aggregates searches over
// Map. It manages the connections to Map as the endpoints come and go.
type AggregateSearcher struct {
	Map  EndpointMap
	Dial func(endpoint string) zoekt.Searcher

	mu      sync.RWMutex
	addrs   []string
	clients []zoekt.Searcher
}

// Search aggregates search over every endpoint in Map.
func (s *AggregateSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
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

	var aggregate zoekt.SearchResult
	for range clients {
		r := <-results
		if r.err != nil {
			return nil, r.err
		}

		aggregate.Files = append(aggregate.Files, r.sr.Files...)
		aggregate.Stats.Add(r.sr.Stats)
	}

	aggregate.Duration = time.Since(start)

	return &aggregate, nil
}

// List aggregates list over every endpoint in Map.
func (s *AggregateSearcher) List(ctx context.Context, q query.Q) (*zoekt.RepoList, error) {
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
func (s *AggregateSearcher) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		c.Close()
	}
	s.addrs = nil
	s.clients = nil
}

func (s *AggregateSearcher) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("AggregateSearcher{%v}", s.addrs)
}

// searchers returns the list of clients to aggregate over.
func (s *AggregateSearcher) searchers() ([]zoekt.Searcher, error) {
	eps, err := s.Map.Endpoints()
	if err != nil {
		return nil, err
	}

	// Fast-path, check if Endpoints matches addrs. If it does we can use
	// s.clients.
	//
	// We structure our state to optimize for the fast-path.
	s.mu.RLock()
	addrs, clients := s.addrs, s.clients
	s.mu.RUnlock()
	if equalKeys(addrs, eps) {
		return clients, nil
	}

	// Slow-path, need to remove/connect.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check someone didn't beat us to the update
	eps, err = s.Map.Endpoints()
	if err != nil {
		return nil, err
	}
	if equalKeys(s.addrs, eps) {
		return s.clients, nil
	}

	// Disconnect first
	for i, addr := range s.addrs {
		if _, ok := eps[addr]; !ok {
			s.clients[i].Close()
		}
	}

	// Use new slices to avoid read conflicts
	addrs = []string{}
	clients = []zoekt.Searcher{}
	for addr := range eps {
		// Try re-use
		var client zoekt.Searcher
		for i, a := range s.addrs {
			if a == addr {
				client = s.clients[i]
				break
			}
		}

		if client == nil {
			client = s.Dial(addr)
		}

		addrs = append(addrs, addr)
		clients = append(clients, client)
	}

	s.addrs = addrs
	s.clients = clients

	return s.clients, nil
}

func equalKeys(keys []string, m map[string]struct{}) bool {
	if len(keys) != len(m) {
		return false
	}
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}
