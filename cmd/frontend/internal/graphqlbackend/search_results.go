package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/trace"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// searchResultsCommon contains fields that should be returned by all funcs
// that contribute to the overall search result set.
type searchResultsCommon struct {
	limitHit bool     // whether the limit on results was hit
	cloning  []string // repos that could not be searched because they were still being cloned
	missing  []string // repos that could not be searched because they do not exist

	// timedout usually contains repos that haven't finished being fetched yet.
	// This should only happen for large repos and the searcher caches are
	// purged.
	timedout []string
}

func (c *searchResultsCommon) LimitHit() bool {
	return c.limitHit
}

func (c *searchResultsCommon) Cloning() []string {
	if c.cloning == nil {
		return []string{}
	}
	return c.cloning
}

func (c *searchResultsCommon) Missing() []string {
	if c.missing == nil {
		return []string{}
	}
	return c.missing
}

func (c *searchResultsCommon) Timedout() []string {
	if c.timedout == nil {
		return []string{}
	}
	return c.timedout
}

// update updates c with the other data, deduping as necessary. It modifies c but
// does not modify other.
func (c *searchResultsCommon) update(other searchResultsCommon) {
	c.limitHit = c.limitHit || other.limitHit

	appendUnique := func(dst *[]string, src []string) {
		dstSet := make(map[string]struct{}, len(*dst))
		for _, s := range *dst {
			dstSet[s] = struct{}{}
		}
		for _, s := range src {
			if _, present := dstSet[s]; !present {
				*dst = append(*dst, s)
			}
		}
	}
	appendUnique(&c.cloning, other.cloning)
	appendUnique(&c.missing, other.missing)
	appendUnique(&c.timedout, other.timedout)
}

type searchResults struct {
	queryForFileMatches bool
	combinedQuery       string
	results             []*searchResult
	searchResultsCommon
	alert *searchAlert
}

func (sr *searchResults) Results() []*searchResult {
	return sr.results
}

func (sr *searchResults) ResultCount() int32 {
	return int32(len(sr.results))
}

func (sr *searchResults) ApproximateResultCount() string {
	if sr.alert != nil {
		return "?"
	}
	if sr.limitHit || len(sr.missing) > 0 || len(sr.cloning) > 0 {
		return fmt.Sprintf("%d+", len(sr.results))
	}
	return strconv.Itoa(len(sr.results))
}

func (sr *searchResults) Alert() *searchAlert { return sr.alert }

// blameFileMatchCache caches Repos.GetByURI, Repos.ResolveRev, and RepoVCS.Open
// operations.
type blameFileMatchCache struct {
	cachedReposMu sync.RWMutex
	cachedRepos   map[string]*sourcegraph.Repo

	cachedRevsMu sync.RWMutex
	cachedRevs   map[string]*sourcegraph.ResolvedRev

	cachedVCSReposMu sync.RWMutex
	cachedVCSRepos   map[string]vcs.Repository
}

// repoVCSOpen is like localstore.Repos.GetByURI except it is cached by b.
func (b *blameFileMatchCache) reposGetByURI(ctx context.Context, repoURI string) (*sourcegraph.Repo, error) {
	b.cachedReposMu.RLock()
	repo, ok := b.cachedRepos[repoURI]
	b.cachedReposMu.RUnlock()
	if ok {
		return repo, nil
	}
	repo, err := db.Repos.GetByURI(ctx, repoURI)
	if err != nil {
		return nil, err
	}
	b.cachedReposMu.Lock()
	b.cachedRepos[repoURI] = repo
	b.cachedReposMu.Unlock()
	return repo, nil
}

// repoVCSOpen is like localstore.Repos.ResolveRev except it is cached by b.
func (b *blameFileMatchCache) reposResolveRev(ctx context.Context, repoID int32, revStr string) (*sourcegraph.ResolvedRev, error) {
	cacheKey := fmt.Sprint(repoID, revStr)
	b.cachedRevsMu.RLock()
	rev, ok := b.cachedRevs[cacheKey]
	b.cachedRevsMu.RUnlock()
	if ok {
		return rev, nil
	}
	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: repoID,
		Rev:  revStr,
	})
	if err != nil {
		return nil, err
	}
	b.cachedRevsMu.Lock()
	b.cachedRevs[cacheKey] = rev
	b.cachedRevsMu.Unlock()
	return rev, nil
}

