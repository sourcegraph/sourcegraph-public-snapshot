package backend

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/endpoint"
	sgzoekt "github.com/sourcegraph/sourcegraph/pkg/search/zoekt"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/query"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/rpc"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Text is a searcher which routes requests for indexed commits to indexed
// search, but fallsback to our just in time search for everything else.
type Text struct {
	// Index is the indexed searcher (Zoekt). It needs to return the list of
	// repositories it can search. This should be Zoekt, but is an interface
	// for testing purposes.
	Index interface {
		sgzoekt.Searcher
		SplitRepositories(context.Context, query.Q, *sgzoekt.Options) (canSearch, cantSearch []api.RepoName, err error)
	}

	// Fallback is the searcher used for anything that Index can't
	// search. This should be TextJIT, but is an interface for testing
	// purposes.
	Fallback sgzoekt.Searcher
}

func (t *Text) Search(ctx context.Context, q query.Q, opts *sgzoekt.Options) (res *sgzoekt.Result, err error) {
	if len(opts.Repositories) == 0 {
		return nil, errors.Errorf("repository list empty for text search on %s", q.String())
	}

	tr, ctx := trace.New(ctx, "Text.Search", fmt.Sprintf("query: %v, numRepoRevs: %d", q, len(opts.Repositories)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	isIndexAvailable := true
	shards := make(chan shard)
	go func() {
		defer close(shards)

		index, fallback, err := t.Index.SplitRepositories(ctx, q, opts)
		if err != nil {
			// Don't hard fail if index is not available yet.
			tr.LogFields(otlog.String("indexErr", err.Error()))
			isIndexAvailable = false
			err = nil
			index = nil
			fallback = opts.Repositories
		}

		if len(index) > 0 {
			o := *opts
			o.Repositories = index
			shards <- shard{Searcher: t.Index, Q: q, Options: &o}
		}

		if len(fallback) > 0 {
			o := *opts
			o.Repositories = fallback
			shards <- shard{Searcher: t.Fallback, Q: q, Options: &o}
		}
	}()

	res, err = shardedSearch(ctx, shards)
	if res != nil && !isIndexAvailable {
		res.Stats.Unavailable = append(res.Stats.Unavailable, SourceZoekt)
	}
	return res, err
}

func (t *Text) Close() {}

func (t *Text) String() string {
	return fmt.Sprintf("text(index=%v fallback=%v)", t.Index, t.Fallback)
}

// SourceSearcher is the source name used by searcher, our textjit service.
const SourceSearcher = sgzoekt.Source("textjit")

// TextJIT is a client for searching our just in time text search (the
// searcher service).
//
// It implements sgzoekt.Searcher
type TextJIT struct {
	Endpoints *endpoint.Map
	Resolve   func(ctx context.Context, name api.RepoName, spec string) (api.CommitID, error)

	mu      sync.Mutex
	clients map[string]sgzoekt.Searcher
}

// Search distributes the search across the searcher replicas, and merges the
// results. opts.Repositories is required to be non-empty.
func (t *TextJIT) Search(ctx context.Context, q query.Q, opts *sgzoekt.Options) (*sgzoekt.Result, error) {
	repos, err := expandRepoRefs(q, opts.Repositories)
	if err != nil {
		return nil, err
	}

	var cancel context.CancelFunc
	if opts.MaxWallTime > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.MaxWallTime)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	sem, err := t.semaphore()
	if err != nil {
		return nil, err
	}

	resC := make(chan searchResponse)
	go func() {
		defer close(resC)

		var wg sync.WaitGroup
		defer wg.Wait()

		origOpts := opts
		for _, r := range repos {
			if err := sem.Acquire(ctx); err != nil {
				return
			}

			wg.Add(1)
			go func(r sgzoekt.Repository) {
				defer sem.Release()
				defer wg.Done()

				commit, err := t.Resolve(ctx, r.Name, r.RefPattern)
				if err != nil {
					if s, err := handleError(SourceSearcher, r, err); err != nil {
						resC <- searchResponse{error: err}
					} else {
						var result sgzoekt.Result
						result.Stats.Status = append(result.Stats.Status, *s)
						resC <- searchResponse{Result: &result}
					}
					return
				}
				r.Commit = commit
				opts := *origOpts
				opts.Repositories = []api.RepoName{r.Name}

				qSearcher, err := expandForRepoAtCommit(q, r)
				if err != nil {
					resC <- searchResponse{error: err}
					return
				}

				client, err := t.client(r)
				if err != nil {
					resC <- searchResponse{error: err}
					return
				}

				result, err := client.Search(ctx, qSearcher, &opts)
				if err != nil {
					resC <- searchResponse{error: err}
					return
				}

				// Searcher doesn't know the repo.RefPattern used, so we set it.
				for i := range result.Stats.Status {
					result.Stats.Status[i].Repository = r
				}
				for i := range result.Files {
					result.Files[i].Repository = r
				}

				resC <- searchResponse{Result: result}
			}(r)
		}
	}()

	all := &sgzoekt.Result{}
	for r := range resC {
		if r.error != nil {
			// Drain resC
			cancel()
			for range resC {
			}
			return nil, r.error
		}
		all.Add(r.Result)
	}

	return all, nil
}

// semaphore returns a semaphore for limiting search concurrency.
func (t *TextJIT) semaphore() (semaphore, error) {
	// We don't want to overload searcher instances. So we use the heuristic
	// of 5 searchers per replica.
	eps, err := t.Endpoints.Endpoints()
	if err != nil {
		return nil, err
	}

	// Additionally we use this opportunity to check if we need to do a GC
	t.maybeGC(len(eps))

	return make(semaphore, len(eps)*5), nil
}

func (t *TextJIT) maybeGC(expectedClientCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.clients) == expectedClientCount {
		return
	}

	eps, err := t.Endpoints.Endpoints()
	if err != nil {
		log15.Warn("TextJIT.maybeGC unexpected error listing endpoints", "error", err)
		return
	}
	for k, c := range t.clients {
		if _, ok := eps[k]; !ok {
			c.Close()
			delete(eps, k)
		}
	}
}

