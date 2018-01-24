package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
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
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery/syntax"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

// searchResultsCommon contains fields that should be returned by all funcs
// that contribute to the overall search result set.
type searchResultsCommon struct {
	limitHit bool          // whether the limit on results was hit
	repos    []api.RepoURI // repos that were matched by the repo-related filters
	searched []api.RepoURI // repos that were searched
	indexed  []api.RepoURI // repos that were searched using an index
	cloning  []api.RepoURI // repos that could not be searched because they were still being cloned
	missing  []api.RepoURI // repos that could not be searched because they do not exist

	maxResultsCount, resultCount int32

	// timedout usually contains repos that haven't finished being fetched yet.
	// This should only happen for large repos and the searcher caches are
	// purged.
	timedout []api.RepoURI
}

func (c *searchResultsCommon) LimitHit() bool {
	return c.limitHit || c.resultCount > c.maxResultsCount
}

func (c *searchResultsCommon) Repositories() []string {
	if c.repos == nil {
		return []string{}
	}
	return repoURIsToStrings(c.repos)
}

func (c *searchResultsCommon) RepositoriesSearched() []string {
	if c.searched == nil {
		return []string{}
	}
	return repoURIsToStrings(c.searched)
}

func (c *searchResultsCommon) IndexedRepositoriesSearched() []string {
	if c.indexed == nil {
		return []string{}
	}
	return repoURIsToStrings(c.indexed)
}

func (c *searchResultsCommon) Cloning() []string {
	if c.cloning == nil {
		return []string{}
	}
	return repoURIsToStrings(c.cloning)
}

func (c *searchResultsCommon) Missing() []string {
	if c.missing == nil {
		return []string{}
	}
	return repoURIsToStrings(c.missing)
}

func (c *searchResultsCommon) Timedout() []string {
	if c.timedout == nil {
		return []string{}
	}
	return repoURIsToStrings(c.timedout)
}

// update updates c with the other data, deduping as necessary. It modifies c but
// does not modify other.
func (c *searchResultsCommon) update(other searchResultsCommon) {
	c.limitHit = c.limitHit || other.limitHit

	appendUnique := func(dst *[]api.RepoURI, src []api.RepoURI) {
		dstSet := make(map[api.RepoURI]struct{}, len(*dst))
		for _, s := range *dst {
			dstSet[s] = struct{}{}
		}
		for _, s := range src {
			if _, present := dstSet[s]; !present {
				*dst = append(*dst, s)
			}
		}
	}
	appendUnique(&c.repos, other.repos)
	appendUnique(&c.searched, other.searched)
	appendUnique(&c.indexed, other.indexed)
	appendUnique(&c.cloning, other.cloning)
	appendUnique(&c.missing, other.missing)
	appendUnique(&c.timedout, other.timedout)
	c.resultCount += other.resultCount
}

type searchResults struct {
	results []*searchResult
	searchResultsCommon
	alert           *searchAlert
	start           time.Time // when the results started being computed
	maxResultsCount int32
}

func (sr *searchResults) Results() []*searchResult {
	return sr.results
}

func (sr *searchResults) ResultCount() int32 {
	var totalResults int32
	for _, result := range sr.results {
		totalResults += result.resultCount()
	}
	return totalResults
}

func (sr *searchResults) ApproximateResultCount() string {
	if sr.alert != nil {
		return "?"
	}
	count := sr.ResultCount()
	if sr.LimitHit() || len(sr.cloning) > 0 || len(sr.timedout) > 0 {
		return fmt.Sprintf("%d+", count)
	}
	return strconv.Itoa(int(count))
}

func (sr *searchResults) Alert() *searchAlert { return sr.alert }

func (sr *searchResults) ElapsedMilliseconds() int32 {
	return int32(time.Since(sr.start).Nanoseconds() / int64(time.Millisecond))
}