// repoVCSOpen is like localstore.RepoVCS.Open except it is cached by b.
func (b *blameFileMatchCache) repoVCSOpen(ctx context.Context, repoID int32) (vcs.Repository, error) {
	b.cachedVCSReposMu.RLock()
	vcsrepo, ok := b.cachedVCSRepos[fmt.Sprint(repoID)]
	b.cachedVCSReposMu.RUnlock()
	if ok {
		return vcsrepo, nil
	}
	vcsrepo, err := db.RepoVCS.Open(ctx, repoID)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	b.cachedVCSReposMu.Lock()
	b.cachedVCSRepos[fmt.Sprint(repoID)] = vcsrepo
	b.cachedVCSReposMu.Unlock()
	return vcsrepo, nil
}

// blameFileMatch blames the specified file match to produce the time at which
// the first line match inside of it was authored.
func (sr *searchResults) blameFileMatch(ctx context.Context, fm *fileMatch, cache *blameFileMatchCache) (t time.Time, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "blameFileMatch")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	u, err := url.Parse(fm.uri)
	if err != nil {
		return time.Time{}, err
	}
	repoURI := u.Host + u.Path
	revStr := u.RawQuery

	repo, err := cache.reposGetByURI(ctx, repoURI)
	if err != nil {
		return time.Time{}, err
	}

	rev, err := cache.reposResolveRev(ctx, repo.ID, revStr)
	if err != nil {
		return time.Time{}, err
	}

	vcsrepo, err := cache.repoVCSOpen(ctx, repo.ID)
	if err != nil {
		return time.Time{}, err
	}

	// Blame the first line match.
	lm := fm.LineMatches()[0]
	hunks, err := vcsrepo.BlameFile(ctx, u.Fragment, &vcs.BlameOptions{
		NewestCommit: vcs.CommitID(rev.CommitID),
		StartLine:    int(lm.LineNumber()),
		EndLine:      int(lm.LineNumber()),
	})
	if err != nil {
		return time.Time{}, err
	}

	return hunks[0].Author.Date, nil
}

var (
	sparklineFileCache        = rcache.NewWithTTL("sparkline_file", 3600) // 1h
	sparklineFileCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "graphql",
		Name:      "sparkline_file_cache_hit",
		Help:      "Counts cache hits and misses for SearchResults.Sparkline file calculations.",
	}, []string{"type"})

	sparklineGenericCache        = rcache.NewWithTTL("sparkline_generic", 5*60) // 5m
	sparklineGenericCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "graphql",
		Name:      "sparkline_generic_cache_hit",
		Help:      "Counts cache hits and misses for SearchResults.Sparkline generic calculations.",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(sparklineFileCacheCounter)
	prometheus.MustRegister(sparklineGenericCacheCounter)
}

func (sr *searchResults) Sparkline(ctx context.Context) (sparkline []int32, err error) {
	if sr.combinedQuery == "" {
		panic("internal search error (expected combined query to be present)")
	}
	cache := sparklineGenericCache
	counter := sparklineGenericCacheCounter
	if sr.queryForFileMatches {
		// The query is for file matches. These are slower for calculating the
		// sparkline information (requires blame), so we use a different cache
		// with higher TTL.
		cache = sparklineFileCache
		counter = sparklineFileCacheCounter
	}

	// Check if value is in the cache.
	jsonRes, ok := cache.Get(sr.combinedQuery)
	if ok {
		counter.WithLabelValues("hit").Inc()
		if err := json.Unmarshal(jsonRes, &sparkline); err != nil {
			return nil, err
		}
		return sparkline, nil
	}

	// Calculate value from scratch.
	counter.WithLabelValues("miss").Inc()
	sparkline, err = sr.doCalculateSparkline(ctx)
	if err != nil {
		return nil, err
	}

	// Store value in the cache.
	jsonRes, err = json.Marshal(sparkline)
	if err != nil {
		return nil, err
	}
	cache.Set(sr.combinedQuery, jsonRes)
	return sparkline, nil
}

