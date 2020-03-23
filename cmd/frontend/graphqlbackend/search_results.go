package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/src-d/enry/v2"

	"github.com/hashicorp/go-multierror"
	"github.com/neelance/parallel"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// searchResultsCommon contains fields that should be returned by all funcs
// that contribute to the overall search result set.
type searchResultsCommon struct {
	limitHit bool // whether the limit on results was hit

	repos    []*types.Repo             // repos that were matched by the repo-related filters
	searched []*types.Repo             // repos that were searched
	indexed  []*types.Repo             // repos that were searched using an index
	cloning  []*types.Repo             // repos that could not be searched because they were still being cloned
	missing  []*types.Repo             // repos that could not be searched because they do not exist
	partial  map[api.RepoName]struct{} // repos that were searched, but have results that were not returned due to exceeded limits

	maxResultsCount, resultCount int32

	// timedout usually contains repos that haven't finished being fetched yet.
	// This should only happen for large repos and the searcher caches are
	// purged.
	timedout []*types.Repo

	indexUnavailable bool // True if indexed search is enabled but was not available during this search.
}

func (c *searchResultsCommon) LimitHit() bool {
	return c.limitHit || c.resultCount > c.maxResultsCount
}

func (c *searchResultsCommon) Repositories() []*RepositoryResolver {
	return RepositoryResolvers(c.repos)
}

func (c *searchResultsCommon) RepositoriesCount() int32 {
	return int32(len(c.repos))
}

func (c *searchResultsCommon) RepositoriesSearched() []*RepositoryResolver {
	return RepositoryResolvers(c.searched)
}

func (c *searchResultsCommon) IndexedRepositoriesSearched() []*RepositoryResolver {
	return RepositoryResolvers(c.indexed)
}

func (c *searchResultsCommon) Cloning() []*RepositoryResolver {
	return RepositoryResolvers(c.cloning)
}

func (c *searchResultsCommon) Missing() []*RepositoryResolver {
	return RepositoryResolvers(c.missing)
}

func (c *searchResultsCommon) Timedout() []*RepositoryResolver {
	return RepositoryResolvers(c.timedout)
}

func (c *searchResultsCommon) IndexUnavailable() bool {
	return c.indexUnavailable
}

func (c *searchResultsCommon) Equal(other *searchResultsCommon) bool {
	return reflect.DeepEqual(c, other)
}

func RepositoryResolvers(repos types.Repos) []*RepositoryResolver {
	dedupSort(&repos)
	return toRepositoryResolvers(repos)
}

// update updates c with the other data, deduping as necessary. It modifies c but
// does not modify other.
func (c *searchResultsCommon) update(other searchResultsCommon) {
	c.limitHit = c.limitHit || other.limitHit
	c.indexUnavailable = c.indexUnavailable || other.indexUnavailable

	c.repos = append(c.repos, other.repos...)
	c.searched = append(c.searched, other.searched...)
	c.indexed = append(c.indexed, other.indexed...)
	c.cloning = append(c.cloning, other.cloning...)
	c.missing = append(c.missing, other.missing...)
	c.timedout = append(c.timedout, other.timedout...)
	c.resultCount += other.resultCount

	if c.partial == nil {
		c.partial = make(map[api.RepoName]struct{})
	}

	for repo := range other.partial {
		c.partial[repo] = struct{}{}
	}
}

// dedupSort sorts (by ID in ascending order) and deduplicates
// the given repos in-place.
func dedupSort(repos *types.Repos) {
	if len(*repos) == 0 {
		return
	}

	sort.Sort(*repos)

	j := 0
	for i := 1; i < len(*repos); i++ {
		if (*repos)[j].ID != (*repos)[i].ID {
			j++
			(*repos)[j] = (*repos)[i]
		}
	}

	*repos = (*repos)[:j+1]
}

// SearchResultsResolver is a resolver for the GraphQL type `SearchResults`
type SearchResultsResolver struct {
	SearchResults []SearchResultResolver
	searchResultsCommon
	alert *searchAlert
	start time.Time // when the results started being computed

	// cursor to return for paginated search requests, or nil if the request
	// wasn't paginated.
	cursor *searchCursor
}

func (sr *SearchResultsResolver) Results() []SearchResultResolver {
	return sr.SearchResults
}

func (sr *SearchResultsResolver) MatchCount() int32 {
	var totalResults int32
	for _, result := range sr.SearchResults {
		totalResults += result.resultCount()
	}
	return totalResults
}

func (sr *SearchResultsResolver) ResultCount() int32 { return sr.MatchCount() }

func (sr *SearchResultsResolver) ApproximateResultCount() string {
	count := sr.ResultCount()
	if sr.LimitHit() || len(sr.cloning) > 0 || len(sr.timedout) > 0 {
		return fmt.Sprintf("%d+", count)
	}
	return strconv.Itoa(int(count))
}

func (sr *SearchResultsResolver) Alert() *searchAlert { return sr.alert }

func (sr *SearchResultsResolver) ElapsedMilliseconds() int32 {
	return int32(time.Since(sr.start).Nanoseconds() / int64(time.Millisecond))
}

// commonFileFilters are common filters used. It is used by DynamicFilters to
// propose them if they match shown results.
var commonFileFilters = []struct {
	Regexp *lazyregexp.Regexp
	Filter string
}{
	// Exclude go tests
	{
		Regexp: lazyregexp.New(`_test\.go$`),
		Filter: `-file:_test\.go$`,
	},
	// Exclude go vendor
	{
		Regexp: lazyregexp.New(`(^|/)vendor/`),
		Filter: `-file:(^|/)vendor/`,
	},
	// Exclude node_modules
	{
		Regexp: lazyregexp.New(`(^|/)node_modules/`),
		Filter: `-file:(^|/)node_modules/`,
	},
}

