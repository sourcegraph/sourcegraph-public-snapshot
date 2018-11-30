package backend

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/endpoint"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/search/rpc"
)

// TextJIT is a client for searching our just in time text search (the
// searcher service).
//
// It implements search.Searcher
type TextJIT struct {
	Endpoints *endpoint.Map

	mu      sync.Mutex
	clients map[string]search.Searcher
}

// Search distributes the search across the searcher replicas, and merges the
// results. opts.Repositories is required to be non-empty.
func (t *TextJIT) Search(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error) {
	if len(opts.Repositories) == 0 {
		return nil, errors.Errorf("repository list empty for text search on %s", q.String())
	}

	all := &search.Result{}

	// TODO parallize, delete missing endpoints, respect MaxWallTime
	origOpts := opts
	for _, r := range origOpts.Repositories {
		opts := *origOpts
		opts.Repositories = []search.Repository{r}

		client, err := t.client(r)
		if err != nil {
			return nil, err
		}

		result, err := client.Search(ctx, q, &opts)
		if err != nil {
			return all, err
		}

		all.Files = append(all.Files, result.Files...)
	}

	return all, nil
}

func (t *TextJIT) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range t.clients {
		c.Close()
	}
	t.clients = make(map[string]search.Searcher)
}

func (t *TextJIT) String() string {
	return fmt.Sprintf("textjit(%v)", t.Endpoints)
}

func (t *TextJIT) client(r search.Repository) (search.Searcher, error) {
	addr, err := t.Endpoints.Get(r.String(), nil)
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	client, ok := t.clients[addr]
	if !ok {
		// Creating a client is non-blocking so can hold lock.
		client = rpc.Client(addr)
		t.clients[addr] = client
	}
	t.mu.Unlock()

	// TODO if we add a new client, check if we need to remove any.
	return client, nil
}

// Zoekt is Searcher which wraps a zoekt.Searcher.
//
// Note: Zoekt starts up background goroutines, so call Close when done using
// the Client.
type Zoekt struct {
	Client zoekt.Searcher

	mu       sync.Mutex
	state    int32 // 0 not running, 1 running, 2 stopped
	listResp *zoekt.RepoList
	listErr  error
}

// Close will tear down the background goroutines.
func (c *Zoekt) Close() {
	c.mu.Lock()
	c.state = 2
	c.mu.Unlock()
}

// ListAll returns the response of List without any restrictions.
func (c *Zoekt) ListAll(ctx context.Context) (*zoekt.RepoList, error) {
	c.mu.Lock()
	r, err := c.listResp, c.listErr
	c.mu.Unlock()

	// No cached responses, start up and just do uncached query.
	if r == nil && err == nil {
		go c.start()
		r, err = c.Client.List(ctx, &zoektquery.Const{Value: true})
	}

	return r, err
}

func (c *Zoekt) start() {
	c.mu.Lock()
	if c.state != 0 {
		// already running or stopped
		c.mu.Unlock()
		return
	}
	c.state = 1 // mark running
	c.mu.Unlock()

	errorCount := 0
	state := int32(1)
	for state == 1 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		listResp, listErr := c.Client.List(ctx, &zoektquery.Const{Value: true})
		cancel()

		// Only update on error once it has happened 3 times in a row. This is
		// to prevent us caching transient errors, and instead fallback on the
		// old list.
		update := true
		if listErr != nil {
			errorCount++
			if errorCount <= 3 {
				update = false
			}
		} else {
			errorCount = 0
		}

		c.mu.Lock()
		state = c.state
		if update {
			c.listResp, c.listErr = listResp, listErr
		}
		c.mu.Unlock()

		randSleep(5*time.Second, 2*time.Second)
	}
}

// randSleep will sleep for an expected d duration with a jitter in [-jitter /
// 2, jitter / 2].
func randSleep(d, jitter time.Duration) {
	delta := time.Duration(rand.Int63n(int64(jitter))) - (jitter / 2)
	time.Sleep(d + delta)
}
