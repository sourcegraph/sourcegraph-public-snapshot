package background

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

// newInsightHistoricalEnqueuer returns a background goroutine which will periodically find all of the search
// insights across all user settings, and determine for which dates they do not have data and attempt
// to backfill them by enqueueing work for executing searches with `before:` and `after:` filter
// ranges.
func newInsightHistoricalEnqueuer(ctx context.Context, workerBaseStore *basestore.Store, settingStore discovery.SettingStore, insightsStore *store.Store, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"insights_historical_enqueuer",
		metrics.WithCountHelp("Total number of insights historical enqueuer executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    fmt.Sprintf("HistoricalEnqueuer.Run"),
		Metrics: metrics,
	})

	historicalEnqueuer := &historicalEnqueuer{
		now:           time.Now,
		settingStore:  settingStore,
		insightsStore: insightsStore,
		repoStore:     database.Repos(workerBaseStore.Handle().DB()),
		enqueueQueryRunnerJob: func(ctx context.Context, job *queryrunner.Job) error {
			_, err := queryrunner.EnqueueJob(ctx, workerBaseStore, job)
			return err
		},

		// Fill the last 52 weeks of data, recording 1 point per week.
		//
		framesToBackfill: 52,
		frameLength:      7 * 24 * time.Hour,

		allReposIterator: &discovery.AllReposIterator{
			DefaultRepoStore: database.DefaultRepos(workerBaseStore.Handle().DB()),
			RepoStore:        database.Repos(workerBaseStore.Handle().DB()),

			// If a new repository is added to Sourcegraph, it can take 0-15m for it to be picked
			// up for backfilling.
			RepositoryListCacheTime: 15 * time.Minute,
		},
	}

	// We use a periodic goroutine here just for metrics tracking. We specify 5s here so it runs as
	// fast as possible without wasting CPU cycles, but in reality the handler itself can take
	// minutes to hours to complete as it intentionally enqueues work slowly to avoid putting
	// pressure on the system.
	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, 5*time.Second, goroutine.NewHandlerWithErrorMessage(
		"insights_historical_enqueuer",
		historicalEnqueuer.Handler,
	), operation)
}

type historicalEnqueuer struct {
	// Required fields used for mocking in tests.
	now                   func() time.Time
	settingStore          discovery.SettingStore
	insightsStore         *store.Store
	repoStore             *database.RepoStore
	enqueueQueryRunnerJob func(ctx context.Context, job *queryrunner.Job) error

	// framesToBackfill describes the number of historical timeframes to backfill data for.
	framesToBackfill int

	// frameLength describes the length of each timeframe to backfill data for.
	frameLength time.Duration

	// The iterator to use for walking over all repositories on Sourcegraph.
	allReposIterator *discovery.AllReposIterator
}

func (h *historicalEnqueuer) Handler(ctx context.Context) error {
	// Discover all insights on the instance.
	insights, err := discovery.Discover(ctx, h.settingStore)
	if err != nil {
		return errors.Wrap(err, "Discover")
	}

	// Deduplicate series that may be unique (e.g. different name/description) but do not have
	// unique data (i.e. use the same exact search query or webhook URL.)
	var (
		uniqueSeries = map[string]*schema.InsightSeries{}
		multi        error
	)
	for _, insight := range insights {
		for _, series := range insight.Series {
			seriesID, err := discovery.EncodeSeriesID(series)
			if err != nil {
				multi = multierror.Append(multi, err)
				continue
			}
			_, exists := uniqueSeries[seriesID]
			if exists {
				continue
			}
			uniqueSeries[seriesID] = series
		}
	}
	if err := h.buildFrames(ctx, uniqueSeries); err != nil {
		return multierror.Append(multi, err)
	}
	return nil
}

