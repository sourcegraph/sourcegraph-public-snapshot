package query

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"

	"github.com/sourcegraph/sourcegraph/internal/conf"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CaptureGroupExecutor struct {
	justInTimeExecutor
	computeSearch func(ctx context.Context, query string) ([]GroupedResults, error)
}

func NewCaptureGroupExecutor(postgres, insightsDb dbutil.DB, clock func() time.Time) *CaptureGroupExecutor {
	executor := CaptureGroupExecutor{
		justInTimeExecutor: justInTimeExecutor{
			db:        database.NewDB(postgres),
			repoStore: database.Repos(postgres),
			// filter:    compression.NewHistoricalFilter(true, clock().Add(time.Hour*24*365*-1), insightsDb),
			filter: &compression.NoopFilter{},
			clock:  clock,
		},
		computeSearch: streamCompute,
	}

	useGraphQL := conf.Get().InsightsComputeGraphql
	if useGraphQL != nil && *useGraphQL {
		executor.computeSearch = graphQLCompute
	}

	return &executor
}

func graphQLCompute(ctx context.Context, query string) ([]GroupedResults, error) {
	searchResults, err := ComputeSearch(ctx, query)
	if err != nil {
		return nil, err
	}
	return GroupByCaptureMatch(searchResults), nil
}

func streamCompute(ctx context.Context, query string) ([]GroupedResults, error) {
	decoder, streamResults := streaming.ComputeDecoder()
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
	log15.Debug("Generated repoIds", "repoids", repoIds)

	frames := BuildFrames(7, interval, c.clock())
	pivoted := make(map[string]timeCounts)

	for _, repository := range repositories {
		firstCommit, err := git.FirstEverCommit(ctx, c.db, api.RepoName(repository), authz.DefaultSubRepoPermsChecker)
		if err != nil {
			return nil, errors.Wrapf(err, "FirstEverCommit")
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
			commits, err := git.Commits(ctx, c.db, api.RepoName(repository), git.CommitsOptions{N: 1, Before: execution.RecordingTime.Format(time.RFC3339), DateOrder: true}, authz.DefaultSubRepoPermsChecker)
			if err != nil {
				return nil, errors.Wrap(err, "git.Commits")
			} else if len(commits) < 1 {
				// there is no commit so skip this execution. Once again faulty logic for the same reasons as above.
				continue
			}

			modifiedQuery, err := querybuilder.SingleRepoQuery(query, repository, string(commits[0].ID))
			if err != nil {
				return nil, errors.Wrap(err, "SingleRepoQuery")
			}

			log15.Debug("executing query", "query", modifiedQuery)
			grouped, err := c.computeSearch(ctx, modifiedQuery)
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
