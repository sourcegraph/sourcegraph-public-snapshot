package background

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xhit/go-str2duration/v2"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	itypes "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbcache"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git/gitapi"
)

// The historical enqueuer takes regular search insights like a search for `errorf` and runs them
// in the past to find how many results for that query occurred in the past. It does this using
// live/unindexed searches slowly in the background, by finding an old Git commit closest to the
// time we're interested in. See the docstring on the historicalEnqueuer struct for an explanation
// of how that works.
//
// There are some major pros/cons of the implementation as it stands today. Pros:
//
// 1. It works and is reliable.
// 2. It is pretty extensively covered by tests.
// 3. It will not harm the rest of Sourcegraph (e.g. by enqueueing too much work, running too many search queries, etc.)
//
// The cons are:
//
// 1. It's a huge glorified series of nested for loops, which makes it complex and hard to read and
//    understand. I spent two full weeks refactoring various parts of it to make it nicer, but it's
//    really challenging to structure this code in a nice way because the problem and solution we're
//    fundamentally representing here is complex. i.e., the code is complex because the problem is.
// 2. The tests are a bit complex/difficult to read. This is a symptom of the algorithmic complexity
//    at play here. I considered testing individual units of the code more aggressively, but the
//    reality is that the individual units (e.g. each for loop) is not complex - it is the aggregate
//    of them that is. If you can find a more clear way to represent this, you are smarter than me.
//
// If you're reading this and frustrated or confused, message @slimsag and I'll help you out.

// newInsightHistoricalEnqueuer returns a background goroutine which will periodically find all of the search
// insights across all user settings, and determine for which dates they do not have data and attempt
// to backfill them by enqueueing work for executing searches with `before:` and `after:` filter
// ranges.
func newInsightHistoricalEnqueuer(ctx context.Context, workerBaseStore *basestore.Store, dataSeriesStore store.DataSeriesStore, insightsStore *store.Store, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"insights_historical_enqueuer",
		metrics.WithCountHelp("Total number of insights historical enqueuer executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    "HistoricalEnqueuer.Run",
		Metrics: metrics,
	})

	defaultRateLimit := rate.Limit(20.0)
	getRateLimit := getRateLimit(defaultRateLimit)

	limiter := rate.NewLimiter(getRateLimit(), 1)

	go conf.Watch(func() {
		val := getRateLimit()
		log15.Info(fmt.Sprintf("Updating insights/historical-worker rate limit value=%v", val))
		limiter.SetLimit(val)
	})

	repoStore := database.Repos(workerBaseStore.Handle().DB())

	framesToBackfill := func() int {
		if frames := conf.Get().InsightsHistoricalFrames; frames != 0 {
			return frames
		}
		return 12 // 1 year by default
	}

	frameLength := func() time.Duration {
		defaultLen := 30 * 24 * time.Hour
		if s := conf.Get().InsightsHistoricalFrameLength; s != "" {
			parsed, err := str2duration.ParseDuration(s)
			if err != nil {
				log15.Error("insights: failed to parse site config insights.historical.frameLength", "error", err)
				return defaultLen
			}
			return parsed
		}
		return defaultLen
	}

	iterator := discovery.NewAllReposIterator(
		dbcache.NewIndexableReposLister(repoStore),
		repoStore,
		time.Now,
		envvar.SourcegraphDotComMode(),
		15*time.Minute,
		&prometheus.CounterOpts{
			Namespace: "src",
			Name:      "insights_historical_repositories_analyzed",
			Help:      "Counter of the number of repositories analyzed and queued for processing for insights.",
		})

	maxTime := time.Now().Add(-time.Duration(framesToBackfill()) * frameLength())

	historicalEnqueuer := &historicalEnqueuer{
		now:             time.Now,
		insightsStore:   insightsStore,
		repoStore:       database.Repos(workerBaseStore.Handle().DB()),
		dataSeriesStore: dataSeriesStore,
		limiter:         limiter,
		enqueueQueryRunnerJob: func(ctx context.Context, job *queryrunner.Job) error {
			_, err := queryrunner.EnqueueJob(ctx, workerBaseStore, job)
			return err
		},
		gitFirstEverCommit: (&cachedGitFirstEverCommit{impl: git.FirstEverCommit}).gitFirstEverCommit,
		gitFindRecentCommit: func(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitapi.Commit, error) {
			return git.Commits(ctx, repoName, git.CommitsOptions{N: 1, Before: target.Format(time.RFC3339), DateOrder: true})
		},

		// Fill e.g. the last 52 weeks of data, recording 1 point per week.
		framesToBackfill: framesToBackfill,
		frameLength:      frameLength,

		frameFilter: compression.NewHistoricalFilter(true, maxTime, insightsStore.Handle().DB()),

		allReposIterator: iterator.ForEach,
	}

	// We use a periodic goroutine here just for metrics tracking. We specify 5s here so it runs as
	// fast as possible without wasting CPU cycles, but in reality the handler itself can take
	// minutes to hours to complete as it intentionally enqueues work slowly to avoid putting
	// pressure on the system.
	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, 15*time.Minute, goroutine.NewHandlerWithErrorMessage(
		"insights_historical_enqueuer",
		historicalEnqueuer.Handler,
	), operation)
}