// buildFrames is invoked to build historical data for all past timeframes that we care about
// backfilling data for. This is done in small chunks, e.g. 52 frames to backfill with each frame
// being 7 days long, specifically so that we perform work incrementally.
//
// It is only called if there is at least one insights series defined.
//
// It will return instantly if there are no unique series.
func (h *historicalEnqueuer) buildFrames(ctx context.Context, uniqueSeries map[string]*schema.InsightSeries) error {
	if len(uniqueSeries) == 0 {
		return nil // nothing to do.
	}
	var multi error
	for frame := 0; frame < h.framesToBackfill; frame++ {
		// Determine the exact start and end time of this timeframe.
		from := h.now().Add(-time.Duration(frame+1) * h.frameLength)
		to := h.now().Add(-time.Duration(frame) * h.frameLength)
		if h.now().After(h.now().Add(24 * time.Hour)) {
			// We exclude today because it is the regular enqueuer's job to enqueue work for today.
			to = to.Add(-24 * time.Hour)
		}

		// Build a function which tells if a given series has data in this timeframe for a specific
		// repository.
		seriesIDsWithData, err := h.insightsStore.DistinctSeriesWithData(ctx, from, to)
		if err != nil {
			return multierror.Append(multi, errors.Wrap(err, "DistinctSeriesWithData")) // DB error, no point in continuing.
		}
		haveData := map[string]struct{}{}
		for _, id := range seriesIDsWithData {
			haveData[id] = struct{}{}
		}
		haveDataForRepo := func(seriesID string, id api.RepoID, from, to time.Time) bool {
			// TODO(slimsag): future: actually check if the insight is missing data for the given
			// repo ID. In the meantime, if there is *any* datapoint in this timeframe we won't do
			// backfilling. e.g., if a new repository is added to Sourcegraph after we built
			// historical data for an insight, it will not be accounted for.
			_, haveData := haveData[seriesID]
			return haveData
		}

		// Build historical data for this timeframe.
		softErr, hardErr := h.buildFrame(ctx, uniqueSeries, from, to, haveDataForRepo)
		if softErr != nil {
			multi = multierror.Append(multi, softErr)
			continue
		}
		if hardErr != nil {
			return multierror.Append(multi, hardErr)
		}
	}
	return multi // return any soft errors we encountered
}

