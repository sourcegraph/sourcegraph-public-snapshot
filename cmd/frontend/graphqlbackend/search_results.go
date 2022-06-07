package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/neelance/parallel"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	searchlogs "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/logs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	searchhoney "github.com/sourcegraph/sourcegraph/internal/honey/search"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/execute"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SearchResultsResolver is a resolver for the GraphQL type `SearchResults`
type SearchResultsResolver struct {
	db          database.DB
	Matches     result.Matches
	Stats       streaming.Stats
	SearchAlert *search.Alert

	// The time it took to compute all results.
	elapsed time.Duration
}

func (c *SearchResultsResolver) LimitHit() bool {
	return c.Stats.IsLimitHit
}

func (c *SearchResultsResolver) matchesRepoIDs() map[api.RepoID]struct{} {
	m := map[api.RepoID]struct{}{}
	for _, id := range c.Matches {
		m[id.RepoName().ID] = struct{}{}
	}
	return m
}

func (c *SearchResultsResolver) Repositories(ctx context.Context) ([]*RepositoryResolver, error) {
	// c.Stats.Repos does not necessarily respect limits that are applied in
	// our graphql layers. Instead we generate the list from the matches.
	m := c.matchesRepoIDs()
	ids := make([]api.RepoID, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	return c.repositoryResolvers(ctx, ids)
}

func (c *SearchResultsResolver) RepositoriesCount() int32 {
	return int32(len(c.matchesRepoIDs()))
}

func (c *SearchResultsResolver) repositoryResolvers(ctx context.Context, ids []api.RepoID) ([]*RepositoryResolver, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	resolvers := make([]*RepositoryResolver, 0, len(ids))
	err := c.db.Repos().StreamMinimalRepos(ctx, database.ReposListOptions{
		IDs: ids,
	}, func(repo *types.MinimalRepo) {
		resolvers = append(resolvers, NewRepositoryResolver(c.db, repo.ToRepo()))
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(resolvers, func(a, b int) bool {
		return resolvers[a].ID() < resolvers[b].ID()
	})
	return resolvers, nil
}

func (c *SearchResultsResolver) repoIDsByStatus(mask search.RepoStatus) []api.RepoID {
	var ids []api.RepoID
	c.Stats.Status.Filter(mask, func(id api.RepoID) {
		ids = append(ids, id)
	})
	return ids
}

func (c *SearchResultsResolver) Cloning(ctx context.Context) ([]*RepositoryResolver, error) {
	return c.repositoryResolvers(ctx, c.repoIDsByStatus(search.RepoStatusCloning))
}

func (c *SearchResultsResolver) Missing(ctx context.Context) ([]*RepositoryResolver, error) {
	return c.repositoryResolvers(ctx, c.repoIDsByStatus(search.RepoStatusMissing))
}

func (c *SearchResultsResolver) Timedout(ctx context.Context) ([]*RepositoryResolver, error) {
	return c.repositoryResolvers(ctx, c.repoIDsByStatus(search.RepoStatusTimedout))
}

func (c *SearchResultsResolver) IndexUnavailable() bool {
	// This used to return c.Stats.IsIndexUnavailable, but it was never set,
	// so would always return false
	return false
}

// Results are the results found by the search. It respects the limits set. To
// access all results directly access the SearchResults field.
func (sr *SearchResultsResolver) Results() []SearchResultResolver {
	return matchesToResolvers(sr.db, sr.Matches)
}

func matchesToResolvers(db database.DB, matches []result.Match) []SearchResultResolver {
	type repoKey struct {
		Name types.MinimalRepo
		Rev  string
	}
	repoResolvers := make(map[repoKey]*RepositoryResolver, 10)
	getRepoResolver := func(repoName types.MinimalRepo, rev string) *RepositoryResolver {
		if existing, ok := repoResolvers[repoKey{repoName, rev}]; ok {
			return existing
		}
		resolver := NewRepositoryResolver(db, repoName.ToRepo())
		resolver.RepoMatch.Rev = rev
		repoResolvers[repoKey{repoName, rev}] = resolver
		return resolver
	}

	resolvers := make([]SearchResultResolver, 0, len(matches))
	for _, match := range matches {
		switch v := match.(type) {
		case *result.FileMatch:
			resolvers = append(resolvers, &FileMatchResolver{
				db:           db,
				FileMatch:    *v,
				RepoResolver: getRepoResolver(v.Repo, ""),
			})
		case *result.RepoMatch:
			resolvers = append(resolvers, getRepoResolver(v.RepoName(), v.Rev))
		case *result.CommitMatch:
			resolvers = append(resolvers, &CommitSearchResultResolver{
				db:          db,
				CommitMatch: *v,
			})
		}
	}
	return resolvers
}

func (sr *SearchResultsResolver) MatchCount() int32 {
	return int32(sr.Matches.ResultCount())
}

// Deprecated. Prefer MatchCount.
func (sr *SearchResultsResolver) ResultCount() int32 { return sr.MatchCount() }

func (sr *SearchResultsResolver) ApproximateResultCount() string {
	count := sr.MatchCount()
	if sr.LimitHit() || sr.Stats.Status.Any(search.RepoStatusCloning|search.RepoStatusTimedout) {
		return fmt.Sprintf("%d+", count)
	}
	return strconv.Itoa(int(count))
}

func (sr *SearchResultsResolver) Alert() *searchAlertResolver {
	return NewSearchAlertResolver(sr.SearchAlert)
}

func (sr *SearchResultsResolver) ElapsedMilliseconds() int32 {
	return int32(sr.elapsed.Milliseconds())
}

func (sr *SearchResultsResolver) DynamicFilters(ctx context.Context) []*searchFilterResolver {
	tr, _ := trace.New(ctx, "DynamicFilters", "", trace.Tag{Key: "resolver", Value: "SearchResultsResolver"})
	defer tr.Finish()

	var filters streaming.SearchFilters
	filters.Update(streaming.SearchEvent{
		Results: sr.Matches,
		Stats:   sr.Stats,
	})

	var resolvers []*searchFilterResolver
	for _, f := range filters.Compute() {
		resolvers = append(resolvers, &searchFilterResolver{filter: *f})
	}
	return resolvers
}

type searchFilterResolver struct {
	filter streaming.Filter
}

func (sf *searchFilterResolver) Value() string {
	return sf.filter.Value
}

func (sf *searchFilterResolver) Label() string {
	return sf.filter.Label
}

func (sf *searchFilterResolver) Count() int32 {
	return int32(sf.filter.Count)
}

func (sf *searchFilterResolver) LimitHit() bool {
	return sf.filter.IsLimitHit
}

func (sf *searchFilterResolver) Kind() string {
	return sf.filter.Kind
}

// blameFileMatch blames the specified file match to produce the time at which
// the first line match inside of it was authored.
func (sr *SearchResultsResolver) blameFileMatch(ctx context.Context, fm *result.FileMatch) (t time.Time, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "blameFileMatch")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Blame the first line match.
	if len(fm.ChunkMatches) == 0 {
		// No line match
		return time.Time{}, nil
	}
	hm := fm.ChunkMatches[0]
	hunks, err := gitserver.NewClient(sr.db).BlameFile(ctx, fm.Repo.Name, fm.Path, &gitserver.BlameOptions{
		NewestCommit: fm.CommitID,
		StartLine:    hm.Ranges[0].Start.Line,
		EndLine:      hm.Ranges[0].Start.Line,
	}, authz.DefaultSubRepoPermsChecker)
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
	for _, r := range sr.Matches {
		r := r // shadow so it doesn't change in the goroutine
		switch m := r.(type) {
		case *result.RepoMatch:
			// We don't care about repo results here.
			continue
		case *result.CommitMatch:
			// Diff searches are cheap, because we implicitly have author date info.
			addPoint(m.Commit.Author.Date)
		case *result.FileMatch:
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
		default:
			panic("SearchResults.Sparkline unexpected union type state")
		}
	}
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("blame_ops", blameOps)
	return sparkline, nil
}

var (
	searchResponseCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_graphql_search_response",
		Help: "Number of searches that have ended in the given status (success, error, timeout, partial_timeout).",
	}, []string{"status", "alert_type", "source", "request_name"})

	searchLatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_search_response_latency_seconds",
		Help:    "Search response latencies in seconds that have ended in the given status (success, error, timeout, partial_timeout).",
		Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 20, 30},
	}, []string{"status", "alert_type", "source", "request_name"})
)

