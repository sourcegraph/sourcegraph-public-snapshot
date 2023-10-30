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
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CaptureGroupExecutor struct {
	gitserverClient internalGitserver.Client
	previewExecutor
	computeSearch func(ctx context.Context, query string) ([]GroupedResults, error)

	logger log.Logger
}

func NewCaptureGroupExecutor(db database.DB, clock func() time.Time) *CaptureGroupExecutor {
	return &CaptureGroupExecutor{
		gitserverClient: internalGitserver.NewClient("insights.capturegroupexecutor"),
		previewExecutor: previewExecutor{
			repoStore: db.Repos(),
			// filter:    compression.NewHistoricalFilter(true, clock().Add(time.Hour*24*365*-1), insightsDb),
			filter: &compression.NoopFilter{},
			clock:  clock,
		},
		computeSearch: streamCompute,
		logger:        log.Scoped("CaptureGroupExecutor"),
	}
}

func streamCompute(ctx context.Context, query string) ([]GroupedResults, error) {
	decoder, streamResults := streaming.MatchContextComputeDecoder()
	err := streaming.ComputeMatchContextStream(ctx, query, decoder)
	if err != nil {
		return nil, err
	}
	if len(streamResults.Errors) > 0 {
		return nil, errors.Errorf("compute streaming search: errors: %v", streamResults.Errors)
	}
	if len(streamResults.Alerts) > 0 {
		return nil, errors.Errorf("compute streaming search: alerts: %v", streamResults.Alerts)
	}
	return computeTabulationResultToGroupedResults(streamResults), nil
}

func (c *CaptureGroupExecutor) Execute(ctx context.Context, query string, repositories []string, interval timeseries.TimeInterval) ([]GeneratedTimeSeries, error) {
	repoIds := make(map[string]api.RepoID)
	for _, repository := range repositories {
		repo, err := c.repoStore.GetByName(ctx, api.RepoName(repository))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch repository information for repository name: %s", repository)
		}
		repoIds[repository] = repo.ID
	}
	c.logger.Debug("Generated repoIds", log.String("repoids", fmt.Sprintf("%v", repoIds)))

	sampleTimes := timeseries.BuildSampleTimes(7, interval, c.clock())
	pivoted := make(map[string]timeCounts)

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

			modifiedQuery, err := querybuilder.SingleRepoQuery(querybuilder.BasicQuery(query), repository, string(commits[0].ID), querybuilder.CodeInsightsQueryDefaults(false))
			if err != nil {
				return nil, errors.Wrap(err, "query validation")
			}

			c.logger.Debug("executing query", log.String("query", modifiedQuery.String()))
			grouped, err := c.computeSearch(ctx, modifiedQuery.String())
			if err != nil {
				errorMsg := "failed to execute capture group search for repository:" + repository
				if execution.Revision != "" {
					errorMsg += " commit:" + execution.Revision
				}
				return nil, errors.Wrap(err, errorMsg)
			}

			sort.Slice(grouped, func(i, j int) bool {
				return grouped[i].Value < grouped[j].Value
			})

			for _, timeGroupElement := range grouped {
				value := timeGroupElement.Value
				if _, ok := pivoted[value]; !ok {
					pivoted[value] = generateTimes(plan)
				}
				pivoted[value][execution.RecordingTime] += timeGroupElement.Count
				for _, children := range execution.SharedRecordings {
					pivoted[value][children] += timeGroupElement.Count
				}
			}
		}
	}

	calculated := makeTimeSeries(pivoted)
	return calculated, nil
}

func makeTimeSeries(pivoted map[string]timeCounts) []GeneratedTimeSeries {
	var calculated []GeneratedTimeSeries
	seriesCount := 1
	for value, timeCounts := range pivoted {
		var ts []TimeDataPoint

		for key, val := range timeCounts {
			ts = append(ts, TimeDataPoint{
				Time:  key,
				Count: val,
			})
		}

		sort.Slice(ts, func(i, j int) bool {
			return ts[i].Time.Before(ts[j].Time)
		})

		calculated = append(calculated, GeneratedTimeSeries{
			Label:    value,
			Points:   ts,
			SeriesId: fmt.Sprintf("dynamic-series-%d", seriesCount),
		})
		seriesCount++
	}
	return calculated
}

func computeTabulationResultToGroupedResults(result *streaming.ComputeTabulationResult) []GroupedResults {
	var grouped []GroupedResults
	for _, match := range result.RepoCounts {
		for value, count := range match.ValueCounts {
			grouped = append(grouped, GroupedResults{
				Value: value,
				Count: count,
			})
		}
	}
	return grouped
}