func (sr *SearchResultsResolver) DynamicFilters() []*searchFilterResolver {
	filters := map[string]*searchFilterResolver{}
	repoToMatchCount := make(map[string]int)
	add := func(value string, label string, count int, limitHit bool, kind string) {
		sf, ok := filters[value]
		if !ok {
			sf = &searchFilterResolver{
				value:    value,
				label:    label,
				count:    int32(count),
				limitHit: limitHit,
				kind:     kind,
			}
			filters[value] = sf
		} else {
			sf.count = int32(count)
		}

		sf.score++
	}

	addRepoFilter := func(uri string, rev string, lineMatchCount int) {
		filter := fmt.Sprintf(`repo:^%s$`, regexp.QuoteMeta(uri))
		if rev != "" {
			filter = filter + fmt.Sprintf(`@%s`, regexp.QuoteMeta(rev))
		}
		_, limitHit := sr.searchResultsCommon.partial[api.RepoName(uri)]
		// Increment number of matches per repo. Add will override previous entry for uri
		repoToMatchCount[uri] += lineMatchCount
		add(filter, uri, repoToMatchCount[uri], limitHit, "repo")
	}

	addFileFilter := func(fileMatchPath string, lineMatchCount int, limitHit bool) {
		for _, ff := range commonFileFilters {
			if ff.Regexp.MatchString(fileMatchPath) {
				add(ff.Filter, ff.Filter, lineMatchCount, limitHit, "file")
			}
		}
	}

	addLangFilter := func(fileMatchPath string, lineMatchCount int, limitHit bool) {
		extensionToLanguageLookup := func(path string) string {
			language, _ := inventory.GetLanguageByFilename(path)
			return strings.ToLower(language)
		}
		if ext := path.Ext(fileMatchPath); ext != "" {
			language := extensionToLanguageLookup(fileMatchPath)
			if language != "" {
				if strings.Contains(language, " ") {
					language = strconv.Quote(language)
				}
				value := fmt.Sprintf(`lang:%s`, language)
				add(value, value, lineMatchCount, limitHit, "lang")
			}
		}
	}

	for _, result := range sr.SearchResults {
		if fm, ok := result.ToFileMatch(); ok {
			rev := ""
			if fm.InputRev != nil {
				rev = *fm.InputRev
			}
			addRepoFilter(string(fm.Repo.Name), rev, len(fm.LineMatches()))
			addLangFilter(fm.JPath, len(fm.LineMatches()), fm.JLimitHit)
			addFileFilter(fm.JPath, len(fm.LineMatches()), fm.JLimitHit)

			if len(fm.symbols) > 0 {
				add("type:symbol", "type:symbol", 1, fm.JLimitHit, "symbol")
			}
		} else if r, ok := result.ToRepository(); ok {
			// It should be fine to leave this blank since revision specifiers
			// can only be used with the 'repo:' scope. In that case,
			// we shouldn't be getting any repositoy name matches back.
			addRepoFilter(r.Name(), "", 1)
		}
	}

	filterSlice := make([]*searchFilterResolver, 0, len(filters))
	repoFilterSlice := make([]*searchFilterResolver, 0, len(filters)/2) // heuristic - half of all filters are repo filters.
	for _, f := range filters {
		if f.kind == "repo" {
			repoFilterSlice = append(repoFilterSlice, f)
		} else {
			filterSlice = append(filterSlice, f)
		}
	}
	sort.Slice(filterSlice, func(i, j int) bool {
		return filterSlice[j].score < filterSlice[i].score
	})
	// limit amount of non-repo filters to be rendered arbitrarily to 12
	if len(filterSlice) > 12 {
		filterSlice = filterSlice[:12]
	}

	allFilters := append(filterSlice, repoFilterSlice...)
	sort.Slice(allFilters, func(i, j int) bool {
		return allFilters[j].score < allFilters[i].score
	})

	return allFilters
}

type searchFilterResolver struct {
	value string

	// the string to be displayed in the UI
	label string

	// the number of matches in a particular repository. Only used for `repo:` filters.
	count int32

	// whether the results returned for a repository are incomplete
	limitHit bool

	// the kind of filter. Should be "repo", "file", or "lang".
	kind string

	// score is used to select potential filters
	score int
}

func (sf *searchFilterResolver) Value() string {
	return sf.value
}

func (sf *searchFilterResolver) Label() string {
	return sf.label
}

func (sf *searchFilterResolver) Count() int32 {
	return sf.count
}

func (sf *searchFilterResolver) LimitHit() bool {
	return sf.limitHit
}

func (sf *searchFilterResolver) Kind() string {
	return sf.kind
}

// blameFileMatch blames the specified file match to produce the time at which
// the first line match inside of it was authored.
func (sr *SearchResultsResolver) blameFileMatch(ctx context.Context, fm *FileMatchResolver) (t time.Time, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "blameFileMatch")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Blame the first line match.
	lineMatches := fm.LineMatches()
	if len(lineMatches) == 0 {
		// No line match
		return time.Time{}, nil
	}
	lm := fm.LineMatches()[0]
	hunks, err := git.BlameFile(ctx, gitserver.Repo{Name: fm.Repo.Name}, fm.JPath, &git.BlameOptions{
		NewestCommit: fm.CommitID,
		StartLine:    int(lm.LineNumber()),
		EndLine:      int(lm.LineNumber()),
	})
	if err != nil {
		return time.Time{}, err
	}

	return hunks[0].Author.Date, nil
}