func expandForRepoAtCommit(q query.Q, repo sgzoekt.Repository) (query.Q, error) {
	var err error
	q = query.Map(q, func(q query.Q) query.Q {
		switch s := q.(type) {
		case *query.Repo:
			err = errors.Errorf("text search expected repo atom to be expanded: %v", q)
			return q
		case *query.RepoSet:
			_, ok := s.Set[string(repo.Name)]
			return &query.Const{Value: ok}
		case *query.Ref:
			return &query.Const{Value: s.Pattern == repo.RefPattern}
		default:
			return q
		}
	}, nil)
	if err != nil {
		return nil, err
	}
	// Include the commit as the only ref
	q = query.NewAnd(&query.Ref{Pattern: string(repo.Commit)}, q)
	return query.Simplify(q), nil
}

func (t *TextJIT) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, c := range t.clients {
		c.Close()
	}
	t.clients = make(map[string]sgzoekt.Searcher)
}

func (t *TextJIT) String() string {
	return fmt.Sprintf("textjit(%v)", t.Endpoints)
}

func (t *TextJIT) client(r sgzoekt.Repository) (sgzoekt.Searcher, error) {
	ep, err := t.Endpoints.Get(r.String(), nil)
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	if t.clients == nil {
		t.clients = make(map[string]sgzoekt.Searcher)
	}
	client, ok := t.clients[ep]
	if !ok {
		// Creating a client is non-blocking so can hold lock.
		addr := strings.TrimPrefix(ep, "http://")
		addr = strings.TrimPrefix(addr, "https://")
		client = rpc.Client(addr)
		t.clients[ep] = client
	}
	t.mu.Unlock()

	// TODO if we add a new client, check if we need to remove any.
	return client, nil
}

