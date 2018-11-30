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
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/endpoint"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/search/rpc"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
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

	// DisableCache when true prevents caching of Client.List. Useful in
	// tests.
	DisableCache bool

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

func (c *Zoekt) Search(ctx context.Context, q query.Q, opts *search.Options) (res *search.Result, err error) {
	repos := opts.Repositories
	if len(repos) == 0 {
		return nil, errors.Errorf("repository list empty for indexed text search on %s", q.String())
	}

	zq, err := mapQueryToZoekt(q)
	if err != nil {
		return nil, err
	}

	// Have a top level AND to ensure we only return results for repositories
	// in opts.Repositories
	repoSet := &zoektquery.RepoSet{Set: make(map[string]bool, len(repos))}
	for _, r := range repos {
		repoSet.Set[string(r.Name)] = true
	}
	zq = zoektquery.NewAnd(repoSet, zq)
	zq = zoektquery.Simplify(zq)

	tr, ctx := trace.New(ctx, "zoekt.Search", fmt.Sprintf("%d %+v", len(repoSet.Set), zq.String()))
	defer func() {
		tr.SetError(err)
		if res != nil && len(res.Files) > 0 {
			tr.LazyPrintf("%d file matches", len(res.Files))
		}
		tr.Finish()
	}()

	// If we're only searching a small number of repositories, return more
	// comprehensive results. This is arbitrary.
	defaultMaxSearchResults := 30
	k := 1
	switch {
	case len(repos) <= 500:
		k = 2
	case len(repos) <= 100:
		k = 3
	case len(repos) <= 50:
		k = 5
	case len(repos) <= 25:
		k = 8
	case len(repos) <= 10:
		k = 10
	case len(repos) <= 5:
		k = 100
	}
	if opts.TotalMaxMatchCount > defaultMaxSearchResults {
		k = int(float64(k) * 3 * float64(opts.TotalMaxMatchCount) / float64(defaultMaxSearchResults))
	}

	searchOpts := zoekt.SearchOptions{
		MaxWallTime:            opts.MaxWallTime,
		ShardMaxMatchCount:     100 * k,
		TotalMaxMatchCount:     100 * k,
		ShardMaxImportantMatch: 15 * k,
		TotalMaxImportantMatch: 25 * k,
		MaxDocDisplayCount:     opts.MaxDocDisplayCount,
	}

	// We want zoekt to return more than TotalMaxMatchCount results since we
	// use the extra results to populate reposLimitHit. Additionally the
	// defaults are very low, so we always want to return at least 2000.
	if opts.TotalMaxMatchCount > defaultMaxSearchResults {
		searchOpts.MaxDocDisplayCount = 2 * int(opts.TotalMaxMatchCount)
	}
	if searchOpts.MaxDocDisplayCount < 2000 {
		searchOpts.MaxDocDisplayCount = 2000
	}

	if userProbablyWantsToWaitLonger := opts.TotalMaxMatchCount > defaultMaxSearchResults; userProbablyWantsToWaitLonger {
		searchOpts.MaxWallTime *= time.Duration(3 * float64(opts.TotalMaxMatchCount) / float64(defaultMaxSearchResults))
	}

	resp, err := c.Client.Search(ctx, zq, &searchOpts)
	if err != nil {
		return nil, err
	}

	files := make([]search.FileMatch, len(resp.Files))
	for i, fm := range resp.Files {
		lines := make([]search.LineMatch, 0, len(fm.LineMatches))
		for _, lm := range fm.LineMatches {
			if lm.FileName {
				continue
			}

			frags := make([]search.LineFragmentMatch, len(lm.LineFragments))
			for i, f := range lm.LineFragments {
				frags[i] = search.LineFragmentMatch{
					LineOffset:  f.LineOffset,
					MatchLength: f.MatchLength,
				}
			}

			lines = append(lines, search.LineMatch{
				Line:          lm.Line,
				LineStart:     lm.LineStart,
				LineEnd:       lm.LineEnd,
				LineNumber:    lm.LineNumber,
				LineFragments: frags,
			})
		}

		files[i] = search.FileMatch{
			Path:        fm.FileName,
			Repository:  search.Repository{Name: api.RepoName(fm.Repository)}, // Branch? Safe to case RepoName?
			LineMatches: lines,
		}
	}

	return &search.Result{Files: files}, nil
}

// mapQueryToZoekt translates q to a zoektquery.Q. Additionally we treat any
// ref: clause as false, since we are only allowed to search HEAD.
func mapQueryToZoekt(q query.Q) (zoektquery.Q, error) {
	switch s := q.(type) {

	// Composite types
	case *query.And:
		qs, err := mapQueriesToZoekt(s.Children)
		if err != nil {
			return nil, err
		}
		return zoektquery.NewAnd(qs...), nil
	case *query.Or:
		qs, err := mapQueriesToZoekt(s.Children)
		if err != nil {
			return nil, err
		}
		return zoektquery.NewOr(qs...), nil
	case *query.Not:
		c, err := mapQueryToZoekt(s.Child)
		if err != nil {
			return nil, err
		}
		return &zoektquery.Not{Child: c}, nil
	case *query.Type:
		c, err := mapQueryToZoekt(s.Child)
		if err != nil {
			return nil, err
		}
		return &zoektquery.Type{Child: c}, nil

		// Atoms we can convert
	case *query.Substring:
		return &zoektquery.Substring{
			Pattern:       s.Pattern,
			CaseSensitive: s.CaseSensitive,
			FileName:      s.FileName,
			Content:       s.Content,
		}, nil
	case *query.Regexp:
		return &zoektquery.Regexp{
			Regexp:        s.Regexp,
			FileName:      s.FileName,
			Content:       s.Content,
			CaseSensitive: s.CaseSensitive,
		}, nil
	case *query.RepoSet:
		repoSet := &zoektquery.RepoSet{Set: make(map[string]bool, len(s.Set))}
		for r := range s.Set {
			repoSet.Set[r] = true
		}
		return repoSet, nil

	case *query.Ref:
		// Refs are false since we only index the default branch.
		return &zoektquery.Const{Value: false}, nil

	case *query.Repo:
		// We only want reposets
		return nil, errors.Errorf("zoekt does not allow repo atom: %v", q)

	default:
		return nil, errors.Errorf("unexpected query atom %T: %v", q, q)
	}
}

func mapQueriesToZoekt(qs []query.Q) ([]zoektquery.Q, error) {
	r := make([]zoektquery.Q, len(qs))
	for i, q := range qs {
		var err error
		r[i], err = mapQueryToZoekt(q)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (c *Zoekt) String() string {
	return fmt.Sprintf("zoekt(%v)", c.Client)
}

// ListAll returns the response of List without any restrictions.
func (c *Zoekt) ListAll(ctx context.Context) (*zoekt.RepoList, error) {
	c.mu.Lock()
	r, err := c.listResp, c.listErr
	c.mu.Unlock()

	// No cached responses, start up and just do uncached query.
	if r == nil && err == nil {
		if !c.DisableCache {
			go c.start()
		}
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
