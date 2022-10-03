package pipeline

import (
	"context"
	golog "log"
	"math/rand"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/search"
	streamclient "github.com/sourcegraph/sourcegraph/internal/search/streaming"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type nilSearchClient struct {
}

func (f *nilSearchClient) Search(ctx context.Context, query string, patternType *string, sender streamclient.Sender) (*search.Alert, error) {
	return nil, nil
}

func NewNilSearchClient() streaming.SearchClient {
	return &nilSearchClient{}
}

type BackfillStatus interface{}
type seriesID string

type BackfillerFactory func(series *types.InsightSeries) Backfiller

type BackfillProgress struct {
	SeriesID    string
	RemaingCost *int
}

type Backfiller interface {
	Run(ctx context.Context, reportProgress func(BackfillProgress)) error
	Status(ctx context.Context) (BackfillStatus, error)
	TotalCost() *int
	RemaingCost() *int
}

type NewSearchClientFunc func() streaming.SearchClient

type PipelineContext struct {
	Series *types.InsightSeries
	Repo   *itypes.Repo
}

type SearchJobGeneratorInput struct {
	PipelineContext
}

type SearchJobGeneratorOutput struct {
	PipelineContext
	Job queryrunner.Job
}

type iterationResult struct {
	repoID   *api.RepoID
	err      error
	duration time.Duration
}

// This says what to iterate over
type UnitOfWorkGenerator func(ctx context.Context, series *types.InsightSeries) (chan<- error, chan<- PipelineContext)

type SearchJobGenerator func(ctx context.Context, in <-chan SearchJobGeneratorInput) (chan<- error, chan<- SearchJobGeneratorOutput)
type SearchJobRunner func(ctx context.Context, in <-chan SearchJobGeneratorOutput) (chan<- error, chan<- SearchJobGeneratorOutput)
type SearchResultsPersister func(ctx context.Context, in <-chan SearchJobGeneratorOutput) (chan<- error, chan<- SearchJobGeneratorOutput)

func NewBackfillerFactory(firstCommit FirstCommitFunc, recentCommit FindRecentCommitFunc, newSearchClient NewSearchClientFunc, repoStore database.RepoStore, insightStore store.Interface, compressionPlan compression.DataFrameFilter) BackfillerFactory {
	return func(series *types.InsightSeries) Backfiller {
		//TODO: load existing backfiller for series if it exists

		return &backfiller{
			firstCommit:     firstCommit,
			recentCommit:    recentCommit,
			newSeachClient:  newSearchClient,
			repoStore:       repoStore,
			compressionPlan: compressionPlan,
			insightStore:    insightStore,
			logger:          log.Scoped("insights_backfill_pipeline", ""),
			series:          series,
		}
	}
}

type backfiller struct {
	//dependencies
	firstCommit     FirstCommitFunc
	recentCommit    FindRecentCommitFunc
	newSeachClient  NewSearchClientFunc
	repoStore       database.RepoStore
	compressionPlan compression.DataFrameFilter
	insightStore    store.Interface

	//state
	series             *types.InsightSeries
	repos              []api.RepoID
	queryProgressIndex int
	logger             log.Logger
}

func (b *backfiller) Run(ctx context.Context, reportProgress func(BackfillProgress)) error {

	errChan := make(chan error)
	results := make(chan iterationResult, len(b.repos))
	reposChan := b.repoGenerator(ctx, errChan)

	var wg sync.WaitGroup
	go func() {
		wg.Wait()
		close(results)
	}()
	// launching a single worker
	wg.Add(1)
	goroutine.Go(func() {
		defer wg.Done()
		b.worker(ctx, b.series, reposChan, results)
	})

	for result := range results {
		if result.err != nil {
			rId := int32(*result.repoID)
			b.logger.Warn("insights pipeline run failed", log.String("series", b.series.SeriesID), log.Int32("repo", rId))
			continue
		}
		b.saveProgress(ctx, reportProgress, result)
	}

	return nil
}

func (b *backfiller) worker(ctx context.Context, series *types.InsightSeries, reposIn <-chan *itypes.Repo, results chan<- iterationResult) {
	for repo := range reposIn {
		start := time.Now()
		getSearchPlan := makeSearchPlanFunc(b.logger, b.firstCommit, b.compressionPlan)
		generateSearchJobs := makeGetSearchJobsFunc(b.logger, b.firstCommit, b.recentCommit)
		runSearches := makeRunSearchFunc(b.logger, b.newSeachClient())
		saveSearches := makeSaveResultsFunc(b.logger, b.insightStore)

		plan, _ := getSearchPlan(ctx, b.series, repo)
		searchJobChan, _ := generateSearchJobs(ctx, b.series, repo, plan)
		searchResultsChan := runSearches(ctx, repo, searchJobChan)
		err := saveSearches(ctx, b.series, searchResultsChan)
		results <- iterationResult{
			repoID:   &repo.ID,
			err:      err,
			duration: time.Since(start),
		}
	}
}

func (b *backfiller) repoGenerator(ctx context.Context, errChan chan<- error) <-chan *itypes.Repo {
	reposChan := make(chan *itypes.Repo)
	defer func() {
		close(reposChan)
	}()

	go func(ctx context.Context) {
		for _, repoId := range b.repos {
			r, err := b.repoStore.Get(ctx, repoId)
			if err != nil {
				errChan <- err
				continue
			}
			reposChan <- r
		}
	}(ctx)

	return reposChan
}

func (b *backfiller) resolveRepos(ctx context.Context) error {
	b.logger.Warn("getting series repos")
	if len(b.repos) != 0 {
		return nil
	}

	var getRepos getSeriesRepos
	if len(b.series.Repositories) == 0 {
		getRepos = getAllRepos(b.repoStore)
	} else {
		getRepos = getNamedRepos(b.repoStore)
	}

	repos, err := getRepos(ctx, b.series)
	if err != nil {
		return err
	}
	b.repos = repos
	return nil
}

func (b *backfiller) saveProgress(ctx context.Context, reportProgress func(BackfillProgress), result iterationResult) error {
	// todo add some persistentce
	if result.err == nil {
		// make result.repoID complete
	}
	b.logger.Warn("saving backfill progress")
	reportProgress(BackfillProgress{SeriesID: b.series.SeriesID, RemaingCost: b.RemaingCost()})
	return nil
}

func (b *backfiller) Status(ctx context.Context) (BackfillStatus, error) {
	return nil, nil
}

func (b *backfiller) TotalCost() *int {
	repos := len(b.repos)
	if repos > 0 {
		return &repos
	}
	return nil
}
func (b *backfiller) RemaingCost() *int {
	total := b.TotalCost()
	if total != nil {
		remaining := *total - b.queryProgressIndex
		return &remaining
	}
	return nil
}

//JOBS

// Enumerate Repos for series

type getSeriesRepos func(ctx context.Context, series *types.InsightSeries) ([]api.RepoID, error)

func getAllRepos(repoStore database.RepoStore) getSeriesRepos {
	return func(ctx context.Context, series *types.InsightSeries) ([]api.RepoID, error) {
		repoIds := make([]api.RepoID, 0)
		repos, err := repoStore.ListMinimalRepos(ctx, database.ReposListOptions{})
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			repoIds = append(repoIds, repo.ID)
		}
		return repoIds, nil
	}
}

