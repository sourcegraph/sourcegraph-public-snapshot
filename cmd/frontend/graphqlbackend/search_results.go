package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	searchlogs "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/logs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	searchhoney "github.com/sourcegraph/sourcegraph/internal/honey/search"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SearchResultsResolver is a resolver for the GraphQL type `SearchResults`
type SearchResultsResolver struct {
	db          database.DB
	Matches     result.Matches
	Stats       streaming.Stats
	SearchAlert *search.Alert

	// limit is the maximum number of SearchResults to send back to the user.
	limit int

	// The time it took to compute all results.
	elapsed time.Duration

	// cache for user settings. Ideally this should be set just once in the code path
	// by an upstream resolver
	UserSettings *schema.Settings
}

func (c *SearchResultsResolver) LimitHit() bool {
	return c.Stats.IsLimitHit || (c.limit > 0 && len(c.Matches) > c.limit)
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

// DEPRECATED: the implementation of this was removed when the SearchResultsStats
// was removed. The method still exists so GraphQL will typecheck, but can be removed
// as soon as all frontend references to it are gone.
func (c *SearchResultsResolver) Sparkline() []int32 { return nil }

// Results are the results found by the search. It respects the limits set. To
// access all results directly access the SearchResults field.
func (sr *SearchResultsResolver) Results() []SearchResultResolver {
	limited := sr.Matches
	if sr.limit > 0 && sr.limit < len(sr.Matches) {
		limited = sr.Matches[:sr.limit]
	}

	return matchesToResolvers(sr.db, limited)
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
			case si.PatternType == query.SearchTypeLiteral:
				types = append(types, "literal")
			case si.PatternType == query.SearchTypeRegex:
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
		} else if len(si.Query.Fields()["file"]) > 0 {
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

// JobArgs converts the parts of search resolver state to values needed to create search jobs.
func (r *searchResolver) JobArgs() *job.Args {
	return &job.Args{
		SearchInputs:        r.SearchInputs,
		OnSourcegraphDotCom: envvar.SourcegraphDotComMode(),
		Zoekt:               r.zoekt,
		SearcherURLs:        r.searcherURLs,
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

func (r *searchResolver) logBatch(ctx context.Context, srr *SearchResultsResolver, err error) {
	var wg sync.WaitGroup
	LogSearchLatency(ctx, r.db, &wg, r.SearchInputs, srr.ElapsedMilliseconds())
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
			OriginalQuery: r.SearchInputs.OriginalQuery,
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

func (r *searchResolver) resultsToResolver(matches result.Matches, alert *search.Alert, stats streaming.Stats) *SearchResultsResolver {
	return &SearchResultsResolver{
		Matches:      matches,
		SearchAlert:  alert,
		Stats:        stats,
		limit:        r.SearchInputs.MaxResults(),
		db:           r.db,
		UserSettings: r.SearchInputs.UserSettings,
	}
}

func (r *searchResolver) Results(ctx context.Context) (*SearchResultsResolver, error) {
	start := time.Now()
	agg := streaming.NewAggregatingStream()
	alert, err := r.results(ctx, agg, r.SearchInputs.Plan)
	srr := r.resultsToResolver(agg.Results, alert, agg.Stats)
	srr.elapsed = time.Since(start)
	r.logBatch(ctx, srr, err)
	return srr, err
}

func (r *searchResolver) StreamResults(ctx context.Context, stream streaming.Sender) (*search.Alert, error) {
	return r.results(ctx, stream, r.SearchInputs.Plan)
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

// expandPredicates takes a query plan, and replaces any predicates with their expansion. The returned plan
// is guaranteed to be predicate-free.
func (r *searchResolver) expandPredicates(ctx context.Context, oldPlan query.Plan) (_ query.Plan, err error) {
	tr, ctx := trace.New(ctx, "expandPredicates", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		mu      sync.Mutex
		newPlan = make(query.Plan, 0, len(oldPlan))
	)
	g, ctx := errgroup.WithContext(ctx)

	for _, q := range oldPlan {
		q := q
		g.Go(func() error {
			predicatePlan, err := substitutePredicates(q, func(pred query.Predicate) (result.Matches, error) {
				plan, err := pred.Plan(q)
				if err != nil {
					return nil, err
				}

				children := make([]job.Job, 0, len(plan))
				for _, basicQuery := range plan {
					child, err := job.ToEvaluateJob(r.JobArgs(), basicQuery)
					if err != nil {
						return nil, err
					}
					children = append(children, child)
				}

				agg := streaming.NewAggregatingStream()
				_, err = r.evaluateJob(ctx, agg, job.NewOrJob(children...))
				if err != nil {
					return nil, err
				}
				return agg.Results, nil
			})
			if errors.Is(err, ErrPredicateNoResults) {
				// The predicate has no results, so neither will this basic query
				return nil
			}
			if err != nil {
				// Fail if predicate processing fails.
				return err
			}

			mu.Lock()
			defer mu.Unlock()

			if predicatePlan != nil {
				// If the predicate generated a new plan, use that
				newPlan = append(newPlan, predicatePlan...)
			} else {
				// Otherwise, just use the original basic query
				newPlan = append(newPlan, q)
			}
			return nil
		})
	}

	return newPlan, g.Wait()
}

func (r *searchResolver) results(ctx context.Context, stream streaming.Sender, plan query.Plan) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "Results", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	plan, err = r.expandPredicates(ctx, plan)
	if err != nil {
		return nil, err
	}

	planJob, err := job.FromExpandedPlan(r.JobArgs(), plan)
	if err != nil {
		return nil, err
	}

	return r.evaluateJob(ctx, stream, planJob)
}

// searchResultsToRepoNodes converts a set of search results into repository nodes
// such that they can be used to replace a repository predicate
func searchResultsToRepoNodes(matches []result.Match) ([]query.Node, error) {
	nodes := make([]query.Node, 0, len(matches))
	for _, match := range matches {
		repoMatch, ok := match.(*result.RepoMatch)
		if !ok {
			return nil, errors.Errorf("expected type %T, but got %T", &result.RepoMatch{}, match)
		}

		repoFieldValue := "^" + regexp.QuoteMeta(string(repoMatch.Name)) + "$"
		if repoMatch.Rev != "" {
			repoFieldValue += "@" + repoMatch.Rev
		}

		nodes = append(nodes, query.Parameter{
			Field: query.FieldRepo,
			Value: repoFieldValue,
		})
	}

	return nodes, nil
}

// searchResultsToFileNodes converts a set of search results into repo/file nodes so that they
// can replace a file predicate
func searchResultsToFileNodes(matches []result.Match) ([]query.Node, error) {
	nodes := make([]query.Node, 0, len(matches))
	for _, match := range matches {
		fileMatch, ok := match.(*result.FileMatch)
		if !ok {
			return nil, errors.Errorf("expected type %T, but got %T", &result.FileMatch{}, match)
		}

		repoFieldValue := "^" + regexp.QuoteMeta(string(fileMatch.Repo.Name)) + "$"
		if fileMatch.InputRev != nil {
			repoFieldValue += "@" + *fileMatch.InputRev
		}

		// We create AND nodes to match both the repo and the file at the same time so
		// we don't get files of the same name from different repositories.
		nodes = append(nodes, query.Operator{
			Kind: query.And,
			Operands: []query.Node{
				query.Parameter{
					Field: query.FieldRepo,
					Value: repoFieldValue,
				},
				query.Parameter{
					Field: query.FieldFile,
					Value: "^" + regexp.QuoteMeta(fileMatch.Path) + "$",
				},
			},
		})
	}

	return nodes, nil
}

// evaluateJob is a toplevel function that runs a search job to yield results.
// A search job represents a tree of evaluation steps. If the deadline
// is exceeded, returns a search alert with a did-you-mean link for the same
// query with a longer timeout.
func (r *searchResolver) evaluateJob(ctx context.Context, stream streaming.Sender, job job.Job) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "evaluateJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	tr.LazyPrintf("job name: %s", job.Name())

	start := time.Now()
	countingStream := streaming.NewResultCountingStream(stream)
	statsObserver := streaming.NewStatsObservingStream(countingStream)
	jobAlert, err := job.Run(ctx, r.db, statsObserver)

	ao := alert.Observer{
		Db:           r.db,
		SearchInputs: r.SearchInputs,
		HasResults:   countingStream.Count() > 0,
	}
	if err != nil {
		ao.Error(ctx, err)
	}
	observerAlert, err := ao.Done()

	// We have an alert for context timeouts and we have a progress
	// notification for timeouts. We don't want to show both, so we only show
	// it if no repos are marked as timedout. This somewhat couples us to how
	// progress notifications work, but this is the third attempt at trying to
	// fix this behaviour so we are accepting that.
	if errors.Is(err, context.DeadlineExceeded) {
		if !statsObserver.Status.Any(search.RepoStatusTimedout) {
			usedTime := time.Since(start)
			suggestTime := longer(2, usedTime)
			return search.AlertForTimeout(usedTime, suggestTime, r.SearchInputs.OriginalQuery, r.SearchInputs.PatternType), nil
		} else {
			err = nil
		}
	}

	return search.MaxPriorityAlert(jobAlert, observerAlert), err
}

// substitutePredicates replaces all the predicates in a query with their expanded form. The predicates
// are expanded using the doExpand function.
func substitutePredicates(q query.Basic, evaluate func(query.Predicate) (result.Matches, error)) (query.Plan, error) {
	var topErr error
	success := false
	newQ := query.MapParameter(q.ToParseTree(), func(field, value string, neg bool, ann query.Annotation) query.Node {
		orig := query.Parameter{
			Field:      field,
			Value:      value,
			Negated:    neg,
			Annotation: ann,
		}

		if !ann.Labels.IsSet(query.IsPredicate) {
			return orig
		}

		if topErr != nil {
			return orig
		}

		name, params := query.ParseAsPredicate(value)
		predicate := query.DefaultPredicateRegistry.Get(field, name)
		predicate.ParseParams(params)
		matches, err := evaluate(predicate)
		if err != nil {
			topErr = err
			return nil
		}

		var nodes []query.Node
		switch predicate.Field() {
		case query.FieldRepo:
			nodes, err = searchResultsToRepoNodes(matches)
			if err != nil {
				topErr = err
				return nil
			}
		case query.FieldFile:
			nodes, err = searchResultsToFileNodes(matches)
			if err != nil {
				topErr = err
				return nil
			}
		default:
			topErr = errors.Errorf("unsupported predicate result type %q", predicate.Field())
			return nil
		}

		// If no results are returned, we need to return a sentinel error rather
		// than an empty expansion because an empty expansion means "everything"
		// rather than "nothing".
		if len(nodes) == 0 {
			topErr = ErrPredicateNoResults
			return nil
		}

		// A predicate was successfully evaluated and has results.
		success = true

		// No need to return an operator for only one result
		if len(nodes) == 1 {
			return nodes[0]
		}

		return query.Operator{
			Kind:     query.Or,
			Operands: nodes,
		}
	})

	if topErr != nil || !success {
		return nil, topErr
	}
	plan, err := query.ToPlan(query.Dnf(newQ))
	if err != nil {
		return nil, err
	}
	return plan, nil
}

var ErrPredicateNoResults = errors.New("no results returned for predicate")

// longer returns a suggested longer time to wait if the given duration wasn't long enough.
func longer(n int, dt time.Duration) time.Duration {
	dt2 := func() time.Duration {
		Ndt := time.Duration(n) * dt
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

// dummySearchResultStats implements the SearchResultsStats interface so
// GraphQL type-checks, but it is not implemented by anything because the
// Stats call is deprecated.
type dummySearchResultsStats interface {
	ApproximateResultCount() string
	Sparkline() []int32
	Languages() []interface {
		Name() string
		TotalBytes() float64
		TotalLines() int32
	}
}

func (r *searchResolver) Stats(ctx context.Context) (dummySearchResultsStats, error) {
	return nil, errors.New("search stats is deprecated")
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

	ResultCount() int32
}