// buildFrame is invoked to build historical data for a specific past timeframe that we care about
// backfilling data for.
//
// It is expected to backfill data for all unique series that are missing data, across all repos
// (using h.allReposIterator.)
//
// It should not backfill data for series that already have data recorded for a given repo, checked
// via haveDataForRepo().
//
// It may return both hard errors (e.g. DB connection failure, future frames are unlikely to build)
// and soft errors (e.g. user made a mistake or we did partial work, future frames will likely
// succeed.)
func (h *historicalEnqueuer) buildFrame(
	ctx context.Context,
	uniqueSeries map[string]*schema.InsightSeries,
	from time.Time,
	to time.Time,
	haveDataForRepo func(seriesID string, id api.RepoID, from, to time.Time) bool,
) (hardErr, softErr error) {
	// We yield frequently for a small period of time for a few reasons:
	//
	// 1. To not call buildSeries too quickly and enqueue millions of jobs rapidly.
	// 2. To avoid calling repoStore.GetByName() and git.FirstEverCommit() in a tight
	//    loop for potentially 500,000+ repositories if there is actually no work to
	//    perform (because all have had historical data built already.)
	//
	lastIteration := time.Now()
	yield := func() {
		if diff := time.Since(lastIteration); diff < 100*time.Millisecond {
			time.Sleep(diff)
			lastIteration = time.Now()
		}
	}

	// For every repository that we want to potentially gather historical data for.
	hardErr = h.allReposIterator.ForEach(ctx, func(repoName string) error {
		yield()

		// Lookup the repository (we need its database ID)
		repo, err := h.repoStore.GetByName(ctx, api.RepoName(repoName))
		if err != nil {
			return err // hard DB error
		}

		// Find the first commit made to the repository on the default branch.
		firstHEADCommit, err := git.FirstEverCommit(ctx, api.RepoName(repoName))
		if err != nil {
			if gitserver.IsRevisionNotFound(err) || vcs.IsRepoNotExist(err) {
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
		for seriesID, series := range uniqueSeries {
			yield()

			// If we already have data for this frame+repo+series, then there's nothing to do.
			if haveDataForRepo(seriesID, repo.ID, from, to) {
				continue
			}

			// Build historical data for this unique timeframe+repo+series.
			softErr, hardErr := h.buildSeries(ctx, &buildSeriesContext{
				from:            from,
				to:              to,
				repo:            repo,
				firstHEADCommit: firstHEADCommit,
				seriesID:        seriesID,
				series:          series,
			})
			if softErr != nil {
				softErr = multierror.Append(softErr, softErr)
				continue
			}
			if hardErr != nil {
				return multierror.Append(softErr, hardErr)
			}
		}
		return nil
	})
	return
}

// buildSeriesContext describes context/parameters for a call to buildSeries()
type buildSeriesContext struct {
	// The timeframe we're building historical data for.
	from, to time.Time

	// The repository we're building historical data for.
	repo *types.Repo

	// The first commit made in the repository on the default branch.
	firstHEADCommit *git.Commit

	// The series we're building historical data for.
	seriesID string
	series   *schema.InsightSeries
}

// buildSeries is invoked to build historical data for every unique timeframe * repo * series that
// could need backfilling. Note that this means that for a single search insight, this means this
// function may be called e.g. (52 timeframes) * (500000 repos) * (1 series) times.
//
// It may return both hard errors (e.g. DB connection failure, future series are unlikely to build)
// and soft errors (e.g. user's search query is invalid, future series are likely to build.)
func (h *historicalEnqueuer) buildSeries(ctx context.Context, bctx *buildSeriesContext) (hardErr, softErr error) {
	// First, can we actually build historical data for this series?
	if bctx.series.Webhook != "" {
		return nil, nil // we cannot build historical data for webhook insights
	}
	query := bctx.series.Search
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

	// We're trying to find the # of search results at the middle of the timeframe, ideally.
	frameDuration := bctx.to.Sub(bctx.from)
	frameMidpoint := bctx.from.Add(frameDuration / 2)

	// Optimization: If the timeframe we're building data for ends before the first commit in the
	// repository, then we know there are no results (the repository didn't have any commits at all
	// at that point in time.)
	repoName := string(bctx.repo.Name)
	if bctx.to.Before(bctx.firstHEADCommit.Author.Date) {
		if err := h.insightsStore.RecordSeriesPoint(ctx, store.RecordSeriesPointArgs{
			SeriesID: bctx.seriesID,
			Point: store.SeriesPoint{
				Time:  frameMidpoint,
				Value: 0, // no matches
			},
			RepoName: &repoName,
			RepoID:   &bctx.repo.ID,
		}); err != nil {
			hardErr = errors.Wrap(err, "RecordSeriesPoint")
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
	// We do the 2nd, and start by trying to locate the commit nearest to the middle of the
	// timeframe we're trying to fill in historical data for.
	nearestCommit, err := git.FindNearestCommit(ctx, bctx.repo.Name, "HEAD", frameMidpoint)
	if err != nil {
		if gitserver.IsRevisionNotFound(err) || vcs.IsRepoNotExist(err) {
			return // no error - repo may not be cloned yet (or not even pushed to code host yet)
		}
		hardErr = errors.Wrap(err, "FindNearestCommit")
		return
	}
	if nearestCommit == nil {
		return // repository has no commits / is empty. Maybe not yet pushed to code host.
	}

	// Build the search query we will run. The most important part here is
	query = withCountUnlimited(query)
	query = fmt.Sprintf("%s repo:^%s$@%s", query, regexp.QuoteMeta(repoName), string(nearestCommit.ID))

	hardErr = h.enqueueQueryRunnerJob(ctx, &queryrunner.Job{
		SeriesID:    bctx.seriesID,
		SearchQuery: query,
		RecordTime:  &frameMidpoint,
		State:       "queued",
	})
	return
}