func getRateLimit(defaultValue rate.Limit) func() rate.Limit {
	return func() rate.Limit {
		val := conf.Get().InsightsHistoricalWorkerRateLimit

		var result rate.Limit
		if val == nil {
			result = defaultValue
		} else {
			result = rate.Limit(*val)
		}

		return result
	}
}

// RepoStore is a subset of the API exposed by the database.Repos() store (only the subset used by
// historicalEnqueuer.)
type RepoStore interface {
	GetByName(ctx context.Context, name api.RepoName) (*types.Repo, error)
}

// historicalEnqueuer effectively enqueues jobs that generate historical data for insights. Right
// now, it only supports search insights. It does this by adjusting the user's search query to be
// for a specific repo and commit like `repo:<repo>@<commit>`, where `<repo>` is every repository
// on Sourcegraph (one search per) and `<commit>` is a Git commit closest in time to the historical
// point in time we're trying to generate data for. A lot of effort is placed into doing the work
// slowly, linearly, and consistently over time without harming any other part of Sourcegraph
// (including the search API, by performing searches slowly and on single repositories at a time
// only.)
//
// It works roughly like this:
//
//   * For every repository on Sourcegraph (a subset on Sourcegraph.com):
//     * Build a list of time frames that we should consider
//	   * Check the commit index to see if any timeframes can be discarded (if they didn't change)
//     * For each frame:
//       * Find the oldest commit in the repository.
//         * For every unique search insight series (i.e. search query):
//           * Consider yielding/sleeping.
//           * If the series has data for this timeframe+repo already, nothing to do.
//           * If the timeframe we're generating data for is before the oldest commit in the repo, record a zero value.
//           * Else, locate the commit nearest to the point in time we're trying to get data for and
//             enqueue a queryrunner job to search that repository commit - recording historical data
//            for it.
//
// As you can no doubt see, there is much complexity and potential room for duplicative API calls
// here (e.g. "for every timeframe we list every repository"). For this exact reason, we do two
// things:
//
// 1. Cache duplicative calls to prevent performing heavy operations multiple times.
// 2. Lift heavy operations to the layer/loop one level higher, when it is sane to do so.
// 3. Ensure we perform work slowly, linearly, and with yielding/sleeping between any substantial
//    work being performed.
//
type historicalEnqueuer struct {
	// Required fields used for mocking in tests.
	now                   func() time.Time
	insightsStore         store.Interface
	dataSeriesStore       store.DataSeriesStore
	repoStore             RepoStore
	enqueueQueryRunnerJob func(ctx context.Context, job *queryrunner.Job) error
	gitFirstEverCommit    func(ctx context.Context, repoName api.RepoName) (*gitapi.Commit, error)
	gitFindRecentCommit   func(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitapi.Commit, error)
	frameFilter           compression.DataFrameFilter

	// framesToBackfill describes the number of historical timeframes to backfill data for.
	framesToBackfill func() int

	// frameLength describes the length of each timeframe to backfill data for.
	frameLength func() time.Duration

	// The iterator to use for walking over all repositories on Sourcegraph.
	allReposIterator func(ctx context.Context, each func(repoName string) error) error
	limiter          *rate.Limiter
}

