package query

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ComputeExecutor struct {
	previewExecutor
	logger        log.Logger
	computeSearch func(ctx context.Context, query string) ([]GroupedResults, error)
}

func NewComputeExecutor(postgres database.DB, clock func() time.Time) *ComputeExecutor {
	executor := ComputeExecutor{
		logger: log.Scoped("ComputeExecutor"),
		previewExecutor: previewExecutor{
			repoStore: postgres.Repos(),
			filter:    &compression.NoopFilter{},
			clock:     clock,
		},
		computeSearch: streamTextExtraCompute,
	}

	return &executor
}

func streamTextExtraCompute(ctx context.Context, query string) ([]GroupedResults, error) {
	decoder, streamResults := streaming.ComputeTextDecoder()
	err := streaming.ComputeTextExtraStream(ctx, query, decoder)
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

func (c *ComputeExecutor) Execute(ctx context.Context, query, groupBy string, repositories []string) ([]GeneratedTimeSeries, error) {
	repoIds := make(map[string]api.RepoID)
	for _, repository := range repositories {
		repo, err := c.repoStore.GetByName(ctx, api.RepoName(repository))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch repository information for repository name: %s", repository)
		}
		repoIds[repository] = repo.ID
	}

	gitserverClient := gitserver.NewClient("insights.computeexecutor")

	groupedValues := make(map[string]int)
	for _, repository := range repositories {
		modifiedQuery := querybuilder.SingleRepoQueryIndexed(querybuilder.BasicQuery(query), repository)
		finalQuery, err := querybuilder.ComputeInsightCommandQuery(modifiedQuery, querybuilder.MapType(strings.ToLower(groupBy)), gitserverClient)
		if err != nil {
			return nil, errors.Wrap(err, "query validation")
		}

		grouped, err := c.computeSearch(ctx, finalQuery.String())
		if err != nil {
			errorMsg := "failed to execute capture group search for repository:" + repository
			return nil, errors.Wrap(err, errorMsg)
		}

		sort.Slice(grouped, func(i, j int) bool {
			return grouped[i].Value < grouped[j].Value
		})

		for _, group := range grouped {
			groupedValues[group.Value] += group.Count
		}
	}

	timeSeries := []GeneratedTimeSeries{}
	seriesCount := 1
	now := time.Now()
	for label, value := range groupedValues {
		timeSeries = append(timeSeries, GeneratedTimeSeries{
			Label:    label,
			SeriesId: fmt.Sprintf("captured-series-%d", seriesCount),
			Points: []TimeDataPoint{{
				Time:  now,
				Count: value,
			}},
		})
		seriesCount++
	}
	return sortAndLimitComputedGroups(timeSeries), nil
}

// Simple sort/limit with reasonable defaults for v1.
func sortAndLimitComputedGroups(timeSeries []GeneratedTimeSeries) []GeneratedTimeSeries {
	descValueSort := func(i, j int) bool {
		if len(timeSeries[i].Points) == 0 || len(timeSeries[j].Points) == 0 {
			return false
		}
		return timeSeries[i].Points[0].Count > timeSeries[j].Points[0].Count
	}
	sort.SliceStable(timeSeries, descValueSort)
	limit := min(20, len(timeSeries))
	return timeSeries[:limit]
}
