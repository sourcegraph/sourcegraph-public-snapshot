package query

import (
	"context"
	"sort"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type StreamingQueryExecutor struct {
	justInTimeExecutor
}

func NewStreamingExecutor(postgres database.DB, clock func() time.Time) *StreamingQueryExecutor {
	return &StreamingQueryExecutor{
		justInTimeExecutor: justInTimeExecutor{
			db:        postgres,
			repoStore: postgres.Repos(),
			filter:    &compression.NoopFilter{},
			clock:     clock,
		},
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
	log15.Debug("Generated repoIds", "repoids", repoIds)

	frames := BuildFrames(7, interval, c.clock().Truncate(time.Hour*24))
	points := timeCounts{}
	timeseries := []TimeDataPoint{}

	for _, repository := range repositories {
		firstCommit, err := discovery.GitFirstEverCommit(ctx, c.db, api.RepoName(repository))
		if err != nil {
			if errors.Is(err, discovery.EmptyRepoErr) {
				continue
			} else {
				return nil, errors.Wrapf(err, "FirstEverCommit")
			}
		}
		// uncompressed plan for now, because there is some complication between the way compressed plans are generated and needing to resolve revhashes
		plan := c.filter.FilterFrames(ctx, frames, repoIds[repository])

		// we need to perform the pivot from time -> {label, count} to label -> {time, count}
		for _, execution := range plan.Executions {
			if execution.RecordingTime.Before(firstCommit.Committer.Date) {
				// this logic is faulty, but works for now. If the plan was compressed (these executions had children) we would have to
				// iterate over the children to ensure they are also all before the first commit date. Otherwise, we would have to promote
				// that child to the new execution, and all of the remaining children (after the promoted one) become children of the new execution.
				// since we are using uncompressed plans (to avoid this problem and others) right now, each execution is standalone
				continue
			}
			commits, err := gitserver.NewClient(c.db).Commits(ctx, api.RepoName(repository), gitserver.CommitsOptions{N: 1, Before: execution.RecordingTime.Format(time.RFC3339), DateOrder: true}, authz.DefaultSubRepoPermsChecker)
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
			err = streaming.Search(ctx, modified.String(), nil, decoder)
			if err != nil {
				return nil, errors.Wrap(err, "streaming.Search")
			}

			tr := *tabulationResult
			if len(tr.SkippedReasons) > 0 {
				log15.Error("insights query issue", "reasons", tr.SkippedReasons, "query", query)
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
		timeseries = append(timeseries, TimeDataPoint{
			Time:  pointTime,
			Count: pointCount,
		})
	}

	sort.Slice(timeseries, func(i, j int) bool {
		return timeseries[i].Time.Before(timeseries[j].Time)
	})
	generated := []GeneratedTimeSeries{{
		Label:    seriesLabel,
		SeriesId: seriesID,
		Points:   timeseries,
	}}
	return generated, nil

}