func getNamedRepos(repoStore database.RepoStore) getSeriesRepos {
	return func(ctx context.Context, series *types.InsightSeries) ([]api.RepoID, error) {
		repoIds := make([]api.RepoID, 0)
		list, err := repoStore.ListMinimalRepos(ctx, database.ReposListOptions{Names: series.Repositories})
		if err != nil {
			return nil, err
		}
		for _, repo := range list {
			repoIds = append(repoIds, repo.ID)
		}
		return repoIds, nil
	}
}

func getExistingRepos(repoIDs []api.RepoID) getSeriesRepos {
	return func(ctx context.Context, series *types.InsightSeries) ([]api.RepoID, error) {
		return repoIDs, nil
	}
}

// Generate Repo Series
type getRepoSeriesData func(ctx context.Context, series *types.InsightSeries, repo api.RepoID) ([]store.RecordSeriesPointArgs, error)

func getSearchSeriesData(ctx context.Context, series *types.InsightSeries, repo api.RepoID) ([]store.RecordSeriesPointArgs, error) {
	return nil, nil
}

func getCaptureGroupSeriesData(ctx context.Context, series *types.InsightSeries, repo api.RepoID) ([]store.RecordSeriesPointArgs, error) {
	return nil, nil
}

type FirstCommitFunc func(context.Context, api.RepoName) (*gitdomain.Commit, error)