func expandRepoRefs(q query.Q, repoNames []api.RepoName) ([]sgzoekt.Repository, error) {
	// Implementation Note: If a query does not have a "ref:" pattern, it
	// implies that the user wants to search the default branch. This gives us
	// a problem. For example this query:
	//
	//   (ref:x repo:a) or repo:b
	//
	// A user would expect this to match a@x and b@master. However, given that
	// expression b actually matches any ref. The issue here is "or" doesn't
	// restrict scope, since it just unions results. So to simplify our
	// algorithm for determining refs for a repo, we only assume a repo is on
	// the default branch if it matches no ref clauses.

	// Create a set of refs to search
	refs := map[string]struct{}{}
	query.VisitAtoms(q, func(q query.Q) {
		if s, ok := q.(*query.Ref); ok {
			refs[s.Pattern] = struct{}{}
		}
	})
	if len(refs) == 0 {
		repos := make([]sgzoekt.Repository, len(repoNames))
		for i := range repoNames {
			repos[i] = sgzoekt.Repository{Name: repoNames[i]}
		}
		return repos, nil
	}

	// Otherwise we need to check if the repository can match with a ref
	// specifier
	var evalErr error
	include := func(name api.RepoName, refPattern string) bool {
		v, ok := query.EvalConstant(q, func(q query.Q) (bool, bool) {
			switch s := q.(type) {
			case *query.Repo:
				evalErr = errors.Errorf("text search expected repo atom to be expanded: %v", q)
				return false, false
			case *query.RepoSet:
				_, ok := s.Set[string(name)]
				return ok, true
			case *query.Ref:
				return s.Pattern == refPattern, true
			default:
				return false, false
			}
		})
		return !ok || v
	}

	var repos []sgzoekt.Repository
	for _, name := range repoNames {
		for p := range refs {
			if include(name, p) {
				repos = append(repos, sgzoekt.Repository{
					Name:       name,
					RefPattern: p,
				})
			}
		}
	}
	if evalErr != nil {
		return nil, evalErr
	}

	return repos, nil
}

// SourceZoekt is the source name used by Zoekt.
const SourceZoekt = sgzoekt.Source("textindexed")

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
	disabled bool
}

// Close will tear down the background goroutines.
func (c *Zoekt) Close() {
	c.mu.Lock()
	c.state = 2
	c.mu.Unlock()
}