func (sr *searchResults) doCalculateSparkline(ctx context.Context) (sparkline []int32, err error) {
	var (
		days     = 30                 // number of days the sparkline represents
		maxBlame = 100                // maximum number of file results to blame for date/time information.
		run      = parallel.NewRun(8) // number of concurrent blame ops
	)

	var (
		sparklineMu sync.Mutex
		blameOps    = 0
		cache       = &blameFileMatchCache{
			cachedRepos:    map[string]*sourcegraph.Repo{},
			cachedRevs:     map[string]*sourcegraph.ResolvedRev{},
			cachedVCSRepos: map[string]vcs.Repository{},
		}
	)
	sparkline = make([]int32, days)
	addPoint := func(t time.Time) {
		// Check if the author date of the search result is inside of our sparkline
		// timerange.
		now := time.Now()
		if t.Before(now.Add(-time.Duration(len(sparkline)) * 24 * time.Hour)) {
			// Outside the range of the sparkline.
			return
		}
		sparklineMu.Lock()
		defer sparklineMu.Unlock()
		for n := range sparkline {
			d1 := now.Add(-time.Duration(n) * 24 * time.Hour)
			d2 := now.Add(-time.Duration(n-1) * 24 * time.Hour)
			if t.After(d1) && t.Before(d2) {
				sparkline[n]++ // on the nth day
			}
		}
	}

	// Consider all of our search results as a potential data point in our
	// sparkline.
loop:
	for _, r := range sr.results {
		switch {
		case r.diff != nil:
			// Diff searches are cheap, because we implicitly have author date info.
			addPoint(r.diff.commit.author.date)
		case r.fileMatch != nil:
			// File match searches are more expensive, because we must blame the
			// (first) line in order to know its placement in our sparkline.
			blameOps++
			if blameOps > maxBlame {
				// We have exceeded our budget of blame operations for
				// calculating this sparkline, so don't do any more file match
				// blaming.
				continue loop
			}

			run.Acquire()
			go func(r *searchResult) {
				defer func() {
					if r := recover(); r != nil {
						run.Error(fmt.Errorf("recover: %v", r))
					}
					run.Release()
				}()

				// Blame the file match in order to retrieve date informatino.
				var err error
				t, err := sr.blameFileMatch(ctx, r.fileMatch, cache)
				if err != nil {
					log15.Warn("failed to blame fileMatch during sparkline generation", "error", err)
					return
				}
				addPoint(t)
			}(r)
		default:
			panic("SearchResults.Sparkline unexpected union type state")
		}
	}
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("blame_ops", blameOps)
	return sparkline, nil
}

func (r *searchResolver) Results(ctx context.Context) (*searchResults, error) {
	return r.doResults(ctx, "")
}