func (sr *SearchResultsResolver) Sparkline(ctx context.Context) (sparkline []int32, err error) {
	var (
		days     = 30                 // number of days the sparkline represents
		maxBlame = 100                // maximum number of file results to blame for date/time information.
		run      = parallel.NewRun(8) // number of concurrent blame ops
	)

	var (
		sparklineMu sync.Mutex
		blameOps    = 0
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
	for _, r := range sr.SearchResults {
		r := r // shadow so it doesn't change in the goroutine
		switch m := r.(type) {
		case *RepositoryResolver:
			// We don't care about repo results here.
			continue
		case *commitSearchResultResolver:
			// Diff searches are cheap, because we implicitly have author date info.
			addPoint(m.commit.author.date)
		case *FileMatchResolver:
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
			goroutine.Go(func() {
				defer run.Release()

				// Blame the file match in order to retrieve date informatino.
				var err error
				t, err := sr.blameFileMatch(ctx, m)
				if err != nil {
					log15.Warn("failed to blame fileMatch during sparkline generation", "error", err)
					return
				}
				addPoint(t)
			})
		case *codemodResultResolver:
			continue
		default:
			panic("SearchResults.Sparkline unexpected union type state")
		}
	}
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("blame_ops", blameOps)
	return sparkline, nil
}

var searchResponseCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "graphql",
	Name:      "search_response",
	Help:      "Number of searches that have ended in the given status (success, error, timeout, partial_timeout).",
}, []string{"status", "alert_type"})

// logSearchLatency records search durations in the event database. This
// function may only be called after a search result is performed, because it
// relies on the invariant that query and pattern error checking has already
// been performed.
func (r *searchResolver) logSearchLatency(ctx context.Context, durationMs int32) {
	var types []string
	resultTypes, _ := r.query.StringValues(query.FieldType)
	for _, typ := range resultTypes {
		switch typ {
		case "repo", "symbol", "diff", "commit":
			types = append(types, typ)
		case "path":
			// Map type:path to file
			types = append(types, "file")
		case "file":
			switch {
			case r.patternType == query.SearchTypeStructural:
				types = append(types, "structural")
			case r.patternType == query.SearchTypeLiteral:
				types = append(types, "literal")
			case r.patternType == query.SearchTypeRegex:
				types = append(types, "regexp")
			}
		}
	}

	// Don't record composite searches that specify more than one type:
	// because we can't break down the search timings into multiple
	// categories.
	if len(types) > 1 {
		return
	}

	options := &getPatternInfoOptions{}
	if r.patternType == query.SearchTypeStructural {
		options = &getPatternInfoOptions{performStructuralSearch: true}
	}
	if r.patternType == query.SearchTypeLiteral {
		options = &getPatternInfoOptions{performLiteralSearch: true}
	}
	p, _ := r.getPatternInfo(options)

	// If no type: was explicitly specified, infer the result type.
	if len(types) == 0 {
		// If a pattern was specified, a content search happened.
		if p.Pattern != "" {
			switch {
			case r.patternType == query.SearchTypeStructural:
				types = append(types, "structural")
			case r.patternType == query.SearchTypeLiteral:
				types = append(types, "literal")
			case r.patternType == query.SearchTypeRegex:
				types = append(types, "regexp")
			}
		} else if len(r.query.Fields["file"]) > 0 {
			// No search pattern specified and file: is specified.
			types = append(types, "file")
		} else {
			// No search pattern or file: is specified, assume repo.
			// This includes accounting for searches of fields that
			// specify repohasfile: and repohascommitafter:.
			types = append(types, "repo")
		}
	}

	// Only log the time if we successfully resolved one search type.
	if len(types) == 1 {
		actor := actor.FromContext(ctx)
		if actor.IsAuthenticated() {
			value := fmt.Sprintf(`{"durationMs": %d}`, durationMs)
			eventName := fmt.Sprintf("search.latencies.%s", types[0])
			err := usagestats.LogBackendEvent(actor.UID, eventName, json.RawMessage(value))
			if err != nil {
				log15.Warn("Could not log search latency", "err", err)
			}
		}
	}
}

func (r *searchResolver) Results(ctx context.Context) (*SearchResultsResolver, error) {
	// If the request is a paginated one, we handle it separately. See
	// paginatedResults for more details.
	if r.pagination != nil {
		return r.paginatedResults(ctx)
	}

	rr, err := r.resultsWithTimeoutSuggestion(ctx)
	if rr != nil {
		r.logSearchLatency(ctx, rr.ElapsedMilliseconds())
	}

	// Record what type of response we sent back via Prometheus.
	var status, alertType string
	switch {
	case err == context.DeadlineExceeded || (err == nil && len(rr.searchResultsCommon.timedout) > 0 && len(rr.searchResultsCommon.timedout) == len(rr.searchResultsCommon.repos)):
		status = "timeout"
	case err == nil && len(rr.searchResultsCommon.timedout) > 0:
		status = "partial_timeout"
	case err == nil && rr.alert != nil:
		status = "alert"
		alertType = rr.alert.prometheusType
	case err != nil:
		status = "error"
	case err == nil:
		status = "success"
	default:
		status = "unknown"
	}
	searchResponseCounter.WithLabelValues(status, alertType).Inc()

	return rr, err
}