// Search implements Searcher.Search
func (c *Zoekt) Search(ctx context.Context, q query.Q, opts *sgzoekt.Options) (res *sgzoekt.Result, err error) {
	if !c.Enabled() {
		return nil, errors.New("indexed search is disabled")
	}

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
	for _, name := range repos {
		repoSet.Set[string(name)] = true
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

	searchOpts := mapOptionsToZoekt(opts)
	tr.LazyPrintf("options: %+v", &searchOpts)

	start := time.Now()
	resp, err := c.Client.Search(ctx, zq, searchOpts)
	if err != nil {
		return nil, err
	}

	// We don't have info on which shards are skipped / timedout. So if we
	// skipped any files, we conservatily mark every repository as having
	// skipped files.
	status := sgzoekt.RepositoryStatusSearched
	if resp.Stats.FilesSkipped+resp.Stats.ShardsSkipped > 0 {
		status = sgzoekt.RepositoryStatusLimitHit
		if time.Since(start) >= searchOpts.MaxWallTime {
			status = sgzoekt.RepositoryStatusTimedOut
		}
	}
	statuses := make([]sgzoekt.RepositoryStatus, len(repos))
	for i, name := range repos {
		statuses[i] = sgzoekt.RepositoryStatus{
			Repository: sgzoekt.Repository{Name: name},
			Source:     SourceZoekt,
			Status:     status,
		}
	}

	return &sgzoekt.Result{
		Stats: sgzoekt.Stats{
			// NOTE: This is different to TextJIT. Zoekt MatchCount is the
			// number of non-overlapping matches.
			MatchCount: resp.Stats.MatchCount,
			Status:     statuses,
		},
		Files: mapZoektFileMatch(resp.Files),
	}, nil
}

// mapOptionsToZoekt translates our search options into Zoekts.
func mapOptionsToZoekt(opts *sgzoekt.Options) *zoekt.SearchOptions {
	repos := opts.Repositories

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

	searchOpts := &zoekt.SearchOptions{
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

	return searchOpts
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
	case *query.Const:
		return q, nil

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

func mapZoektFileMatch(zf []zoekt.FileMatch) []sgzoekt.FileMatch {
	files := make([]sgzoekt.FileMatch, len(zf))
	for i, fm := range zf {
		lines := make([]sgzoekt.LineMatch, 0, len(fm.LineMatches))
		for _, lm := range fm.LineMatches {
			if lm.FileName {
				continue
			}

			frags := make([]sgzoekt.LineFragmentMatch, len(lm.LineFragments))
			for i, f := range lm.LineFragments {
				frags[i] = sgzoekt.LineFragmentMatch{
					LineOffset:  f.LineOffset,
					MatchLength: f.MatchLength,
				}
			}

			lines = append(lines, sgzoekt.LineMatch{
				Line:          lm.Line,
				LineNumber:    lm.LineNumber,
				LineFragments: frags,
			})
		}

		files[i] = sgzoekt.FileMatch{
			Path:        fm.FileName,
			Repository:  sgzoekt.Repository{Name: api.RepoName(fm.Repository)}, // Branch? Safe to case RepoName?
			LineMatches: lines,
		}
	}
	return files
}

func (c *Zoekt) String() string {
	return fmt.Sprintf("zoekt(%v)", c.Client)
}

// SplitRepositories splits repos into two lists: indexed contains
// repositories that can be searched by zoekt, unindexed contains everything
// else.
func (c *Zoekt) SplitRepositories(ctx context.Context, q query.Q, opts *sgzoekt.Options) (indexed, unindexed []api.RepoName, err error) {
	// We only use zoekt if we have no ref atoms. See notes about the master
	// branch where we calculate which commits to search for a repository in
	// TextJIT
	hasRef := false
	query.VisitAtoms(q, func(q query.Q) {
		if _, ok := q.(*query.Ref); ok {
			hasRef = true
		}
	})
	if hasRef {
		return nil, opts.Repositories, nil
	}

	// Otherwise we want to search all repositories via Zoekt, but we need
	// prune out the ones that aren't indexed. Make a copy of
	// opts.Repositories since we are going to mutate it.
	indexed = append(indexed, opts.Repositories...)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	resp, err := c.ListAll(ctx)
	if err != nil {
		return nil, opts.Repositories, err
	}

	// Everything currently in indexed is at HEAD. Filter out repos which
	// zoekt hasn't indexed yet.
	set := map[string]struct{}{}
	for _, repo := range resp.Repos {
		set[repo.Repository.Name] = struct{}{}
	}
	unindexedSet := map[api.RepoName]struct{}{}
	for _, name := range unindexed {
		unindexedSet[name] = struct{}{}
	}
	head := indexed
	indexed = indexed[:0]
	for _, name := range head {
		if _, ok := set[string(name)]; ok {
			indexed = append(indexed, name)
		} else if _, ok = unindexedSet[name]; !ok {
			unindexed = append(unindexed, name)
		}
	}

	return indexed, unindexed, nil
}

// ListAll returns the response of List without any restrictions.
func (c *Zoekt) ListAll(ctx context.Context) (*zoekt.RepoList, error) {
	if !c.Enabled() {
		// By returning an empty list Text.Search won't send any queries to
		// Zoekt.
		return &zoekt.RepoList{}, nil
	}

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

// SetEnabled will disable zoekt if b is false.
func (c *Zoekt) SetEnabled(b bool) {
	c.mu.Lock()
	c.disabled = !b
	c.mu.Unlock()
}

// Enabled returns true if Zoekt is enabled. It is enabled if Client is
// non-nil and it hasn't been disabled by SetEnable.
func (c *Zoekt) Enabled() bool {
	c.mu.Lock()
	b := c.disabled
	c.mu.Unlock()
	return c.Client != nil && !b
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
		if !c.Enabled() {
			// If we haven't been stopped, reset state so start() is called
			// again when we are enabled. We can defer unlocking since we will
			// return.
			c.mu.Lock()
			defer c.mu.Unlock()
			if c.state == 1 {
				c.state = 0
				c.listResp, c.listErr = nil, nil
			}
			return
		}

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