func (r *searchResolver) doResults(ctx context.Context, forceOnlyResultType string) (res *searchResults, err error) {
	tr := trace.New("graphql.searchResults", fmt.Sprintf("%s %s", r.args.Query, r.args.ScopeQuery))
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	repos, missingRepoRevs, _, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}
	tr.LazyPrintf("searching %d repos, %d missing", len(repos), len(missingRepoRevs))
	if len(repos) == 0 {
		alert, err := r.alertForNoResolvedRepos(ctx)
		if err != nil {
			return nil, err
		}
		return &searchResults{alert: alert}, nil
	}
	if overLimit {
		alert, err := r.alertForOverRepoLimit(ctx)
		if err != nil {
			return nil, err
		}
		return &searchResults{alert: alert}, nil
	}

	var patternsToCombine []string
	for _, v := range r.combinedQuery.Values(searchquery.FieldDefault) {
		// Treat quoted strings as literal strings to match, not regexps.
		var pattern string
		switch {
		case v.String != nil:
			pattern = regexp.QuoteMeta(*v.String)
		case v.Regexp != nil:
			pattern = v.Regexp.String()
		}
		if pattern == "" {
			continue
		}
		patternsToCombine = append(patternsToCombine, pattern)
	}
	includePatterns, excludePatterns := r.combinedQuery.RegexpPatterns(searchquery.FieldFile)
	args := repoSearchArgs{
		query: &patternInfo{
			IsRegExp:                     true,
			IsCaseSensitive:              r.combinedQuery.IsCaseSensitive(),
			FileMatchLimit:               300,
			Pattern:                      regexpPatternMatchingExprsInOrder(patternsToCombine),
			IncludePatterns:              includePatterns,
			PathPatternsAreRegExps:       true,
			PathPatternsAreCaseSensitive: r.combinedQuery.IsCaseSensitive(),
		},
		repos: repos,
	}
	if len(excludePatterns) > 0 {
		excludePattern := unionRegExps(excludePatterns)
		args.query.ExcludePattern = &excludePattern
	}

	// Determine which types of results to return.
	var searchFuncs []func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error)
	var resultTypes []string
	if forceOnlyResultType != "" {
		resultTypes = []string{forceOnlyResultType}
	} else {
		resultTypes, _ = r.combinedQuery.StringValues(searchquery.FieldType)
		if len(resultTypes) == 0 {
			resultTypes = []string{"file"}

			// TODO(sqs): env var for opting into potentially slower default search type
			if v, _ := strconv.ParseBool(os.Getenv("EXP_SEARCH_PATHS")); v {
				resultTypes = append(resultTypes, "path")
			}
		}
	}
	seenResultTypes := make(map[string]struct{}, len(resultTypes))
	queryForFileMatches := false
	for _, resultType := range resultTypes {
		if _, seen := seenResultTypes[resultType]; seen {
			continue
		}
		seenResultTypes[resultType] = struct{}{}
		switch resultType {
		case "file":
			queryForFileMatches = true
			if len(patternsToCombine) == 0 {
				return nil, errors.New("no query terms or regexp specified")
			}
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchRepos(ctx, &args)
			})
		case "path":
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchPathsInRepos(ctx, args.repos, r.combinedQuery)
			})
		case "diff":
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchCommitDiffsInRepos(ctx, &args, r.combinedQuery)
			})
		case "commit":
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchCommitLogInRepos(ctx, &args, r.combinedQuery)
			})
		}
	}

	// Run all search funcs.
	results := &searchResults{
		queryForFileMatches: queryForFileMatches,
		combinedQuery:       r.args.ScopeQuery + " " + r.args.Query,
	}
	for _, searchFunc := range searchFuncs {
		results1, common1, err := searchFunc(ctx)
		if err != nil {
			return nil, err
		}
		if results1 == nil && common1 == nil {
			continue
		}
		results.results = append(results.results, results1...)
		// TODO(sqs): combine diff and commit results that refer to the same underlying
		// commit (and match on the commit's diff and message, respectively).
		results.searchResultsCommon.update(*common1)
	}

	tr.LazyPrintf("results=%d limitHit=%v cloning=%d missing=%d timedout=%d", len(results.results), results.searchResultsCommon.limitHit, len(results.searchResultsCommon.cloning), len(results.searchResultsCommon.missing), len(results.searchResultsCommon.timedout))

	if len(missingRepoRevs) > 0 {
		results.alert = r.alertForMissingRepoRevs(missingRepoRevs)
	}

	return results, nil
}

type searchResult struct {
	fileMatch *fileMatch
	diff      *commitSearchResult
}

func (g *searchResult) ToFileMatch() (*fileMatch, bool) { return g.fileMatch, g.fileMatch != nil }
func (g *searchResult) ToCommitSearchResult() (*commitSearchResult, bool) {
	return g.diff, g.diff != nil
}

// regexpPatternMatchingExprsInOrder returns a regexp that matches lines that contain
// non-overlapping matches for each pattern in order.
func regexpPatternMatchingExprsInOrder(patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}
	if len(patterns) == 1 {
		return patterns[0]
	}
	return "(" + strings.Join(patterns, ").*?(") + ")" // "?" makes it prefer shorter matches
}
