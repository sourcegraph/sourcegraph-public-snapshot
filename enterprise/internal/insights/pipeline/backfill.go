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
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BackfillRequest struct {
	Series *types.InsightSeries
	Repo   *itypes.MinimalRepo
}

type Backfiller interface {
	Run(ctx context.Context, request BackfillRequest) error
}

type gitCommitClient interface {
	FirstCommit(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error)
	RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error)
}

type SearchJobGeneratorOutput struct {
	*BackfillRequest
	Job *queryrunner.Job
}

type SearchResultOutput struct {
	*BackfillRequest
	err    error
	result searchResult
}

type SearchJobGenerator func(ctx context.Context, req BackfillRequest) <-chan SearchJobGeneratorOutput
type SearchRunner func(ctx context.Context, input <-chan SearchJobGeneratorOutput) <-chan SearchResultOutput
type ResultsPersister func(ctx context.Context, input <-chan SearchResultOutput) error

func NewBackfiller(jobGenerator SearchJobGenerator, searchRunner SearchRunner, resultsPersister ResultsPersister) Backfiller {
	return &backfiller{
		searchJobGenerator: jobGenerator,
		searchRunner:       searchRunner,
		persister:          resultsPersister,
		logger:             log.Scoped("insights_backfill_pipeline", ""),
	}

}

type backfiller struct {
	//dependencies
	searchJobGenerator SearchJobGenerator
	searchRunner       SearchRunner
	persister          ResultsPersister
	logger             log.Logger
}

func (b *backfiller) Run(ctx context.Context, req BackfillRequest) error {

	jobsChan := b.searchJobGenerator(ctx, req)
	searchResultsChan := b.searchRunner(ctx, jobsChan)
	return b.persister(ctx, searchResultsChan)

}

// Implimentation of steps for Backfill process

func makeSearchJobsFunc(logger log.Logger, commitClient gitCommitClient, compressionPlan compression.DataFrameFilter) func(ctx context.Context, req BackfillRequest) <-chan SearchJobGeneratorOutput {
	return func(ctx context.Context, req BackfillRequest) <-chan SearchJobGeneratorOutput {
		output := make(chan SearchJobGeneratorOutput, 12)
		buildJob := makeHistoricalSearchJobFunc(logger, commitClient)

		var wg sync.WaitGroup
		go func() {
			wg.Wait()
			close(output)
		}()
		// launching a single worker
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()

			logger.Debug("making search plan")
			// Find the first commit made to the repository on the default branch.
			firstHEADCommit, err := commitClient.FirstCommit(ctx, req.Repo.Name)
			if err != nil {

				if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) || gitdomain.IsRepoNotExist(err) {
					//return nil, err // error - repo may not be cloned yet (or not even pushed to code host yet)
				}
				if errors.Is(err, discovery.EmptyRepoErr) {
					return
				}
				//TODO: deal with errors here
				//return nil, err
			}

			frames := timeseries.BuildFrames(12, timeseries.TimeInterval{
				Unit:  types.IntervalUnit(req.Series.SampleIntervalUnit),
				Value: req.Series.SampleIntervalValue,
			}, req.Series.CreatedAt.Truncate(time.Hour*24))

			searchPlan := compressionPlan.FilterFrames(ctx, frames, req.Repo.ID)
			for i := len(searchPlan.Executions) - 1; i >= 0; i-- {
				queryExecution := searchPlan.Executions[i]
				// Build historical data for this unique timeframe+repo+series.
				_, job, _ := buildJob(ctx, &buildSeriesContext{
					execution:       queryExecution,
					repoName:        req.Repo.Name,
					id:              req.Repo.ID,
					firstHEADCommit: firstHEADCommit,
					seriesID:        req.Series.SeriesID,
					series:          req.Series,
				})
				output <- SearchJobGeneratorOutput{BackfillRequest: &req, Job: job}
			}

		})
		return output

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

type searchJobFunc func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.Job, preempted []store.RecordSeriesPointArgs)

func makeHistoricalSearchJobFunc(logger log.Logger, commitClient gitCommitClient) searchJobFunc {
	return func(ctx context.Context, bctx *buildSeriesContext) (err error, job *queryrunner.Job, preempted []store.RecordSeriesPointArgs) {
		logger.Debug("making search job")
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
	repo        *itypes.MinimalRepo
	pointInTime time.Time
}

func makeRunSearchFunc(logger log.Logger, searchClient streaming.SearchClient) func(context.Context, <-chan SearchJobGeneratorOutput) <-chan SearchResultOutput {
	return func(ctx context.Context, in <-chan SearchJobGeneratorOutput) <-chan SearchResultOutput {

		out := make(chan SearchResultOutput)
		go func(ctx context.Context, outputChannel chan SearchResultOutput) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					golog.Printf("goroutine panic: %v\n%s", err, stack)
				}
				close(out)
			}()
			for r := range in {

				// run search
				// some made up values
				time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
				logger.Debug("running the search job")
				outputChannel <- SearchResultOutput{
					BackfillRequest: r.BackfillRequest,
					result:          searchResult{count: 10, capture: "", repo: r.BackfillRequest.Repo, pointInTime: *r.Job.RecordTime},
					err:             nil}
			}
		}(ctx, out)
		return out
	}
}

func makeSaveResultsFunc(logger log.Logger, insightStore store.Interface) func(ctx context.Context, in <-chan SearchResultOutput) error {
	return func(ctx context.Context, in <-chan SearchResultOutput) error {
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
					SeriesID: search.BackfillRequest.Series.SeriesID,
					Point: store.SeriesPoint{
						SeriesID: search.BackfillRequest.Series.SeriesID,
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
		logger.Debug("writing search results")
		return insightStore.RecordSeriesPoints(ctx, points)
	}

}