func (h *historicalEnqueuer) Handler(ctx context.Context) error {
	// Discover all insights on the instance.
	foundInsights, err := h.dataSeriesStore.GetDataSeries(ctx, store.GetDataSeriesArgs{BackfillIncomplete: true})
	if err != nil {
		return errors.Wrap(err, "Discover")
	}

	// Deduplicate series that may be unique (e.g. different name/description) but do not have
	// unique data (i.e. use the same exact search query or webhook URL.)
	var (
		uniqueSeries    = map[string]itypes.InsightSeries{}
		sortedSeriesIDs []string
		multi           error
	)
	for _, series := range foundInsights {
		seriesID := series.SeriesID
		log15.Info("Loaded insight data series for historical processing", "series_id", seriesID)

		if _, exists := uniqueSeries[seriesID]; exists {
			continue
		}
		uniqueSeries[seriesID] = series
		sortedSeriesIDs = append(sortedSeriesIDs, seriesID)
	}
	if err := h.buildFrames(ctx, uniqueSeries, sortedSeriesIDs); err != nil {
		multi = multierror.Append(multi, err)
	}
	if err == nil {
		// we successfully performed a full repo iteration without any "hard" errors, so we will update the metadata
		// of each insight series to reflect they have seen a full iteration. This does not mean they were necessarily successful,
		// only that they had a chance to queue up queries for each repo.
		h.markInsightsComplete(ctx, foundInsights)
	}

	return multi
}

func (h *historicalEnqueuer) markInsightsComplete(ctx context.Context, completed []itypes.InsightSeries) {
	for _, series := range completed {
		_, err := h.dataSeriesStore.StampBackfill(ctx, series)
		if err != nil {
			// do nothing to preserve at least once semantics
			continue
		}
		log15.Info("insights: Insight marked backfill complete.", "series_id", series.SeriesID)
	}
}

// buildFrames is invoked to build historical data for all past timeframes that we care about
// backfilling data for. This is done in small chunks, specifically so that we perform work incrementally.
//
// It is only called if there is at least one insights series defined.
//
// It will return instantly if there are no unique series.
func (h *historicalEnqueuer) buildFrames(ctx context.Context, uniqueSeries map[string]itypes.InsightSeries, sortedSeriesIDs []string) error {
	if len(uniqueSeries) == 0 {
		return nil // nothing to do.
	}
	var multi error

	hardErr := h.allReposIterator(ctx, h.buildForRepo(ctx, uniqueSeries, sortedSeriesIDs, multi))
	return hardErr
}

func (h *historicalEnqueuer) buildForRepo(ctx context.Context, uniqueSeries map[string]itypes.InsightSeries, sortedSeriesIDs []string, softErr error) func(repoName string) error {
	return func(repoName string) error {
		// Lookup the repository (we need its database ID)
		repo, err := h.repoStore.GetByName(ctx, api.RepoName(repoName))
		if err != nil {
			// Ignore RepoNotFoundErr because it could just be that the repository was actually
			// deleted and allReposIterator had it cached.
			if errors.HasType(err, &database.RepoNotFoundErr{}) {
				return err // hard DB error
			} else {
				return nil
			}
		}
		// Find the first commit made to the repository on the default branch.
		firstHEADCommit, err := h.gitFirstEverCommit(ctx, api.RepoName(repoName))
		if err != nil {
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) || gitdomain.IsRepoNotExist(err) {
				return nil // no error - repo may not be cloned yet (or not even pushed to code host yet)
			}
			if strings.Contains(err.Error(), `failed (output: "usage: git rev-list [OPTION] <commit-id>...`) {
				return nil // repository is empty
			}
			// soft error, repo may be in a bad state but others might be OK.
			softErr = multierror.Append(softErr, errors.Wrap(err, "FirstEverCommit "+repoName))
			return nil
		}

		// For every series that we want to potentially gather historical data for, try.
		for _, seriesID := range sortedSeriesIDs {
			series := uniqueSeries[seriesID]

			frames := FirstOfMonthFrames(12, series.CreatedAt.Truncate(time.Hour*24))

			log15.Debug("insights: starting frames", "repo_id", repo.ID, "series_id", series.SeriesID, "frames", frames)
			plan := h.frameFilter.FilterFrames(ctx, frames, repo.ID)
			log15.Debug("insights: sampling historical data frames", "repo_id", repo.ID, "series_id", series.SeriesID, "frames", frames)

			for i := len(plan.Executions) - 1; i >= 0; i-- {
				queryExecution := plan.Executions[i]

				err := h.limiter.Wait(ctx)
				if err != nil {
					return err
				}

				to := queryExecution.RecordingTime.Add(time.Hour * 24)
				// If we already have data for this frame+repo+series, then there's nothing to do.
				var numDataPoints int
				numDataPoints, err = h.insightsStore.CountData(ctx, store.CountDataOpts{
					From:     &queryExecution.RecordingTime,
					To:       &to,
					SeriesID: &seriesID,
					RepoID:   &repo.ID,
				})
				if err != nil {
					softErr = multierror.Append(softErr, err)
					// In this case we will assume the point does not exist and query for it anyway.
				} else if numDataPoints > 0 {
					continue
				}

				// Build historical data for this unique timeframe+repo+series.
				hardErr, err := h.buildSeries(ctx, &buildSeriesContext{
					execution:       queryExecution,
					repo:            repo,
					firstHEADCommit: firstHEADCommit,
					seriesID:        seriesID,
					series:          series,
				})
				if err != nil {
					softErr = multierror.Append(softErr, err)
					continue
				}
				if hardErr != nil {
					return multierror.Append(softErr, hardErr)
				}
			}

		}
		return nil
	}
}