func makeSearchPlanFunc(logger log.Logger, getFirstEverCommit FirstCommitFunc, compressionPlan compression.DataFrameFilter) func(ctx context.Context, series *types.InsightSeries, repo *itypes.Repo) (plan *compression.BackfillPlan, err error) {
	return func(ctx context.Context, series *types.InsightSeries, repo *itypes.Repo) (plan *compression.BackfillPlan, err error) {
		// span, ctx := ot.StartSpanFromContext(policy.WithShouldTrace(ctx, true), "backfiller.getSearchJobs")
		// span.SetTag("repo_id", repo.ID)
		// defer func() {
		// 	// if err != nil {
		// 	// 	span.LogFields(log.Error(err))
		// 	// }
		// 	span.Finish()
		// }()
		logger.Warn("making search plan")
		// Find the first commit made to the repository on the default branch.
		_, err = getFirstEverCommit(ctx, repo.Name)
		if err != nil {
			// span.LogFields(log.Error(err))
			// for _, stats := range statistics {
			// 	// mark all series as having one error since this error is at the repo level (affects all series)
			// 	stats.Errored += 1
			// }

			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) || gitdomain.IsRepoNotExist(err) {
				return nil, err // error - repo may not be cloned yet (or not even pushed to code host yet)
			}
			if errors.Is(err, discovery.EmptyRepoErr) {
				return nil, nil
			}
			return nil, err
		}

		frames := timeseries.BuildFrames(12, timeseries.TimeInterval{
			Unit:  types.IntervalUnit(series.SampleIntervalUnit),
			Value: series.SampleIntervalValue,
		}, series.CreatedAt.Truncate(time.Hour*24))

		searchPlan := compressionPlan.FilterFrames(ctx, frames, repo.ID)
		return &searchPlan, nil
	}
}

type generateJobResult struct {
	job *queryrunner.Job
	err error
}

func makeGetSearchJobsFunc(
	logger log.Logger,
	getFirstEverCommit FirstCommitFunc,
	gitFindRecentCommit FindRecentCommitFunc,
) func(ctx context.Context, series *types.InsightSeries, repo *itypes.Repo, plan *compression.BackfillPlan) (<-chan generateJobResult, error) {
	return func(ctx context.Context, series *types.InsightSeries, repo *itypes.Repo, plan *compression.BackfillPlan) (<-chan generateJobResult, error) {
		//setup helper func
		buildJob := makeSearchJobFunc(logger, gitFindRecentCommit)
		outputChannel := make(chan generateJobResult)

		if plan == nil {
			close(outputChannel)
			return outputChannel, nil
		}

		// don't want to duplciate this, but also don't want to hold ref to first commit
		firstHEADCommit, err := getFirstEverCommit(ctx, repo.Name)
		if err != nil {
			close(outputChannel)
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) || gitdomain.IsRepoNotExist(err) {
				return outputChannel, err // error - repo may not be cloned yet (or not even pushed to code host yet)
			}
			if errors.Is(err, discovery.EmptyRepoErr) {
				return outputChannel, nil
			}
			return outputChannel, err
		}
		go func(ctx context.Context, outputChannel chan generateJobResult) {
			defer func() {
				close(outputChannel)
				if err := recover(); err != nil {
					stack := debug.Stack()
					golog.Printf("goroutine panic: %v\n%s", err, stack)
				}
			}()

			for i := len(plan.Executions) - 1; i >= 0; i-- {
				if ctx.Err() != nil {
					outputChannel <- generateJobResult{job: nil, err: ctx.Err()}
					break
				}
				queryExecution := plan.Executions[i]
				// Build historical data for this unique timeframe+repo+series.
				analyzeErr, job, _ := buildJob(ctx, &buildSeriesContext{
					execution:       queryExecution,
					repoName:        repo.Name,
					id:              repo.ID,
					firstHEADCommit: firstHEADCommit,
					seriesID:        series.SeriesID,
					series:          *series,
				})
				outputChannel <- generateJobResult{job: job, err: analyzeErr}
			}
		}(ctx, outputChannel)

		return outputChannel, err
	}
}

type buildSeriesContext struct {
	// The timeframe we're building historical data for.

	execution *compression.QueryExecution

	// The repository we're building historical data for.
	id       api.RepoID
	repoName api.RepoName

	// The first commit made in the repository on the default branch.
	firstHEADCommit *gitdomain.Commit

	// The series we're building historical data for.
	seriesID string
	series   types.InsightSeries
}

type searchJobFunc func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.Job, preempted []store.RecordSeriesPointArgs)

type FindRecentCommitFunc func(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error)

