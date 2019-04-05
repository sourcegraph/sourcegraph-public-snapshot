package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/neelance/parallel"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory/filelang"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

// searchResultsCommon contains fields that should be returned by all funcs
// that contribute to the overall search result set.
type searchResultsCommon struct {
	limitHit bool                      // whether the limit on results was hit
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

func (c *searchResultsCommon) Repositories() []*repositoryResolver {
	if c.repos == nil {
		return []*repositoryResolver{}
	}
	return toRepositoryResolvers(c.repos)
}

func (c *searchResultsCommon) RepositoriesSearched() []*repositoryResolver {
	if c.searched == nil {
		return nil
	}
	return toRepositoryResolvers(c.searched)
}

func (c *searchResultsCommon) IndexedRepositoriesSearched() []*repositoryResolver {
	if c.indexed == nil {
		return nil
	}
	return toRepositoryResolvers(c.indexed)
}

func (c *searchResultsCommon) Cloning() []*repositoryResolver {
	if c.cloning == nil {
		return nil
	}
	return toRepositoryResolvers(c.cloning)
}

func (c *searchResultsCommon) Missing() []*repositoryResolver {
	if c.missing == nil {
		return nil
	}
	return toRepositoryResolvers(c.missing)
}

func (c *searchResultsCommon) Timedout() []*repositoryResolver {
	if c.timedout == nil {
		return nil
	}
	return toRepositoryResolvers(c.timedout)
}

func (c *searchResultsCommon) IndexUnavailable() bool {
	return c.indexUnavailable
}

// update updates c with the other data, deduping as necessary. It modifies c but
// does not modify other.
func (c *searchResultsCommon) update(other searchResultsCommon) {
	c.limitHit = c.limitHit || other.limitHit
	c.indexUnavailable = c.indexUnavailable || other.indexUnavailable

	appendUnique := func(dstp *[]*types.Repo, src []*types.Repo) {
		dst := *dstp
		sort.Slice(dst, func(i, j int) bool { return dst[i].ID < dst[j].ID })
		sort.Slice(src, func(i, j int) bool { return src[i].ID < src[j].ID })
		for _, r := range dst {
			for len(src) > 0 && src[0].ID <= r.ID {
				if r != src[0] {
					dst = append(dst, src[0])
				}
				src = src[1:]
			}
		}
		dst = append(dst, src...)
		sort.Slice(dst, func(i, j int) bool { return dst[i].ID < dst[j].ID })
		*dstp = dst
	}
	appendUnique(&c.repos, other.repos)
	appendUnique(&c.searched, other.searched)
	appendUnique(&c.indexed, other.indexed)
	appendUnique(&c.cloning, other.cloning)
	appendUnique(&c.missing, other.missing)
	appendUnique(&c.timedout, other.timedout)
	c.resultCount += other.resultCount

	if c.partial == nil {
		c.partial = make(map[api.RepoName]struct{})
	}

	for repo := range other.partial {
		c.partial[repo] = struct{}{}
	}
}

// searchResultsResolver is a resolver for the GraphQL type `SearchResults`
type searchResultsResolver struct {
	results []*searchResultResolver
	searchResultsCommon
	alert *searchAlert
	start time.Time // when the results started being computed
}

func (sr *searchResultsResolver) Results() []*searchResultResolver {
	return sr.results
}

func (sr *searchResultsResolver) ResultCount() int32 {
	var totalResults int32
	for _, result := range sr.results {
		totalResults += result.resultCount()
	}
	return totalResults
}

func (sr *searchResultsResolver) ApproximateResultCount() string {
	count := sr.ResultCount()
	if sr.LimitHit() || len(sr.cloning) > 0 || len(sr.timedout) > 0 {
		return fmt.Sprintf("%d+", count)
	}
	return strconv.Itoa(int(count))
}

func (sr *searchResultsResolver) Alert() *searchAlert { return sr.alert }

func (sr *searchResultsResolver) ElapsedMilliseconds() int32 {
	return int32(time.Since(sr.start).Nanoseconds() / int64(time.Millisecond))
}

// commonFileFilters are common filters used. It is used by DynamicFilters to
// propose them if they match shown results.
var commonFileFilters = []struct {
	Regexp *regexp.Regexp
	Filter string
}{
	// Exclude go tests
	{
		Regexp: regexp.MustCompile(`_test\.go$`),
		Filter: `-file:_test\.go$`,
	},
	// Exclude go vendor
	{
		Regexp: regexp.MustCompile(`(^|/)vendor/`),
		Filter: `-file:(^|/)vendor/`,
	},
	// Exclude node_modules
	{
		Regexp: regexp.MustCompile(`(^|/)node_modules/`),
		Filter: `-file:(^|/)node_modules/`,
	},
}

func (sr *searchResultsResolver) DynamicFilters() []*searchFilterResolver {
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
		extensionToLanguageLookup := func(ext string) string {
			for _, lang := range filelang.Langs {
				for _, langExt := range lang.Extensions {
					if ext == langExt {
						return strings.ToLower(lang.Name)
					}
				}
			}
			return ""
		}
		if ext := path.Ext(fileMatchPath); ext != "" {
			language := extensionToLanguageLookup(path.Ext(fileMatchPath))
			if language != "" {
				value := fmt.Sprintf(`lang:%s`, language)
				add(value, value, lineMatchCount, limitHit, "lang")
			}
		}
	}

	for _, result := range sr.results {
		if result.fileMatch != nil {
			rev := ""
			if result.fileMatch.inputRev != nil {
				rev = *result.fileMatch.inputRev
			}
			addRepoFilter(string(result.fileMatch.repo.Name), rev, len(result.fileMatch.LineMatches()))
			addLangFilter(result.fileMatch.JPath, len(result.fileMatch.LineMatches()), result.fileMatch.JLimitHit)
			addFileFilter(result.fileMatch.JPath, len(result.fileMatch.LineMatches()), result.fileMatch.JLimitHit)

			if len(result.fileMatch.symbols) > 0 {
				add("type:symbol", "type:symbol", 1, result.fileMatch.JLimitHit, "symbol")
			}
		}

		if result.repo != nil {
			// It should be fine to leave this blank since revision specifiers
			// can only be used with the 'repo:' scope. In that case,
			// we shouldn't be getting any repositoy name matches back.
			addRepoFilter(result.repo.URI(), "", 1)
		}
		// Add `case:yes` filter to offer easier access to search results matching with case sensitive set to yes
		// We use count == 0 and limitHit == false since we can't determine that information without
		// running the search query. This causes it to display as just `case:yes`.
		add("case:yes", "case:yes", 0, false, "case")
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
func (sr *searchResultsResolver) blameFileMatch(ctx context.Context, fm *fileMatchResolver) (t time.Time, err error) {
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
	hunks, err := git.BlameFile(ctx, gitserver.Repo{Name: fm.repo.Name}, fm.JPath, &git.BlameOptions{
		NewestCommit: fm.commitID,
		StartLine:    int(lm.LineNumber()),
		EndLine:      int(lm.LineNumber()),
	})
	if err != nil {
		return time.Time{}, err
	}

	return hunks[0].Author.Date, nil
}

func (sr *searchResultsResolver) Sparkline(ctx context.Context) (sparkline []int32, err error) {
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
	for _, r := range sr.results {
		r := r // shadow so it doesn't change in the goroutine
		switch {
		case r.repo != nil:
			// We don't care about repo results here.
			continue
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
			goroutine.Go(func() {
				defer run.Release()

				// Blame the file match in order to retrieve date informatino.
				var err error
				t, err := sr.blameFileMatch(ctx, r.fileMatch)
				if err != nil {
					log15.Warn("failed to blame fileMatch during sparkline generation", "error", err)
					return
				}
				addPoint(t)
			})
		default:
			panic("SearchResults.Sparkline unexpected union type state")
		}
	}
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("blame_ops", blameOps)
	return sparkline, nil
}

func (r *searchResolver) Results(ctx context.Context) (*searchResultsResolver, error) {
	start := time.Now()
	rr, err := r.doResults(ctx, "")
	if err != nil {
		log15.Debug("graphql search failed", "query", r.rawQuery(), "duration", time.Since(start), "error", err)
		return nil, err
	}
	log15.Debug("graphql search success", "query", r.rawQuery(), "count", rr.ResultCount(), "duration", time.Since(start))
	return rr, nil
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
	var v *searchResultsResolver
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

type getPatternInfoOptions struct {
	// forceFileSearch, when true, specifies that the search query should be
	// treated as if every default term had `file:` before it. This can be used
	// to allow users to jump to files by just typing their name.
	forceFileSearch bool
}

// getPatternInfo gets the search pattern info for the query in the resolver.
func (r *searchResolver) getPatternInfo(opts *getPatternInfoOptions) (*search.PatternInfo, error) {
	var patternsToCombine []string
	if opts == nil || !opts.forceFileSearch {
		for _, v := range r.query.Values(query.FieldDefault) {
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
	} else {
		// TODO: We must have some pattern that always matches here, or else
		// cmd/searcher/search/matcher.go:97 would cause a nil regexp panic
		// when not using indexed search. I am unsure what the right solution
		// is here. Would this code path go away when we switch fully to
		// indexed search @keegan? This workaround is OK for now though.
		patternsToCombine = append(patternsToCombine, ".")
	}

	// Handle file: and -file: filters.
	includePatterns, excludePatterns := r.query.RegexpPatterns(query.FieldFile)

	if opts != nil && opts.forceFileSearch {
		for _, v := range r.query.Values(query.FieldDefault) {
			includePatterns = append(includePatterns, asString(v))
		}
	}

	// Handle lang: and -lang: filters.
	langIncludePatterns, langExcludePatterns, err := langIncludeExcludePatterns(r.query.StringValues(query.FieldLang))
	if err != nil {
		return nil, err
	}
	includePatterns = append(includePatterns, langIncludePatterns...)
	excludePatterns = append(excludePatterns, langExcludePatterns...)

	patternInfo := &search.PatternInfo{
		IsRegExp:                     true,
		IsCaseSensitive:              r.query.IsCaseSensitive(),
		FileMatchLimit:               r.maxResults(),
		Pattern:                      regexpPatternMatchingExprsInOrder(patternsToCombine),
		IncludePatterns:              includePatterns,
		PathPatternsAreRegExps:       true,
		PathPatternsAreCaseSensitive: r.query.IsCaseSensitive(),
	}
	if len(excludePatterns) > 0 {
		patternInfo.ExcludePattern = unionRegExps(excludePatterns)
	}
	return patternInfo, nil
}

var (
	// The default timeout to use for queries.
	defaultTimeout = 10 * time.Second
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
		// If `count:` is set but `timeout:` is not explicitely set, use the max timeout
		d = maxTimeout
	}
	// don't run queries longer than 1 minute.
	if d.Minutes() > 1 {
		d = maxTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, d)
	return ctx, cancel, nil
}

func (r *searchResolver) doResults(ctx context.Context, forceOnlyResultType string) (res *searchResultsResolver, err error) {
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
		return &searchResultsResolver{alert: alert, start: start}, nil
	}
	if overLimit {
		alert, err := r.alertForOverRepoLimit(ctx)
		if err != nil {
			return nil, err
		}
		return &searchResultsResolver{alert: alert, start: start}, nil
	}

	p, err := r.getPatternInfo(nil)
	if err != nil {
		return nil, err
	}
	args := search.Args{
		Pattern:         p,
		Repos:           repos,
		Query:           r.query,
		UseFullDeadline: r.searchTimeoutFieldSet(),
	}
	if err := args.Pattern.Validate(); err != nil {
		return nil, &badRequestError{err}
	}

	// Determine which types of results to return.
	var rawResultTypes []string
	if forceOnlyResultType != "" {
		rawResultTypes = []string{forceOnlyResultType}
	} else {
		rawResultTypes, _ = r.query.StringValues(query.FieldType)
		if len(rawResultTypes) == 0 {
			rawResultTypes = []string{"file", "path", "repo", "ref"}
		}
	}
	resultTypes := make(map[string]struct{}, len(rawResultTypes)) // deduplicated
	for _, resultType := range rawResultTypes {
		resultTypes[resultType] = struct{}{}
		if resultType == "file" {
			args.Pattern.PatternMatchesContent = true
		} else if resultType == "path" {
			args.Pattern.PatternMatchesPath = true
		}
	}
	tr.LazyPrintf("resultTypes: %v", rawResultTypes)

	// A mapping of result types to functions that search for those result types.
	searchers := map[string]struct {
		search func(ctx context.Context, args *search.Args, limit int32) ([]*searchResultResolver, *searchResultsCommon, error)

		// When true, these results can optionally be left out of the response
		// if they are not retrieved fast enough AND the user did not
		// explicitly ask for this type of result.
		optional bool
	}{
		"repo":   {searchRepositories, false},
		"symbol": {searchSymbols, true},
		"file":   {searchFilesInRepos, false},
		"path":   {searchFilesInRepos, false},
		"diff":   {searchCommitDiffsInRepos, true},
		"commit": {searchCommitLogInRepos, true},
	}

	// Start searchers in parallel, only for the types of results the user wants.
	type searcherResult struct {
		resultType string
		results    []*searchResultResolver
		common     *searchResultsCommon
		err        error
	}
	var (
		expectRequired, expectOptional int // How many required and optional results we can expect from the channel.
		required                       = make(chan searcherResult, len(resultTypes))
		optional                       = make(chan searcherResult, len(resultTypes))
	)
	for resultType := range resultTypes {
		resultType := resultType // shadow variable for goroutine below
		if resultType == "file" {
			if _, ok := resultTypes["path"]; ok {
				// We are searching for both file and path results. They are provided by the
				// same searchFilesInRepos, so we would just end up with duplicate results if
				// we called them 2x.
				continue
			}
		}
		searcher, ok := searchers[resultType]
		if !ok {
			// We don't have anything to handle this type of query.
			continue
		}
		isOptional := searcher.optional
		if args.UseFullDeadline {
			// When a custom timeout is specified, all searches are required.
			isOptional = false
		}
		if isOptional {
			expectOptional++
		} else {
			expectRequired++
		}
		goroutine.Go(func() {
			results, common, err := searcher.search(ctx, &args, r.maxResults())
			if isOptional {
				optional <- searcherResult{resultType, results, common, err}
			} else {
				required <- searcherResult{resultType, results, common, err}
			}
		})
	}

	// Gather results from the workers. We do so in no particular order, as
	// we will sort them later.
	var searcherResults []searcherResult
	for i := 0; i < expectRequired; i++ {
		searcherResults = append(searcherResults, <-required)
	}
	// Optional search results only have 100ms (since the start of the search
	// query) to return if we already got required search results. This is to
	// prevent optional search results (which are often slower) from slowing
	// down the overall search.
	budget := 100 * time.Millisecond
	elapsed := time.Since(start)
	timer := time.AfterFunc(budget-elapsed, cancel)
optionalSearches:
	for i := 0; i < expectOptional; i++ {
		select {
		case r := <-optional:
			searcherResults = append(searcherResults, r)
		case <-timer.C:
			break optionalSearches
		}
	}
	timer.Stop()

	// Handle the results from the workers, again in no particular order as we
	// will sort later.
	var (
		results  []*searchResultResolver
		common   = searchResultsCommon{maxResultsCount: r.maxResults()}
		multiErr *multierror.Error

		// fileMatches is a map from git:// URI of the file to FileMatch resolver
		// to merge multiple results of different types for the same file
		fileMatches = make(map[string]*fileMatchResolver)
	)
	for _, sr := range searcherResults {
		// Timeouts are reported through searchResultsCommon so don't report an
		// error for them (context errors).
		if sr.err != nil && !isContextError(ctx, sr.err) {
			multiErr = multierror.Append(multiErr, errors.Wrap(sr.err, sr.resultType+" search failed"))
		}
		if sr.common != nil {
			common.update(*sr.common)
		}
		for _, result := range sr.results {
			if result.fileMatch != nil {
				// When we encounter file match results, we deduplicate them as the
				// "symbol", "file", and "path" searchers registered above can produce
				// duplicates as they aren't aware of eachother.
				key := result.fileMatch.uri
				if m, ok := fileMatches[key]; ok {
					// Result for this file already exists, so merge it, so that we get
					// the best result possible (i.e. because one searcher may produce
					// nil line matches).
					m.JLimitHit = m.JLimitHit || result.fileMatch.JLimitHit
					if m.JLineMatches == nil {
						m.JLineMatches = result.fileMatch.JLineMatches
					}
				} else {
					fileMatches[key] = result.fileMatch
					results = append(results, result)
				}
				continue
			}
			results = append(results, result)
		}
	}

	tr.LazyPrintf("results=%d limitHit=%v cloning=%d missing=%d timedout=%d", len(results), common.limitHit, len(common.cloning), len(common.missing), len(common.timedout))

	// Alert is a potential alert shown to the user.
	var alert *searchAlert

	if len(missingRepoRevs) > 0 {
		alert = r.alertForMissingRepoRevs(missingRepoRevs)
	}

	// If we have some results, only log the error instead of returning it,
	// because otherwise the client would not receive the partial results
	if len(results) > 0 && multiErr != nil {
		log15.Error("Errors during search", "error", multiErr)
		multiErr = nil
	}

	sortResults(results)
	return &searchResultsResolver{
		start:               start,
		searchResultsCommon: common,
		results:             results,
		alert:               alert,
	}, multiErr.ErrorOrNil()
}

// isContextError returns true if ctx.Err() is not nil or if err
// is an error caused by context cancelation or timeout.
func isContextError(ctx context.Context, err error) bool {
	return ctx.Err() != nil || err == context.Canceled || err == context.DeadlineExceeded
}

// searchResultResolver is a resolver for the GraphQL union type `SearchResult`
//
// Note: Any new result types added here also need to be handled properly in search_results.go:301 (sparklines)
type searchResultResolver struct {
	repo      *repositoryResolver         // repo name match
	fileMatch *fileMatchResolver          // text match
	diff      *commitSearchResultResolver // diff or commit match
}

// getSearchResultURIs returns the repo name and file uri respectiveley
func getSearchResultURIs(c *searchResultResolver) (string, string) {
	if c.fileMatch != nil {
		return string(c.fileMatch.repo.Name), c.fileMatch.JPath
	}
	if c.repo != nil {
		return string(c.repo.repo.Name), ""
	}
	// Diffs aren't going to be returned with other types of results
	// and are already ordered in the desired order, so we'll just leave them in place.
	return "~", "~" // lexicographically last in ASCII
}

// compareSearchResults checks to see if a is less than b.
// It is implemented separately for easier testing.
func compareSearchResults(a, b *searchResultResolver) bool {
	arepo, afile := getSearchResultURIs(a)
	brepo, bfile := getSearchResultURIs(b)

	if arepo == brepo {
		return afile < bfile
	}

	return arepo < brepo

}

func sortResults(r []*searchResultResolver) {
	sort.Slice(r, func(i, j int) bool { return compareSearchResults(r[i], r[j]) })
}

func (g *searchResultResolver) ToRepository() (*repositoryResolver, bool) {
	return g.repo, g.repo != nil
}
func (g *searchResultResolver) ToFileMatch() (*fileMatchResolver, bool) {
	return g.fileMatch, g.fileMatch != nil
}
func (g *searchResultResolver) ToCommitSearchResult() (*commitSearchResultResolver, bool) {
	return g.diff, g.diff != nil
}

func (g *searchResultResolver) resultCount() int32 {
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