// resultsWithTimeoutSuggestion calls doResults, and in case of deadline
// exceeded returns a search alert with a did-you-mean link for the same
// query with a longer timeout.
func (r *searchResolver) resultsWithTimeoutSuggestion(ctx context.Context) (*SearchResultsResolver, error) {
	start := time.Now()
	rr, err := r.doResults(ctx, "")

	// If we encountered a context timeout, it indicates one of the many result
	// type searchers (file, diff, symbol, etc) completely timed out and could not
	// produce even partial results. Other searcher types may have produced results.
	//
	// In this case, or if we got a partial timeout where ALL repositories timed out,
	// we do not return partial results and instead display a timeout alert.
	shouldShowAlert := err == context.DeadlineExceeded
	if err == nil && len(rr.searchResultsCommon.timedout) > 0 && len(rr.searchResultsCommon.timedout) == len(rr.searchResultsCommon.repos) {
		shouldShowAlert = true
	}
	if shouldShowAlert {
		usedTime := time.Since(start)
		suggestTime := longer(2, usedTime)
		return &SearchResultsResolver{alert: alertForTimeout(usedTime, suggestTime, r)}, nil
	}
	return rr, err
}

// longer returns a suggested longer time to wait if the given duration wasn't long enough.
func longer(N int, dt time.Duration) time.Duration {
	dt2 := func() time.Duration {
		Ndt := time.Duration(N) * dt
		dceil := func(x float64) time.Duration {
			return time.Duration(math.Ceil(x))
		}
		switch {
		case math.Floor(Ndt.Hours()) > 0:
			return dceil(Ndt.Hours()) * time.Hour
		case math.Floor(Ndt.Minutes()) > 0:
			return dceil(Ndt.Minutes()) * time.Minute
		case math.Floor(Ndt.Seconds()) > 0:
			return dceil(Ndt.Seconds()) * time.Second
		default:
			return 0
		}
	}()
	lowest := 2 * time.Second
	if dt2 < lowest {
		return lowest
	}
	return dt2
}

var decimalRx = lazyregexp.New(`\d+\.\d+`)

// roundStr rounds the first number containing a decimal within a string
func roundStr(s string) string {
	return decimalRx.ReplaceAllStringFunc(s, func(ns string) string {
		f, err := strconv.ParseFloat(ns, 64)
		if err != nil {
			return s
		}
		f = math.Round(f)
		return strconv.Itoa(int(f))
	})
}

type searchResultsStats struct {
	JApproximateResultCount string
	JSparkline              []int32

	sr *searchResolver

	once   sync.Once
	srs    *SearchResultsResolver
	srsErr error
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
		stats.sr = r
		return stats, nil
	}

	// Calculate value from scratch.
	searchResultsStatsCounter.WithLabelValues("miss").Inc()
	attempts := 0
	var v *SearchResultsResolver
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
		sr:                      r,
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

type getPatternInfoOptions struct {
	// forceFileSearch, when true, specifies that the search query should be
	// treated as if every default term had `file:` before it. This can be used
	// to allow users to jump to files by just typing their name.
	forceFileSearch         bool
	performStructuralSearch bool
	performLiteralSearch    bool

	fileMatchLimit int32
}

// getPatternInfo gets the search pattern info for the query in the resolver.
func (r *searchResolver) getPatternInfo(opts *getPatternInfoOptions) (*search.TextPatternInfo, error) {
	if opts == nil {
		opts = &getPatternInfoOptions{}
	}

	if opts.fileMatchLimit == 0 {
		opts.fileMatchLimit = r.maxResults()
	}

	return getPatternInfo(r.query, opts)
}

// processSearchPattern processes the search pattern for a query. It handles the interpretation of search patterns
// as literal, regex, or structural patterns, and applies fuzzy regex matching if applicable.
func processSearchPattern(q *query.Query, opts *getPatternInfoOptions) (string, bool, bool) {
	var pattern string
	var pieces []string
	var contentFieldSet bool
	isRegExp := false
	isStructuralPat := false

	patternValues := q.Values(query.FieldDefault)
	if overridePattern := q.Values(query.FieldContent); len(overridePattern) > 0 {
		patternValues = overridePattern
		contentFieldSet = true
	}

	if opts.performStructuralSearch {
		isStructuralPat = true
		for _, v := range patternValues {
			if piece := v.ToString(); piece != "" {
				pieces = append(pieces, piece)
			}
		}
		pattern = strings.Join(pieces, " ")
	} else if !opts.forceFileSearch {
		isRegExp = true
		for _, v := range patternValues {
			var piece string
			switch {
			case v.String != nil:
				if contentFieldSet && !opts.performLiteralSearch {
					piece = *v.String
				} else {
					// Treat quoted strings as literal
					// strings to match, not regexps.
					piece = regexp.QuoteMeta(*v.String)
				}
			case v.Regexp != nil:
				piece = v.Regexp.String()
			}
			if piece == "" {
				continue
			}
			pieces = append(pieces, piece)
		}
		pattern = orderedFuzzyRegexp(pieces)
	} else {
		// TODO: We must have some pattern that always matches here, or else
		// cmd/searcher/search/matcher.go:97 would cause a nil regexp panic
		// when not using indexed search. I am unsure what the right solution
		// is here. Would this code path go away when we switch fully to
		// indexed search @keegan? This workaround is OK for now though.
		isRegExp = true
		pattern = "."
	}

	return pattern, isRegExp, isStructuralPat
}