// LogSearchLatency records search durations in the event database. This
// function may only be called after a search result is performed, because it
// relies on the invariant that query and pattern error checking has already
// been performed.
func LogSearchLatency(ctx context.Context, db database.DB, wg *sync.WaitGroup, si *run.SearchInputs, durationMs int32) {
	tr, ctx := trace.New(ctx, "LogSearchLatency", "")
	defer tr.Finish()
	var types []string
	resultTypes, _ := si.Query.StringValues(query.FieldType)
	for _, typ := range resultTypes {
		switch typ {
		case "repo", "symbol", "diff", "commit":
			types = append(types, typ)
		case "path":
			// Map type:path to file
			types = append(types, "file")
		case "file":
			switch {
			case si.PatternType == query.SearchTypeStructural:
				types = append(types, "structural")
			case si.PatternType == query.SearchTypeLiteralDefault:
				types = append(types, "literal")
			case si.PatternType == query.SearchTypeRegex:
				types = append(types, "regexp")
			case si.PatternType == query.SearchTypeLucky:
				types = append(types, "lucky")
			}
		}
	}

	// Don't record composite searches that specify more than one type:
	// because we can't break down the search timings into multiple
	// categories.
	if len(types) > 1 {
		return
	}

	q, err := query.ToBasicQuery(si.Query)
	if err != nil {
		// Can't convert to a basic query, can't guarantee accurate reporting.
		return
	}
	if !query.IsPatternAtom(q) {
		// Not an atomic pattern, can't guarantee accurate reporting.
		return
	}

	// If no type: was explicitly specified, infer the result type.
	if len(types) == 0 {
		// If a pattern was specified, a content search happened.
		if q.IsLiteral() {
			types = append(types, "literal")
		} else if q.IsRegexp() {
			types = append(types, "regexp")
		} else if q.IsStructural() {
			types = append(types, "structural")
		} else if si.Query.Exists(query.FieldFile) {
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
		a := actor.FromContext(ctx)
		if a.IsAuthenticated() && !a.IsMockUser() { // Do not log in tests
			value := fmt.Sprintf(`{"durationMs": %d}`, durationMs)
			eventName := fmt.Sprintf("search.latencies.%s", types[0])
			featureFlags := featureflag.FromContext(ctx)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := usagestats.LogBackendEvent(db, a.UID, deviceid.FromContext(ctx), eventName, json.RawMessage(value), json.RawMessage(value), featureFlags, nil)
				if err != nil {
					log15.Warn("Could not log search latency", "err", err)
				}
			}()
		}
	}
}

