package query

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	internalGitserver "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/internal/insights/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/internal/insights/timeseries"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type StreamingQueryExecutor struct {
	gitserverClient internalGitserver.Client
	previewExecutor

	logger log.Logger
}

func NewStreamingExecutor(db database.DB, clock func() time.Time) *StreamingQueryExecutor {
	return &StreamingQueryExecutor{
		gitserverClient: internalGitserver.NewClient("insights.queryexecutor"),
		previewExecutor: previewExecutor{
			repoStore: db.Repos(),
			filter:    &compression.NoopFilter{},
			clock:     clock,
		},
		logger: log.Scoped("StreamingQueryExecutor"),
	}
}

func (c *StreamingQueryExecutor) Execute(ctx context.Context, query string, seriesLabel string, seriesID string, repositories []string, interval timeseries.TimeInterval) ([]GeneratedTimeSeries, error) {
	repoIds := make(map[string]api.RepoID)
	for _, repository := range repositories {
		repo, err := c.repoStore.GetByName(ctx, api.RepoName(repository))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch repository information for repository name: %s", repository)
		}
		repoIds[repository] = repo.ID
	}
	c.logger.Debug("Generated repoIds", log.String("repoids", fmt.Sprintf("%v", repoIds)))

	sampleTimes := timeseries.BuildSampleTimes(7, interval, c.clock().Truncate(time.Minute))
	points := timeCounts{}
	timeDataPoints := []TimeDataPoint{}

	for _, repository := range repositories {
		firstCommit, err := gitserver.GitFirstEverCommit(ctx, c.gitserverClient, api.RepoName(repository))
		if err != nil {
			if errors.Is(err, gitserver.EmptyRepoErr) {
				continue
			} else {
				return nil, errors.Wrapf(err, "FirstEverCommit")
			}
		}
		// uncompressed plan for now, because there is some complication between the way compressed plans are generated and needing to resolve revhashes
		plan := c.filter.Filter(ctx, sampleTimes, api.RepoName(repository))

		// we need to perform the pivot from time -> {label, count} to label -> {time, count}
		for _, execution := range plan.Executions {
			if execution.RecordingTime.Before(firstCommit.Committer.Date) {
				// this logic is faulty, but works for now. If the plan was compressed (these executions had children) we would have to
				// iterate over the children to ensure they are also all before the first commit date. Otherwise, we would have to promote
				// that child to the new execution, and all of the remaining children (after the promoted one) become children of the new execution.
				// since we are using uncompressed plans (to avoid this problem and others) right now, each execution is standalone
				continue
			}
			commits, err := gitserver.NewGitCommitClient(c.gitserverClient).RecentCommits(ctx, api.RepoName(repository), execution.RecordingTime, "")
			if err != nil {
				return nil, errors.Wrap(err, "git.Commits")
			} else if len(commits) < 1 {
				// there is no commit so skip this execution. Once again faulty logic for the same reasons as above.
				continue
			}

			modified, err := querybuilder.SingleRepoQuery(querybuilder.BasicQuery(query), repository, string(commits[0].ID), querybuilder.CodeInsightsQueryDefaults(false))
			if err != nil {
				return nil, errors.Wrap(err, "query validation")
			}

			decoder, tabulationResult := streaming.TabulationDecoder()
			c.logger.Debug("executing query", log.String("query", modified.String()))
			err = streaming.Search(ctx, modified.String(), nil, decoder)
			if err != nil {
				return nil, errors.Wrap(err, "streaming.Search")
			}

			tr := *tabulationResult
			if len(tr.SkippedReasons) > 0 {
				c.logger.Error("insights query issue", log.String("reasons", fmt.Sprintf("%v", tr.SkippedReasons)), log.String("query", query))
			}
			if len(tr.Errors) > 0 {
				return nil, errors.Errorf("streaming search: errors: %v", tr.Errors)
			}
			if len(tr.Alerts) > 0 {
				return nil, errors.Errorf("streaming search: alerts: %v", tr.Alerts)
			}

			points[execution.RecordingTime] += tr.TotalCount
		}
	}

	for pointTime, pointCount := range points {
		timeDataPoints = append(timeDataPoints, TimeDataPoint{
			Time:  pointTime,
			Count: pointCount,
		})
	}

	sort.Slice(timeDataPoints, func(i, j int) bool {
		return timeDataPoints[i].Time.Before(timeDataPoints[j].Time)
	})
	generated := []GeneratedTimeSeries{{
		Label:    seriesLabel,
		SeriesId: seriesID,
		Points:   timeDataPoints,
	}}
	return generated, nil
}

type RepoQueryExecutor interface {
	ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimalRepo, error)
}

type StreamingRepoQueryExecutor struct {
	logger log.Logger
}

func NewStreamingRepoQueryExecutor(logger log.Logger) RepoQueryExecutor {
	return &StreamingRepoQueryExecutor{
		logger: logger,
	}
}

func (c *StreamingRepoQueryExecutor) ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimalRepo, error) {
	decoder, result := streaming.RepoDecoder()
	err := streaming.Search(ctx, query, nil, decoder)
	if err != nil {
		return nil, errors.Wrap(err, "RepoDecoder")
	}

	repoResult := *result
	if len(repoResult.SkippedReasons) > 0 {
		c.logger.Error("repo search encountered skipped events", log.String("reasons", fmt.Sprintf("%v", repoResult.SkippedReasons)), log.String("query", query))
	}
	if len(repoResult.Errors) > 0 {
		return nil, errors.Errorf("streaming repo search: errors: %v", repoResult.Errors)
	}
	if len(repoResult.Alerts) > 0 {
		return nil, errors.Errorf("streaming repo search: alerts: %v", repoResult.Alerts)
	}
	return repoResult.Repos, nil
}