// getPatternInfo gets the search pattern info for q
func getPatternInfo(q *query.Query, opts *getPatternInfoOptions) (*search.TextPatternInfo, error) {
	pattern, isRegExp, isStructuralPat := processSearchPattern(q, opts)

	// Handle file: and -file: filters.
	includePatterns, excludePatterns := q.RegexpPatterns(query.FieldFile)
	filePatternsReposMustInclude, filePatternsReposMustExclude := q.RegexpPatterns(query.FieldRepoHasFile)

	if opts.forceFileSearch {
		for _, v := range q.Values(query.FieldDefault) {
			includePatterns = append(includePatterns, v.ToString())
		}
	}

	var combyRule []string
	for _, v := range q.Values(query.FieldCombyRule) {
		combyRule = append(combyRule, v.ToString())
	}

	// Handle lang: and -lang: filters.
	langIncludePatterns, langExcludePatterns, err := langIncludeExcludePatterns(q.StringValues(query.FieldLang))
	if err != nil {
		return nil, err
	}
	includePatterns = append(includePatterns, langIncludePatterns...)
	excludePatterns = append(excludePatterns, langExcludePatterns...)

	languages, _ := q.StringValues(query.FieldLang)

	patternInfo := &search.TextPatternInfo{
		IsRegExp:                     isRegExp,
		IsStructuralPat:              isStructuralPat,
		IsCaseSensitive:              q.IsCaseSensitive(),
		FileMatchLimit:               opts.fileMatchLimit,
		Pattern:                      pattern,
		IncludePatterns:              includePatterns,
		FilePatternsReposMustInclude: filePatternsReposMustInclude,
		FilePatternsReposMustExclude: filePatternsReposMustExclude,
		PathPatternsAreRegExps:       true,
		Languages:                    languages,
		PathPatternsAreCaseSensitive: q.IsCaseSensitive(),
		CombyRule:                    strings.Join(combyRule, ""),
	}
	if len(excludePatterns) > 0 {
		patternInfo.ExcludePattern = unionRegExps(excludePatterns)
	}
	return patternInfo, nil
}

// langIncludeExcludePatterns returns regexps for the include/exclude path patterns given the lang:
// and -lang: filter values in a search query. For example, a query containing "lang:go" should
// include files whose paths match /\.go$/.
func langIncludeExcludePatterns(values, negatedValues []string) (includePatterns, excludePatterns []string, err error) {
	do := func(values []string, patterns *[]string) error {
		for _, value := range values {
			lang, ok := enry.GetLanguageByAlias(value)
			if !ok {
				return fmt.Errorf("unknown language: %q", value)
			}
			exts := enry.GetLanguageExtensions(lang)
			extPatterns := make([]string, len(exts))
			for i, ext := range exts {
				// Add `\.ext$` pattern to match files with the given extension.
				extPatterns[i] = regexp.QuoteMeta(ext) + "$"
			}
			*patterns = append(*patterns, unionRegExps(extPatterns))
		}
		return nil
	}

	if err := do(values, &includePatterns); err != nil {
		return nil, nil, err
	}
	if err := do(negatedValues, &excludePatterns); err != nil {
		return nil, nil, err
	}
	return includePatterns, excludePatterns, nil
}

var (
	// The default timeout to use for queries.
	defaultTimeout = 20 * time.Second
	// The max timeout to use for queries.
	maxTimeout = time.Minute
)

func (r *searchResolver) searchTimeoutFieldSet() bool {
	timeout, _ := r.query.StringValue(query.FieldTimeout)
	return timeout != "" || r.countIsSet()
}

func (r *searchResolver) withTimeout(ctx context.Context) (context.Context, context.CancelFunc, error) {
	d := defaultTimeout
	timeout, _ := r.query.StringValue(query.FieldTimeout)
	if timeout != "" {
		var err error
		d, err = time.ParseDuration(timeout)
		if err != nil {
			return nil, nil, errors.WithMessage(err, `invalid "timeout:" value (examples: "timeout:2s", "timeout:200ms")`)
		}
	} else if r.countIsSet() {
		// If `count:` is set but `timeout:` is not explicitly set, use the max timeout
		d = maxTimeout
	}
	// don't run queries longer than 1 minute.
	if d.Minutes() > 1 {
		d = maxTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, d)
	return ctx, cancel, nil
}

func (r *searchResolver) determineResultTypes(args search.TextParameters, forceOnlyResultType string) (resultTypes []string) {
	// Determine which types of results to return.
	if forceOnlyResultType != "" {
		resultTypes = []string{forceOnlyResultType}
	} else if len(r.query.Values(query.FieldReplace)) > 0 {
		resultTypes = []string{"codemod"}
	} else {
		resultTypes, _ = r.query.StringValues(query.FieldType)
		if len(resultTypes) == 0 {
			resultTypes = []string{"file", "path", "repo"}
		}
	}
	for _, resultType := range resultTypes {
		if resultType == "file" {
			args.PatternInfo.PatternMatchesContent = true
		} else if resultType == "path" {
			args.PatternInfo.PatternMatchesPath = true
		}
	}
	return resultTypes
}

func (r *searchResolver) determineRepos(ctx context.Context, tr *trace.Trace, start time.Time) (repos, missingRepoRevs []*search.RepositoryRevisions, res *SearchResultsResolver, err error) {
	repos, missingRepoRevs, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		if errors.Is(err, authz.ErrStalePermissions{}) {
			log15.Debug("searchResolver.determineRepos", "err", err)
			alert := alertForStalePermissions()
			return nil, nil, &SearchResultsResolver{alert: alert, start: start}, nil
		}
		return nil, nil, nil, err
	}

	tr.LazyPrintf("searching %d repos, %d missing", len(repos), len(missingRepoRevs))
	if len(repos) == 0 {
		alert := r.alertForNoResolvedRepos(ctx)
		return nil, nil, &SearchResultsResolver{alert: alert, start: start}, nil
	}
	if overLimit {
		alert, err := r.alertForOverRepoLimit(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		return nil, nil, &SearchResultsResolver{alert: alert, start: start}, nil
	}
	return repos, missingRepoRevs, nil, nil
}