func (r *searchResolver) JobClients() job.RuntimeClients {
	return job.RuntimeClients{
		DB:           r.db,
		Zoekt:        r.zoekt,
		SearcherURLs: r.searcherURLs,
		Gitserver:    gitserver.NewClient(r.db),
	}
}

func logPrometheusBatch(status, alertType, requestSource, requestName string, elapsed time.Duration) {
	searchResponseCounter.WithLabelValues(
		status,
		alertType,
		requestSource,
		requestName,
	).Inc()

	searchLatencyHistogram.WithLabelValues(
		status,
		alertType,
		requestSource,
		requestName,
	).Observe(elapsed.Seconds())
}

func logBatch(ctx context.Context, db database.DB, searchInputs *run.SearchInputs, srr *SearchResultsResolver, err error) {
	var wg sync.WaitGroup
	LogSearchLatency(ctx, db, &wg, searchInputs, srr.ElapsedMilliseconds())
	defer wg.Wait()

	var status, alertType string
	status = DetermineStatusForLogs(srr.SearchAlert, srr.Stats, err)
	if srr.SearchAlert != nil {
		alertType = srr.SearchAlert.PrometheusType
	}
	requestSource := string(trace.RequestSource(ctx))
	requestName := trace.GraphQLRequestName(ctx)
	logPrometheusBatch(status, alertType, requestSource, requestName, srr.elapsed)

	isSlow := srr.elapsed > searchlogs.LogSlowSearchesThreshold()
	if honey.Enabled() || isSlow {
		var n int
		if srr != nil {
			n = len(srr.Matches)
		}
		ev := searchhoney.SearchEvent(ctx, searchhoney.SearchEventArgs{
			OriginalQuery: searchInputs.OriginalQuery,
			Typ:           requestName,
			Source:        requestSource,
			Status:        status,
			AlertType:     alertType,
			DurationMs:    srr.elapsed.Milliseconds(),
			ResultSize:    n,
			Error:         err,
		})

		_ = ev.Send()

		if isSlow {
			log15.Warn("slow search request", searchlogs.MapToLog15Ctx(ev.Fields())...)
		}
	}
}

