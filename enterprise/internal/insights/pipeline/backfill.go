package pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

type BackfillRequest struct {
	Series *types.InsightSeries
	Repo   *itypes.MinimalRepo
}

type requestContext struct {
	backfillRequest *BackfillRequest
}

type Backfiller interface {
	Run(ctx context.Context, request BackfillRequest) error
}

type GitCommitClient interface {
	FirstCommit(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error)
	RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error)
}

type SearchJobGenerator func(ctx context.Context, req requestContext) (context.Context, *requestContext, []*queryrunner.SearchJob, error)
type SearchRunner func(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.SearchJob, err error) (context.Context, *requestContext, []store.RecordSeriesPointArgs, error)
type ResultsPersister func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs, err error) (*requestContext, error)

type BackfillerConfig struct {
	CommitClient    GitCommitClient
	CompressionPlan compression.DataFrameFilter
	SearchHandlers  map[types.GenerationMethod]queryrunner.InsightsHandler
	InsightStore    store.Interface

	SearchPlanWorkerLimit   int
	SearchRunnerWorkerLimit int
}

func NewDefaultBackfiller(config BackfillerConfig) Backfiller {
	logger := log.Scoped("insightsBackfiller", "")
	searchJobGenerator := makeSearchJobsFunc(logger, config.CommitClient, config.CompressionPlan, config.SearchPlanWorkerLimit)
	searchRunner := makeRunSearchFunc(logger, config.SearchHandlers, config.SearchRunnerWorkerLimit)
	persister := makeSaveResultsFunc(logger, config.InsightStore)
	return newBackfiller(searchJobGenerator, searchRunner, persister)

}

func newBackfiller(jobGenerator SearchJobGenerator, searchRunner SearchRunner, resultsPersister ResultsPersister) Backfiller {
	return &backfiller{
		searchJobGenerator: jobGenerator,
		searchRunner:       searchRunner,
		persister:          resultsPersister,
	}

}

type backfiller struct {
	//dependencies
	searchJobGenerator SearchJobGenerator
	searchRunner       SearchRunner
	persister          ResultsPersister
}

func (b *backfiller) Run(ctx context.Context, req BackfillRequest) error {
	_, err := b.persister(b.searchRunner(b.searchJobGenerator(ctx, requestContext{backfillRequest: &req})))
	return err
}

// Implementation of steps for Backfill process

func makeSearchJobsFunc(logger log.Logger, commitClient GitCommitClient, compressionPlan compression.DataFrameFilter, searchJobWorkerLimit int) SearchJobGenerator {
	return func(ctx context.Context, reqContext requestContext) (context.Context, *requestContext, []*queryrunner.SearchJob, error) {
		jobs := make([]*queryrunner.SearchJob, 0, 12)
		if reqContext.backfillRequest == nil {
			return ctx, &reqContext, jobs, errors.New("backfill request provided")
		}

		req := reqContext.backfillRequest
		buildJob := makeHistoricalSearchJobFunc(logger, commitClient)
		logger.Debug("making search plan")
		// Find the first commit made to the repository on the default branch.
		firstHEADCommit, err := commitClient.FirstCommit(ctx, req.Repo.Name)
		if err != nil {
			if errors.Is(err, discovery.EmptyRepoErr) {
				// This is fine it's empty there is no work to be done
				return ctx, &reqContext, jobs, nil
			}

			return ctx, &reqContext, jobs, err
		}

		frames := timeseries.BuildFrames(12, timeseries.TimeInterval{
			Unit:  types.IntervalUnit(req.Series.SampleIntervalUnit),
			Value: req.Series.SampleIntervalValue,
		}, req.Series.CreatedAt.Truncate(time.Hour*24))

		searchPlan := compressionPlan.FilterFrames(ctx, frames, req.Repo.ID)

		mu := &sync.Mutex{}

		groupContext, groupCancel := context.WithCancel(ctx)
		defer groupCancel()
		g := group.New().WithContext(groupContext).WithMaxConcurrency(searchJobWorkerLimit).WithCancelOnError()
		for i := len(searchPlan.Executions) - 1; i >= 0; i-- {
			execution := searchPlan.Executions[i]
			g.Go(func(ctx context.Context) error {
				// Build historical data for this unique timeframe+repo+series.
				err, job, _ := buildJob(ctx, &buildSeriesContext{
					execution:       execution,
					repoName:        req.Repo.Name,
					id:              req.Repo.ID,
					firstHEADCommit: firstHEADCommit,
					seriesID:        req.Series.SeriesID,
					series:          req.Series,
				})
				mu.Lock()
				defer mu.Unlock()
				if job != nil {
					jobs = append(jobs, job)
				}
				return err
			})
		}
		err = g.Wait()
		if err != nil {
			jobs = nil
		}
		return ctx, &reqContext, jobs, err
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
	series   *types.InsightSeries
}

type searchJobFunc func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.SearchJob, preempted []store.RecordSeriesPointArgs)