// buildSeriesContext describes context/parameters for a call to buildSeries()
type buildSeriesContext struct {
	// The timeframe we're building historical data for.

	execution *compression.QueryExecution

	// The repository we're building historical data for.
	repo *types.Repo

	// The first commit made in the repository on the default branch.
	firstHEADCommit *gitapi.Commit

	// The series we're building historical data for.
	seriesID string
	series   itypes.InsightSeries
}

// FirstOfMonthFrames builds a set of frames with a specific number of elements, such that all of the
// starting times of each frame < current will fall on the first of a month.
func FirstOfMonthFrames(numPoints int, current time.Time) []compression.Frame {
	if numPoints < 1 {
		return nil
	}
	times := make([]time.Time, 0, numPoints)
	year, month, _ := current.Date()
	firstOfCurrent := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	for i := 0 - numPoints + 1; i < 0; i++ {
		times = append(times, firstOfCurrent.AddDate(0, i, 0))
	}
	times = append(times, firstOfCurrent)
	times = append(times, current)

	frames := make([]compression.Frame, 0, len(times)-1)
	for i := 1; i < len(times); i++ {
		prev := times[i-1]
		frames = append(frames, compression.Frame{
			From: prev,
			To:   times[i],
		})
	}
	return frames
}

// buildSeries is invoked to build historical data for every unique timeframe * repo * series that
// could need backfilling. Note that this means that for a single search insight, this means this
// function may be called e.g. (52 timeframes) * (500000 repos) * (1 series) times.
//
// It may return both hard errors (e.g. DB connection failure, future series are unlikely to build)
// and soft errors (e.g. user's search query is invalid, future series are likely to build.)
func (h *historicalEnqueuer) buildSeries(ctx context.Context, bctx *buildSeriesContext) (hardErr, softErr error) {
	query := bctx.series.Query
	// TODO(slimsag): future: use the search query parser here to avoid any false-positives like a
	// search query with `content:"repo:"`.
	if strings.Contains(query, "repo:") {
		// We need to specify the repo: filter ourselves, so rewriting their query which already
		// contains this would be complex (we would need to enumerate all repos their query would
		// have matched the same way the search backend would've). We don't support this today.
		//
		// Another possibility is that they are specifying a non-default branch with the `repo:`
		// filter. We would need to handle this below if so - we don't today.
		return nil, nil
	}

	// Optimization: If the timeframe we're building data for starts (or ends) before the first commit in the
	// repository, then we know there are no results (the repository didn't have any commits at all
	// at that point in time.)
	repoName := string(bctx.repo.Name)
	if bctx.execution.RecordingTime.Before(bctx.firstHEADCommit.Author.Date) {
		args := bctx.execution.ToRecording(bctx.seriesID, repoName, bctx.repo.ID, 0.0)
		if err := h.insightsStore.RecordSeriesPoints(ctx, args); err != nil {
			hardErr = errors.Wrap(err, "RecordSeriesPoints Zero Value")
			return // DB error
		}
		return // success - nothing else to do
	}

	// At this point, we know:
	//
	// 1. We're building data for the `[from, to]` timeframe.
	// 2. We're building data for the search `query`.
	//
	// We need a way to find out in that historical timeframe what the total # of results was.
	// There are only two ways to do that:
	//
	// 1. Run `type:diff` searches, this would give us matching lines added/removed/changed over
	//    time. To use this, we would need to ensure we *start* looking for historical data at the
	//    very first commit in the repo, and keep a running tally of added/removed/changed lines -
	//    this requires a lot of book-keeping.
	// 2. Choose some commits in the timeframe `[from, to]` (or, if none exist in that timeframe,
	//    whatever commit is closest) and perform a live/unindexed search for that `repo:<repo>@commit`
	//    which will effectively search the repo at that point in time.
	//
	// We do the 2nd, and start by trying to locate the commit most recent to the start of the
	// timeframe we're trying to fill in historical data for.
	// If we have a revision already derived from the execution plan, we will use that revision. Otherwise we will
	// look it up from gitserver.
	var revision string

	if len(bctx.execution.Revision) > 0 {
		revision = bctx.execution.Revision
	} else {
		recentCommits, err := h.gitFindRecentCommit(ctx, bctx.repo.Name, bctx.execution.RecordingTime)
		if err != nil {
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) || gitdomain.IsRepoNotExist(err) {
				return // no error - repo may not be cloned yet (or not even pushed to code host yet)
			}
			hardErr = errors.Wrap(err, "FindNearestCommit")
			return
		}
		var nearestCommit *gitapi.Commit
		if len(recentCommits) > 0 {
			nearestCommit = recentCommits[0]
		}
		if nearestCommit == nil {
			log15.Error("null commit", "repo_id", bctx.repo.ID, "series_id", bctx.series.SeriesID, "from", bctx.execution.RecordingTime)
			return // repository has no commits / is empty. Maybe not yet pushed to code host.
		}
		if nearestCommit.Committer == nil {
			log15.Error("null committer", "repo_id", bctx.repo.ID, "series_id", bctx.series.SeriesID, "from", bctx.execution.RecordingTime)
			return
		}
		log15.Debug("nearest_commit", "repo_id", bctx.repo.ID, "series_id", bctx.series.SeriesID, "from", bctx.execution.RecordingTime, "revhash", nearestCommit.ID.Short(), "time", nearestCommit.Committer.Date)
		revision = string(nearestCommit.ID)
	}

	// Build the search query we will run. The most important part here is
	query = withCountUnlimited(query)
	query = fmt.Sprintf("%s repo:^%s$@%s", query, regexp.QuoteMeta(repoName), revision)

	job := bctx.execution.ToQueueJob(bctx.seriesID, query, priority.Unindexed, priority.FromTimeInterval(bctx.execution.RecordingTime, bctx.series.CreatedAt))
	hardErr = h.enqueueQueryRunnerJob(ctx, job)
	return
}

// cachedGitFirstEverCommit is a simple in-memory cache for gitFirstEverCommit calls. It does so
// using a map, and entries are never evicted because they are expected to be small and in general
// unchanging.
type cachedGitFirstEverCommit struct {
	impl func(ctx context.Context, repoName api.RepoName) (*gitapi.Commit, error)

	mu    sync.Mutex
	cache map[api.RepoName]*gitapi.Commit
}

func (c *cachedGitFirstEverCommit) gitFirstEverCommit(ctx context.Context, repoName api.RepoName) (*gitapi.Commit, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = map[api.RepoName]*gitapi.Commit{}
	}
	if cached, ok := c.cache[repoName]; ok {
		return cached, nil
	}
	entry, err := c.impl(ctx, repoName)
	if err != nil {
		return nil, err
	}
	c.cache[repoName] = entry
	return entry, nil
}