func makeSearchJobFunc(logger log.Logger, gitFindRecentCommit FindRecentCommitFunc) searchJobFunc {
	return func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.Job, preempted []store.RecordSeriesPointArgs) {
		logger.Warn("making search job")
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
			return nil, nil, nil
		}

		// Optimization: If the timeframe we're building data for starts (or ends) before the first commit in the
		// repository, then we know there are no results (the repository didn't have any commits at all
		// at that point in time.)
		repoName := string(bctx.repoName)
		if bctx.execution.RecordingTime.Before(bctx.firstHEADCommit.Author.Date) {
			//a.statistics[bctx.seriesID].Preempted += 1
			return err, nil, bctx.execution.ToRecording(bctx.seriesID, repoName, bctx.id, 0.0)

			// return // success - nothing else to do
		}

		var revision string
		recentCommits, err := gitFindRecentCommit(ctx, bctx.repoName, bctx.execution.RecordingTime)
		if err != nil {
			if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) || gitdomain.IsRepoNotExist(err) {
				return // no error - repo may not be cloned yet (or not even pushed to code host yet)
			}
			err = errors.Append(err, errors.Wrap(err, "FindNearestCommit"))
			return
		}
		var nearestCommit *gitdomain.Commit
		if len(recentCommits) > 0 {
			nearestCommit = recentCommits[0]
		}
		if nearestCommit == nil {
			//a.statistics[bctx.seriesID].Errored += 1
			return // repository has no commits / is empty. Maybe not yet pushed to code host.
		}
		if nearestCommit.Committer == nil {
			//a.statistics[bctx.seriesID].Errored += 1
			return
		}
		revision = string(nearestCommit.ID)

		// Construct the search query that will generate data for this repository and time (revision) tuple.
		var newQueryStr string
		modifiedQuery, err := querybuilder.SingleRepoQuery(querybuilder.BasicQuery(query), repoName, revision, querybuilder.CodeInsightsQueryDefaults(len(bctx.series.Repositories) == 0))
		if err != nil {
			err = errors.Append(err, errors.Wrap(err, "SingleRepoQuery"))
			return
		}
		newQueryStr = modifiedQuery.String()
		if bctx.series.GroupBy != nil {
			computeQuery, computeErr := querybuilder.ComputeInsightCommandQuery(modifiedQuery, querybuilder.MapType(*bctx.series.GroupBy))
			if computeErr != nil {
				err = errors.Append(err, errors.Wrap(err, "ComputeInsightCommandQuery"))
				return
			}
			newQueryStr = computeQuery.String()
		}

		job = queryrunner.ToQueueJob(bctx.execution, bctx.seriesID, newQueryStr, priority.Unindexed, priority.FromTimeInterval(bctx.execution.RecordingTime, bctx.series.CreatedAt))
		return err, job, preempted
	}
}

type searchResult struct {
	count       int
	capture     string
	repo        *itypes.Repo
	pointInTime time.Time
}

type runSearchResult struct {
	err    error
	result searchResult
}

func makeRunSearchFunc(logger log.Logger, searchClient streaming.SearchClient) func(context.Context, *itypes.Repo, <-chan generateJobResult) <-chan runSearchResult {
	return func(ctx context.Context, repo *itypes.Repo, in <-chan generateJobResult) <-chan runSearchResult {

		out := make(chan runSearchResult)
		go func(ctx context.Context, outputChannel chan runSearchResult) {
			defer func() {
				close(out)
				if err := recover(); err != nil {
					stack := debug.Stack()
					golog.Printf("goroutine panic: %v\n%s", err, stack)
				}
			}()
			for r := range in {
				if r.err != nil {
					//TODO: what to do
					return
				}
				// run search
				// some made up values
				time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
				logger.Warn("running the search job")
				outputChannel <- runSearchResult{err: nil, result: searchResult{count: 10, capture: "", repo: repo, pointInTime: *r.job.RecordTime}}
			}
		}(ctx, out)
		return out
	}
}

func makeSaveResultsFunc(logger log.Logger, insightStore store.Interface) func(ctx context.Context, series *types.InsightSeries, in <-chan runSearchResult) error {
	return func(ctx context.Context, series *types.InsightSeries, in <-chan runSearchResult) error {
		points := make([]store.RecordSeriesPointArgs, 0, 12)
		for search := range in {
			if search.err != nil {
				//TODO: what to do
				continue
			}
			repoName := string(search.result.repo.Name)
			repoID := search.result.repo.ID
			capture := search.result.capture
			points = append(points,
				store.RecordSeriesPointArgs{
					SeriesID: series.SeriesID,
					Point: store.SeriesPoint{
						SeriesID: series.SeriesID,
						Time:     search.result.pointInTime,
						Value:    float64(search.result.count),
						Capture:  &capture,
					},
					RepoName:    &repoName,
					RepoID:      &repoID,
					PersistMode: store.RecordMode,
				},
			)

		}
		logger.Warn("writing search results")
		return insightStore.RecordSeriesPoints(ctx, points)
	}

}
