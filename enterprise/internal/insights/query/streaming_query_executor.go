package query

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	itypes "github.com/sourcegraph/sourcegraph/internal/types"
)

type StreamingQueryExecutor struct {
	justInTimeExecutor

	logger log.Logger
}

func NewStreamingExecutor(postgres database.DB, clock func() time.Time) *StreamingQueryExecutor {
	return &StreamingQueryExecutor{
		justInTimeExecutor: justInTimeExecutor{
			db:        postgres,
			repoStore: postgres.Repos(),
			filter:    &compression.NoopFilter{},
			clock:     clock,
		},
		logger: log.Scoped("StreamingQueryExecutor", ""),
	}
}

func (c *StreamingQueryExecutor) ExecuteRepoList(ctx context.Context, query string) ([]itypes.MinimalRepo, error) {
	decoder, selectRepoResult := streaming.SelectRepoDecoder()
	err := streaming.Search(ctx, query, nil, decoder)
	if err != nil {
		return nil, errors.Wrap(err, "streaming.Search")
	}

	repoResult := *selectRepoResult
	if len(repoResult.SkippedReasons) > 0 {
		c.logger.Error("insights query issue", log.String("reasons", fmt.Sprintf("%v", repoResult.SkippedReasons)), log.String("query", query))
	}
	if len(repoResult.Errors) > 0 {
		return nil, errors.Errorf("streaming search: errors: %v", repoResult.Errors)
	}
	if len(repoResult.Alerts) > 0 {
		return nil, errors.Errorf("streaming search: alerts: %v", repoResult.Alerts)
	}

	return repoResult.Repos, nil
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

	frames := timeseries.BuildSampleTimes(7, interval, c.clock().Truncate(time.Minute))
	points := timeCounts{}
	timeDataPoints := []TimeDataPoint{}

	for _, repository := range repositories {
		firstCommit, err := gitserver.GitFirstEverCommit(ctx, c.db, api.RepoName(repository))
		if err != nil {
			if errors.Is(err, gitserver.EmptyRepoErr) {
				continue
			} else {
				return nil, errors.Wrapf(err, "FirstEverCommit")
			}
		}
		// uncompressed plan for now, because there is some complication between the way compressed plans are generated and needing to resolve revhashes
		plan := c.filter.Filter(ctx, frames, api.RepoName(repository))

		// we need to perform the pivot from time -> {label, count} to label -> {time, count}
		for _, execution := range plan.Executions {
			if execution.RecordingTime.Before(firstCommit.Committer.Date) {
				// this logic is faulty, but works for now. If the plan was compressed (these executions had children) we would have to
				// iterate over the children to ensure they are also all before the first commit date. Otherwise, we would have to promote
				// that child to the new execution, and all of the remaining children (after the promoted one) become children of the new execution.
				// since we are using uncompressed plans (to avoid this problem and others) right now, each execution is standalone
				continue
			}
			commits, err := gitserver.NewGitCommitClient(c.db).RecentCommits(ctx, api.RepoName(repository), execution.RecordingTime, "")
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