func (r *searchResolver) Results(ctx context.Context) (*SearchResultsResolver, error) {
	start := time.Now()
	agg := streaming.NewAggregatingStream()
	alert, err := execute.Execute(ctx, agg, r.SearchInputs, r.JobClients())
	srr := r.resultsToResolver(agg.Results, alert, agg.Stats)
	srr.elapsed = time.Since(start)
	logBatch(ctx, r.db, r.SearchInputs, srr, err)
	return srr, err
}

func (r *searchResolver) resultsToResolver(matches result.Matches, alert *search.Alert, stats streaming.Stats) *SearchResultsResolver {
	return &SearchResultsResolver{
		Matches:     matches,
		SearchAlert: alert,
		Stats:       stats,
		db:          r.db,
	}
}

// DetermineStatusForLogs determines the final status of a search for logging
// purposes.
func DetermineStatusForLogs(alert *search.Alert, stats streaming.Stats, err error) string {
	switch {
	case err == context.DeadlineExceeded:
		return "timeout"
	case err != nil:
		return "error"
	case stats.Status.All(search.RepoStatusTimedout) && stats.Status.Len() == len(stats.Repos):
		return "timeout"
	case stats.Status.Any(search.RepoStatusTimedout):
		return "partial_timeout"
	case alert != nil:
		return "alert"
	default:
		return "success"
	}
}

type searchResultsStats struct {
	JApproximateResultCount string
	JSparkline              []int32

	sr *searchResolver

	// These items are lazily populated by getResults
	once    sync.Once
	results result.Matches
	err     error
}

func (srs *searchResultsStats) ApproximateResultCount() string { return srs.JApproximateResultCount }
func (srs *searchResultsStats) Sparkline() []int32             { return srs.JSparkline }

var (
	searchResultsStatsCache   = rcache.NewWithTTL("search_results_stats", 3600) // 1h
	searchResultsStatsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_graphql_search_results_stats_cache_hit",
		Help: "Counts cache hits and misses for search results stats (e.g. sparklines).",
	}, []string{"type"})
)

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

	cacheKey := r.SearchInputs.OriginalQuery
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
		b, err := query.ToBasicQuery(r.SearchInputs.Query)
		if err != nil {
			return nil, err
		}
		j, err := jobutil.NewBasicJob(r.SearchInputs, b)
		if err != nil {
			return nil, err
		}
		agg := streaming.NewAggregatingStream()
		alert, err := j.Run(ctx, r.JobClients(), agg)
		if err != nil {
			return nil, err // do not cache errors.
		}
		v = r.resultsToResolver(agg.Results, alert, agg.Stats)
		if v.MatchCount() > 0 {
			break
		}

		status := v.Stats.Status
		if !status.Any(search.RepoStatusCloning) && !status.Any(search.RepoStatusTimedout) {
			break // zero results, but no cloning or timed out repos. No point in retrying.
		}

		var cloning, timedout int
		status.Filter(search.RepoStatusCloning, func(api.RepoID) {
			cloning++
		})
		status.Filter(search.RepoStatusTimedout, func(api.RepoID) {
			timedout++
		})

		if attempts > 5 {
			log15.Error("failed to generate sparkline due to cloning or timed out repos", "cloning", cloning, "timedout", timedout)
			return nil, errors.Errorf("failed to generate sparkline due to %d cloning %d timedout repos", cloning, timedout)
		}

		// We didn't find any search results. Some repos are cloning or timed
		// out, so try again in a few seconds.
		attempts++
		log15.Warn("sparkline generation found 0 search results due to cloning or timed out repos (retrying in 5s)", "cloning", cloning, "timedout", timedout)
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

// SearchResultResolver is a resolver for the GraphQL union type `SearchResult`.
//
// Supported types:
//
//   - *RepositoryResolver         // repo name match
//   - *fileMatchResolver          // text match
//   - *commitSearchResultResolver // diff or commit match
//
// Note: Any new result types added here also need to be handled properly in search_results.go:301 (sparklines)
type SearchResultResolver interface {
	ToRepository() (*RepositoryResolver, bool)
	ToFileMatch() (*FileMatchResolver, bool)
	ToCommitSearchResult() (*CommitSearchResultResolver, bool)
}