func makeHistoricalSearchJobFunc(logger log.Logger, commitClient GitCommitClient) searchJobFunc {
	return func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.SearchJob, preempted []store.RecordSeriesPointArgs) {
		logger.Debug("making search job")
		rawQuery := bctx.series.Query
		containsRepo, err := querybuilder.ContainsField(rawQuery, query.FieldRepo)
		if err != nil {
			return err, nil, nil
		}
		if containsRepo {
			// This maintains existing behavior that searches with a repo filter are ignored
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
		recentCommits, err := commitClient.RecentCommits(ctx, bctx.repoName, bctx.execution.RecordingTime)
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
		modifiedQuery, err := querybuilder.SingleRepoQuery(querybuilder.BasicQuery(rawQuery), repoName, revision, querybuilder.CodeInsightsQueryDefaults(len(bctx.series.Repositories) == 0))
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

		job = &queryrunner.SearchJob{
			SeriesID:        bctx.seriesID,
			SearchQuery:     newQueryStr,
			RecordTime:      &bctx.execution.RecordingTime,
			PersistMode:     string(store.RecordMode),
			DependentFrames: bctx.execution.SharedRecordings,
		}
		return err, job, preempted
	}
}

func makeRunSearchFunc(logger log.Logger, searchHandlers map[types.GenerationMethod]queryrunner.InsightsHandler, searchWorkerLimit int) SearchRunner {
	return func(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.SearchJob, incomingErr error) (context.Context, *requestContext, []store.RecordSeriesPointArgs, error) {
		points := make([]store.RecordSeriesPointArgs, 0, len(jobs))
		// early return
		if incomingErr != nil || ctx.Err() != nil {
			return ctx, reqContext, points, incomingErr
		}
		series := reqContext.backfillRequest.Series
		mu := &sync.Mutex{}
		groupContext, groupCancel := context.WithCancel(ctx)
		defer groupCancel()
		g := group.New().WithContext(groupContext).WithMaxConcurrency(searchWorkerLimit).WithCancelOnError()
		for i := 0; i < len(jobs); i++ {
			job := jobs[i]
			g.Go(func(ctx context.Context) error {
				h := searchHandlers[series.GenerationMethod]
				searchPoints, err := h(ctx, job, series, *job.RecordTime)
				if err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				points = append(points, searchPoints...)
				return nil
			})
		}
		err := g.Wait()
		// don't return any points if they don't all succeed
		if err != nil {
			points = nil
		}
		return ctx, reqContext, points, err
	}
}

func makeSaveResultsFunc(logger log.Logger, insightStore store.Interface) ResultsPersister {
	return func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs, incomingErr error) (*requestContext, error) {
		if incomingErr != nil {
			return reqContext, incomingErr
		}
		if ctx.Err() != nil {
			return reqContext, ctx.Err()
		}
		logger.Debug("writing search results")
		err := insightStore.RecordSeriesPoints(ctx, points)
		return reqContext, err
	}

}