// Surface an alert if a query exceeds limits that we place on search. Currently limits
// diff and commit searches where more than repoLimit repos need to be searched.
func alertOnSearchLimit(resultTypes []string, args *search.TextParameters) ([]string, *searchAlert) {
	var alert *searchAlert
	repoLimit := 50
	if len(args.Repos) > repoLimit {
		if len(resultTypes) == 1 {
			resultType := resultTypes[0]
			switch resultType {
			case "commit", "diff":
				if _, afterPresent := args.Query.Fields["after"]; afterPresent {
					break
				}
				if _, beforePresent := args.Query.Fields["before"]; beforePresent {
					break
				}
				resultTypes = []string{}
				alert = &searchAlert{
					prometheusType: "exceeded_diff_commit_search_limit",
					title:          fmt.Sprintf("Too many matching repositories for %s search to handle", resultType),
					description:    fmt.Sprintf(`%s search can currently only handle searching over %d repositories at a time. Try using the "repo:" filter to narrow down which repositories to search, or using 'after:"1 week ago"'. Tracking issue: https://github.com/sourcegraph/sourcegraph/issues/6826`, strings.Title(resultType), repoLimit),
				}
			}
		}
	}
	return resultTypes, alert
}

// doResults is one of the highest level search functions that handles finding results.
//
// If forceOnlyResultType is specified, only results of the given type are returned,
// regardless of what `type:` is specified in the query string.
//
// Partial results AND an error may be returned.
func (r *searchResolver) doResults(ctx context.Context, forceOnlyResultType string) (res *SearchResultsResolver, err error) {
	tr, ctx := trace.New(ctx, "graphql.SearchResults", r.rawQuery())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	start := time.Now()

	ctx, cancel, err := r.withTimeout(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	repos, missingRepoRevs, alertResult, err := r.determineRepos(ctx, tr, start)
	if err != nil {
		return nil, err
	}
	if alertResult != nil {
		return alertResult, nil
	}

	options := &getPatternInfoOptions{}
	if r.patternType == query.SearchTypeStructural {
		options = &getPatternInfoOptions{performStructuralSearch: true}
		forceOnlyResultType = "file"
	}
	if r.patternType == query.SearchTypeLiteral {
		options = &getPatternInfoOptions{performLiteralSearch: true}
	}
	p, err := r.getPatternInfo(options)
	if err != nil {
		return nil, err
	}

	// Fallback to literal search for searching repos and files if
	// the structural search pattern is empty.
	if r.patternType == query.SearchTypeStructural && p.Pattern == "" {
		r.patternType = query.SearchTypeLiteral
		p.IsStructuralPat = false
		forceOnlyResultType = ""
	}

	args := search.TextParameters{
		PatternInfo:     p,
		Repos:           repos,
		Query:           r.query,
		UseFullDeadline: r.searchTimeoutFieldSet(),
		Zoekt:           r.zoekt,
		SearcherURLs:    r.searcherURLs,
	}
	if err := args.PatternInfo.Validate(); err != nil {
		return nil, &badRequestError{err}
	}

	err = validateRepoHasFileUsage(r.query)
	if err != nil {
		return nil, err
	}

	resultTypes := r.determineResultTypes(args, forceOnlyResultType)
	tr.LazyPrintf("resultTypes: %v", resultTypes)

	var (
		requiredWg sync.WaitGroup
		optionalWg sync.WaitGroup
		results    []SearchResultResolver
		resultsMu  sync.Mutex
		common     = searchResultsCommon{maxResultsCount: r.maxResults()}
		commonMu   sync.Mutex
		multiErr   *multierror.Error
		multiErrMu sync.Mutex
		// fileMatches is a map from git:// URI of the file to FileMatch resolver
		// to merge multiple results of different types for the same file
		fileMatches   = make(map[string]*FileMatchResolver)
		fileMatchesMu sync.Mutex
		// Alert is a potential alert shown to the user.
		alert           *searchAlert
		seenResultTypes = make(map[string]struct{})
	)

	waitGroup := func(required bool) *sync.WaitGroup {
		if args.UseFullDeadline {
			// When a custom timeout is specified, all searches are required and get the full timeout.
			return &requiredWg
		}
		if required {
			return &requiredWg
		}
		return &optionalWg
	}

	// Apply search limits and generate warnings before firing off workers.
	// This currently limits diff and commit search to a set number of
	// repos, and removes the diff and commit resultTypes if it is breached.
	resultTypes, alert = alertOnSearchLimit(resultTypes, &args)

	searchedFileContentsOrPaths := false
	for _, resultType := range resultTypes {
		resultType := resultType // shadow so it doesn't change in the goroutine
		if _, seen := seenResultTypes[resultType]; seen {
			continue
		}
		seenResultTypes[resultType] = struct{}{}
		switch resultType {
		case "repo":
			// Search for repos
			wg := waitGroup(true)
			wg.Add(1)
			goroutine.Go(func() {
				defer wg.Done()

				repoResults, repoCommon, err := searchRepositories(ctx, &args, r.maxResults())
				// Timeouts are reported through searchResultsCommon so don't report an error for them
				if err != nil && !isContextError(ctx, err) {
					multiErrMu.Lock()
					multiErr = multierror.Append(multiErr, errors.Wrap(err, "repository search failed"))
					multiErrMu.Unlock()
				}
				if repoResults != nil {
					resultsMu.Lock()
					results = append(results, repoResults...)
					resultsMu.Unlock()
				}
				if repoCommon != nil {
					commonMu.Lock()
					common.update(*repoCommon)
					commonMu.Unlock()
				}
			})
		case "symbol":
			wg := waitGroup(len(resultTypes) == 1)
			wg.Add(1)
			goroutine.Go(func() {
				defer wg.Done()

				symbolFileMatches, symbolsCommon, err := searchSymbols(ctx, &args, int(r.maxResults()))
				// Timeouts are reported through searchResultsCommon so don't report an error for them
				if err != nil && !isContextError(ctx, err) {
					multiErrMu.Lock()
					multiErr = multierror.Append(multiErr, errors.Wrap(err, "symbol search failed"))
					multiErrMu.Unlock()
				}
				for _, symbolFileMatch := range symbolFileMatches {
					key := symbolFileMatch.uri
					fileMatchesMu.Lock()
					if m, ok := fileMatches[key]; ok {
						m.symbols = symbolFileMatch.symbols
					} else {
						fileMatches[key] = symbolFileMatch
						resultsMu.Lock()
						results = append(results, symbolFileMatch)
						resultsMu.Unlock()
					}
					fileMatchesMu.Unlock()
				}
				if symbolsCommon != nil {
					commonMu.Lock()
					common.update(*symbolsCommon)
					commonMu.Unlock()
				}
			})
		case "file", "path":
			if searchedFileContentsOrPaths {
				// type:file and type:path use same searchFilesInRepos, so don't call 2x.
				continue
			}
			searchedFileContentsOrPaths = true
			wg := waitGroup(true)
			wg.Add(1)
			goroutine.Go(func() {
				defer wg.Done()

				fileResults, fileCommon, err := searchFilesInRepos(ctx, &args)
				// Timeouts are reported through searchResultsCommon so don't report an error for them
				if err != nil && !isContextError(ctx, err) {
					multiErrMu.Lock()
					multiErr = multierror.Append(multiErr, errors.Wrap(err, "text search failed"))
					multiErrMu.Unlock()
				}
				if args.PatternInfo.IsStructuralPat && args.PatternInfo.FileMatchLimit == defaultMaxSearchResults && len(fileResults) == 0 {
					// No results for structural search? Automatically search again and force Zoekt to resolve
					// more potential file matches by setting a higher FileMatchLimit.
					args.PatternInfo.FileMatchLimit = 1000
					fileResults, fileCommon, err = searchFilesInRepos(ctx, &args)
					if err != nil && !isContextError(ctx, err) {
						multiErrMu.Lock()
						multiErr = multierror.Append(multiErr, errors.Wrap(err, "text search failed"))
						multiErrMu.Unlock()
					}
					if len(fileResults) == 0 && fileCommon.limitHit {
						// Still no results? Give up.
						log15.Warn("Structural search gives up after more exhaustive attempt. Results may have been missed.")
						fileCommon.limitHit = false // Ensure we don't display "Show more".
					}
				}
				for _, r := range fileResults {
					key := r.uri
					fileMatchesMu.Lock()
					m, ok := fileMatches[key]
					if ok {
						// merge line match results with an existing symbol result
						m.JLimitHit = m.JLimitHit || r.JLimitHit
						m.JLineMatches = r.JLineMatches
					} else {
						fileMatches[key] = r
						resultsMu.Lock()
						results = append(results, r)
						resultsMu.Unlock()
					}
					fileMatchesMu.Unlock()
				}
				if fileCommon != nil {
					commonMu.Lock()
					common.update(*fileCommon)
					commonMu.Unlock()
				}
			})
		case "diff":
			wg := waitGroup(len(resultTypes) == 1)
			wg.Add(1)
			goroutine.Go(func() {
				defer wg.Done()
				old := args.PatternInfo
				patternInfo := &search.CommitPatternInfo{
					Pattern:                      old.Pattern,
					IsRegExp:                     old.IsRegExp,
					IsCaseSensitive:              old.IsCaseSensitive,
					FileMatchLimit:               old.FileMatchLimit,
					IncludePatterns:              old.IncludePatterns,
					ExcludePattern:               old.ExcludePattern,
					PathPatternsAreRegExps:       old.PathPatternsAreRegExps,
					PathPatternsAreCaseSensitive: p.PathPatternsAreCaseSensitive,
				}
				args := search.TextParametersForCommitParameters{
					PatternInfo: patternInfo,
					Repos:       args.Repos,
					Query:       args.Query,
				}
				diffResults, diffCommon, err := searchCommitDiffsInRepos(ctx, &args)
				// Timeouts are reported through searchResultsCommon so don't report an error for them
				if err != nil && !isContextError(ctx, err) {
					multiErrMu.Lock()
					multiErr = multierror.Append(multiErr, errors.Wrap(err, "diff search failed"))
					multiErrMu.Unlock()
				}
				if diffResults != nil {
					resultsMu.Lock()
					results = append(results, diffResults...)
					resultsMu.Unlock()
				}
				if diffCommon != nil {
					commonMu.Lock()
					common.update(*diffCommon)
					commonMu.Unlock()
				}
			})
		case "commit":
			wg := waitGroup(len(resultTypes) == 1)
			wg.Add(1)
			goroutine.Go(func() {
				defer wg.Done()

				old := args.PatternInfo
				patternInfo := &search.CommitPatternInfo{
					Pattern:                      old.Pattern,
					IsRegExp:                     old.IsRegExp,
					IsCaseSensitive:              old.IsCaseSensitive,
					FileMatchLimit:               old.FileMatchLimit,
					IncludePatterns:              old.IncludePatterns,
					ExcludePattern:               old.ExcludePattern,
					PathPatternsAreRegExps:       old.PathPatternsAreRegExps,
					PathPatternsAreCaseSensitive: old.PathPatternsAreCaseSensitive,
				}
				args := search.TextParametersForCommitParameters{
					PatternInfo: patternInfo,
					Repos:       args.Repos,
					Query:       args.Query,
				}
				commitResults, commitCommon, err := searchCommitLogInRepos(ctx, &args)
				// Timeouts are reported through searchResultsCommon so don't report an error for them
				if err != nil && !isContextError(ctx, err) {
					multiErrMu.Lock()
					multiErr = multierror.Append(multiErr, errors.Wrap(err, "commit search failed"))
					multiErrMu.Unlock()
				}
				if commitResults != nil {
					resultsMu.Lock()
					results = append(results, commitResults...)
					resultsMu.Unlock()
				}
				if commitCommon != nil {
					commonMu.Lock()
					common.update(*commitCommon)
					commonMu.Unlock()
				}
			})
		case "codemod":
			wg := waitGroup(true)
			wg.Add(1)
			goroutine.Go(func() {
				defer wg.Done()

				codemodResults, codemodCommon, err := performCodemod(ctx, &args)
				// Timeouts are reported through searchResultsCommon so don't report an error for them
				if err != nil && !isContextError(ctx, err) {
					multiErrMu.Lock()
					multiErr = multierror.Append(multiErr, errors.Wrap(err, "codemod search failed"))
					multiErrMu.Unlock()
				}
				if codemodResults != nil {
					resultsMu.Lock()
					results = append(results, codemodResults...)
					resultsMu.Unlock()
				}
				if codemodCommon != nil {
					commonMu.Lock()
					common.update(*codemodCommon)
					commonMu.Unlock()
				}
			})
		}
	}

	// Wait for required searches.
	requiredWg.Wait()

	// Give optional searches some minimum budget in case required searches return quickly.
	// Cancel all remaining searches after this minimum budget.
	budget := 100 * time.Millisecond
	elapsed := time.Since(start)
	timer := time.AfterFunc(budget-elapsed, cancel)

	// Wait for remaining optional searches to finish or get cancelled.
	optionalWg.Wait()

	timer.Stop()

	tr.LazyPrintf("results=%d limitHit=%v cloning=%d missing=%d timedout=%d", len(results), common.limitHit, len(common.cloning), len(common.missing), len(common.timedout))

	multiErr, newAlert := alertForStructuralSearch(multiErr)
	if newAlert != nil {
		alert = newAlert // takes higher precedence
	}

	if len(missingRepoRevs) > 0 {
		alert = alertForMissingRepoRevs(r.patternType, missingRepoRevs)
	}

	if len(results) == 0 && strings.Contains(r.originalQuery, `"`) && r.patternType == query.SearchTypeLiteral {
		alert = alertForQuotesInQueryInLiteralMode(r.parseTree)
	}

	// If we have some results, only log the error instead of returning it,
	// because otherwise the client would not receive the partial results
	if len(results) > 0 && multiErr != nil {
		log15.Error("Errors during search", "error", multiErr)
		multiErr = nil
	}

	sortResults(results)

	resultsResolver := SearchResultsResolver{
		start:               start,
		searchResultsCommon: common,
		SearchResults:       results,
		alert:               alert,
	}

	return &resultsResolver, multiErr.ErrorOrNil()
}

// isContextError returns true if ctx.Err() is not nil or if err
// is an error caused by context cancelation or timeout.
func isContextError(ctx context.Context, err error) bool {
	return ctx.Err() != nil || err == context.Canceled || err == context.DeadlineExceeded
}

// SearchResultResolver is a resolver for the GraphQL union type `SearchResult`.
//
// Supported types:
//
//   - *RepositoryResolver         // repo name match
//   - *fileMatchResolver          // text match
//   - *commitSearchResultResolver // diff or commit match
//   - *codemodResultResolver      // code modification
//
// Note: Any new result types added here also need to be handled properly in search_results.go:301 (sparklines)
type SearchResultResolver interface {
	ToRepository() (*RepositoryResolver, bool)
	ToFileMatch() (*FileMatchResolver, bool)
	ToCommitSearchResult() (*commitSearchResultResolver, bool)
	ToCodemodResult() (*codemodResultResolver, bool)

	// SearchResultURIs returns the repo name and file uri respectiveley
	searchResultURIs() (string, string)
	resultCount() int32
}

// compareSearchResults checks to see if a is less than b.
// It is implemented separately for easier testing.
func compareSearchResults(a, b SearchResultResolver) bool {
	arepo, afile := a.searchResultURIs()
	brepo, bfile := b.searchResultURIs()

	if arepo == brepo {
		return afile < bfile
	}

	return arepo < brepo
}

func sortResults(r []SearchResultResolver) {
	sort.Slice(r, func(i, j int) bool { return compareSearchResults(r[i], r[j]) })
}

// orderedFuzzyRegexp interpolate a lazy 'match everything' regexp pattern
// to achieve an ordered fuzzy regexp match.
func orderedFuzzyRegexp(pieces []string) string {
	if len(pieces) == 0 {
		return ""
	}
	if len(pieces) == 1 {
		return pieces[0]
	}
	return "(" + strings.Join(pieces, ").*?(") + ")"
}

// Validates usage of the `repohasfile` filter
func validateRepoHasFileUsage(q *query.Query) error {
	// Query only contains "repohasfile:" and "type:symbol"
	if len(q.Fields) == 2 && q.Fields["repohasfile"] != nil && q.Fields["type"] != nil && len(q.Fields["type"]) == 1 && q.Fields["type"][0].Value() == "symbol" {
		return errors.New("repohasfile does not currently return symbol results. Support for symbol results is coming soon. Subscribe to https://github.com/sourcegraph/sourcegraph/issues/4610 for updates")
	}
	return nil
}