// blameFileMatchCache caches Repos.GetByURI, Repos.ResolveRev, and RepoVCS.Open
// operations.
type blameFileMatchCache struct {
	cachedReposMu sync.RWMutex
	cachedRepos   map[api.RepoURI]*types.Repo

	cachedRevsMu sync.RWMutex
	cachedRevs   map[string]api.CommitID

	cachedVCSReposMu sync.RWMutex
	cachedVCSRepos   map[string]vcs.Repository
}

// repoVCSOpen is like localstore.Repos.GetByURI except it is cached by b.
func (b *blameFileMatchCache) reposGetByURI(ctx context.Context, repoURI api.RepoURI) (*types.Repo, error) {
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
func (b *blameFileMatchCache) reposResolveRev(ctx context.Context, repo api.RepoID, revStr string) (api.CommitID, error) {
	cacheKey := fmt.Sprint(repo, revStr)
	b.cachedRevsMu.RLock()
	rev, ok := b.cachedRevs[cacheKey]
	b.cachedRevsMu.RUnlock()
	if ok {
		return rev, nil
	}
	rev, err := backend.Repos.ResolveRev(ctx, repo, revStr)
	if err != nil {
		return "", err
	}
	b.cachedRevsMu.Lock()
	b.cachedRevs[cacheKey] = rev
	b.cachedRevsMu.Unlock()
	return rev, nil
}

// repoVCSOpen is like localstore.RepoVCS.Open except it is cached by b.
func (b *blameFileMatchCache) repoVCSOpen(ctx context.Context, repo api.RepoID) (vcs.Repository, error) {
	b.cachedVCSReposMu.RLock()
	vcsrepo, ok := b.cachedVCSRepos[fmt.Sprint(repo)]
	b.cachedVCSReposMu.RUnlock()
	if ok {
		return vcsrepo, nil
	}
	vcsrepo, err := db.RepoVCS.Open(ctx, repo)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	b.cachedVCSReposMu.Lock()
	b.cachedVCSRepos[fmt.Sprint(repo)] = vcsrepo
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
	repoURI := api.RepoURI(u.Host + u.Path)
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
		NewestCommit: rev,
		StartLine:    int(lm.LineNumber()),
		EndLine:      int(lm.LineNumber()),
	})
	if err != nil {
		return time.Time{}, err
	}

	return hunks[0].Author.Date, nil
}

