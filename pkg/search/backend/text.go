package backend

import (
	"context"
	"fmt"
	"sync"

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