func (sr *searchResults) Sparkline(ctx context.Context) (sparkline []int32, err error) {
	var (
		days     = 30                 // number of days the sparkline represents
		maxBlame = 100                // maximum number of file results to blame for date/time information.
		run      = parallel.NewRun(8) // number of concurrent blame ops
	)

	var (
		sparklineMu sync.Mutex
		blameOps    = 0
		cache       = &blameFileMatchCache{
			cachedRepos:    map[api.RepoURI]*types.Repo{},
			cachedRevs:     map[string]api.CommitID{},
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

type searchResultsStats struct {
	JApproximateResultCount string
	JSparkline              []int32
}

func (srs *searchResultsStats) ApproximateResultCount() string { return srs.JApproximateResultCount }
func (srs *searchResultsStats) Sparkline() []int32             { return srs.JSparkline }

var (
	searchResultsStatsCache   = rcache.NewWithTTL("search_results_stats", 3600) // 1h
	searchResultsStatsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "graphql",
		Name:      "search_results_stats_cache_hit",
		Help:      "Counts cache hits and misses for search results stats (e.g. sparklines).",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(searchResultsStatsCounter)
}

func (r *searchResolver) Stats(ctx context.Context) (stats *searchResultsStats, err error) {
	// Override user context to ensure that stats for this query are cached
	// regardless of the user context's cancellation. For example, if
	// stats/sparklines are slow to load on the homepage and all users navigate
	// away from that page before they load, no user would ever see them and we
	// would never cache them. This fixes that by ensuring the first request
	// 'kicks off loading' and places the result into cache regardless of
	// whether or not the original querier of this information still wants it.
	originalCtx := ctx
	ctx = context.Background()
	ctx = opentracing.ContextWithSpan(ctx, opentracing.SpanFromContext(originalCtx))

	cacheKey := r.rawQuery()
	// Check if value is in the cache.
	jsonRes, ok := searchResultsStatsCache.Get(cacheKey)
	if ok {
		searchResultsStatsCounter.WithLabelValues("hit").Inc()
		if err := json.Unmarshal(jsonRes, &stats); err != nil {
			return nil, err
		}
		return stats, nil
	}

	// Calculate value from scratch.
	searchResultsStatsCounter.WithLabelValues("miss").Inc()
	attempts := 0
	var v *searchResults
	for {
		// Query search results.
		var err error
		v, err = r.doResults(ctx, "")
		if err != nil {
			return nil, err // do not cache errors.
		}
		if v.ResultCount() > 0 {
			break
		}

		cloning := len(v.Cloning())
		timedout := len(v.Timedout())
		if cloning == 0 && timedout == 0 {
			break // zero results, but no cloning or timed out repos. No point in retrying.
		}

		if attempts > 5 {
			log15.Error("failed to generate sparkline due to cloning or timed out repos", "cloning", len(v.Cloning()), "timedout", len(v.Timedout()))
			return nil, fmt.Errorf("failed to generate sparkline due to %d cloning %d timedout repos", len(v.Cloning()), len(v.Timedout()))
		}

		// We didn't find any search results. Some repos are cloning or timed
		// out, so try again in a few seconds.
		attempts++
		log15.Warn("sparkline generation found 0 search results due to cloning or timed out repos (retrying in 5s)", "cloning", len(v.Cloning()), "timedout", len(v.Timedout()))
		time.Sleep(5 * time.Second)
	}

	sparkline, err := v.Sparkline(ctx)
	if err != nil {
		return nil, err // sparkline generation failed, so don't cache.
	}
	stats = &searchResultsStats{
		JApproximateResultCount: v.ApproximateResultCount(),
		JSparkline:              sparkline,
	}

	// Store in the cache if we got non-zero results. If we got zero results,
	// it should be quick and caching is not desired because e.g. it could be
	// a query for a repo that has not been added by the user yet.
	if v.ResultCount() > 0 {
		jsonRes, err = json.Marshal(stats)
		if err != nil {
			return nil, err
		}
		searchResultsStatsCache.Set(cacheKey, jsonRes)
	}
	return stats, nil

}

func (r *searchResolver) doResults(ctx context.Context, forceOnlyResultType string) (res *searchResults, err error) {
	traceName, ctx := traceutil.TraceName(ctx, "graphql.SearchResults")
	tr := trace.New(traceName, r.rawQuery())
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	start := time.Now()

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
	for _, v := range r.query.Values(searchquery.FieldDefault) {
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
	includePatterns, excludePatterns := r.query.RegexpPatterns(searchquery.FieldFile)
	args := repoSearchArgs{
		query: &patternInfo{
			IsRegExp:                     true,
			IsCaseSensitive:              r.query.IsCaseSensitive(),
			FileMatchLimit:               r.maxResults(),
			Pattern:                      regexpPatternMatchingExprsInOrder(patternsToCombine),
			IncludePatterns:              includePatterns,
			PathPatternsAreRegExps:       true,
			PathPatternsAreCaseSensitive: r.query.IsCaseSensitive(),
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
		resultTypes, _ = r.query.StringValues(searchquery.FieldType)
		if len(resultTypes) == 0 {
			resultTypes = []string{"file", "path"}
		}
	}
	seenResultTypes := make(map[string]struct{}, len(resultTypes))
	for _, resultType := range resultTypes {
		if resultType == "file" {
			args.query.PatternMatchesContent = true
		} else if resultType == "path" {
			args.query.PatternMatchesPath = true
		}
	}
	tr.LazyPrintf("resultTypes: %v", resultTypes)

	searchedFileContentsOrPaths := false
	for _, resultType := range resultTypes {
		if _, seen := seenResultTypes[resultType]; seen {
			continue
		}
		seenResultTypes[resultType] = struct{}{}
		switch resultType {
		case "file", "path":
			if searchedFileContentsOrPaths {
				// type:file and type:path use same searchRepos, so don't call 2x.
				continue
			}
			searchedFileContentsOrPaths = true
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchRepos(ctx, &args, r.query)
			})
		case "diff":
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchCommitDiffsInRepos(ctx, &args, r.query)
			})
		case "commit":
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchCommitLogInRepos(ctx, &args, r.query)
			})
		}
	}

	// Run all search funcs.
	results := searchResults{
		maxResultsCount:     r.maxResults(),
		start:               start,
		searchResultsCommon: searchResultsCommon{maxResultsCount: r.maxResults()},
	}
	for _, searchFunc := range searchFuncs {
		results1, common1, err := searchFunc(ctx)
		if results1 != nil {
			results.results = append(results.results, results1...)
			// TODO(sqs): combine diff and commit results that refer to the same underlying
			// commit (and match on the commit's diff and message, respectively).
		}
		if common1 != nil {
			results.searchResultsCommon.update(*common1)
		}
		if err != nil {
			// Return partial results if an error occurs.
			return &results, err
		}
	}

	tr.LazyPrintf("results=%d limitHit=%v cloning=%d missing=%d timedout=%d", len(results.results), results.searchResultsCommon.limitHit, len(results.searchResultsCommon.cloning), len(results.searchResultsCommon.missing), len(results.searchResultsCommon.timedout))

	if _, isDiff := seenResultTypes["diff"]; isDiff && results.alert == nil && !results.limitHit && len(r.query.Values("before")) == 0 && len(r.query.Values("after")) == 0 {
		results.alert = &searchAlert{
			description: "Diff search limited to last month by default. Use after: to search older commits.",
			proposedQueries: []*searchQueryDescription{
				{
					description: "commits in the last 6 months",
					query:       searchQuery{syntax.ExprString(r.query.Query.Syntax.Expr) + " after:\"6 months ago\""},
				},
				{
					description: "commits in the last 2 years",
					query:       searchQuery{syntax.ExprString(r.query.Query.Syntax.Expr) + " after:\"2 years ago\""},
				},
			},
		}
		if len(results.results) == 0 {
			results.alert.title = "No results found"
		} else {
			results.alert.title = "Only diff search results from last month are shown"
		}
	}
	if results.alert == nil && args.query.Pattern == "" {
		results.alert = &searchAlert{
			title:       "Type a query",
			description: "What do you want to search for?",
			proposedQueries: []*searchQueryDescription{
				{
					description: "Files containing the string \"func\"",
					query:       searchQuery{syntax.ExprString(r.query.Query.Syntax.Expr) + " func"},
				},
				{
					description: "Files containing the string \"open(\"",
					query:       searchQuery{syntax.ExprString(r.query.Query.Syntax.Expr) + " \"open(\""},
				},
			},
		}
	}
	if len(missingRepoRevs) > 0 {
		results.alert = r.alertForMissingRepoRevs(missingRepoRevs)
	}

	return &results, nil
}

type searchResult struct {
	fileMatch *fileMatch
	diff      *commitSearchResult
}

func (g *searchResult) ToFileMatch() (*fileMatch, bool) { return g.fileMatch, g.fileMatch != nil }
func (g *searchResult) ToCommitSearchResult() (*commitSearchResult, bool) {
	return g.diff, g.diff != nil
}

func (g *searchResult) resultCount() int32 {
	switch {
	case g.fileMatch != nil:
		if l := len(g.fileMatch.LineMatches()); l > 0 {
			return int32(l)
		}
		return 1 // 1 to count "empty" results like type:path results
	case g.diff != nil:
		return 1
	default:
		return 1
	}
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
