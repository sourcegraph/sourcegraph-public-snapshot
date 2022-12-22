package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const maxPreviewRepos = 20

func (r *Resolver) SearchInsightLivePreview(ctx context.Context, args graphqlbackend.SearchInsightLivePreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	if !args.Input.GeneratedFromCaptureGroups {
		return nil, errors.New("live preview is currently only supported for generated series from capture groups")
	}
	previewArgs := graphqlbackend.SearchInsightPreviewArgs{
		Input: graphqlbackend.SearchInsightPreviewInput{
			RepositoryScope: args.Input.RepositoryScope,
			TimeScope:       args.Input.TimeScope,
			Series: []graphqlbackend.SearchSeriesPreviewInput{
				{
					Query:                      args.Input.Query,
					Label:                      args.Input.Label,
					GeneratedFromCaptureGroups: args.Input.GeneratedFromCaptureGroups,
					GroupBy:                    args.Input.GroupBy,
				},
			},
		},
	}
	return r.SearchInsightPreview(ctx, previewArgs)
}

func (r *Resolver) SearchInsightPreview(ctx context.Context, args graphqlbackend.SearchInsightPreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {

	err := isValidPreviewArgs(args)
	if err != nil {
		return nil, err
	}

	var resolvers []graphqlbackend.SearchInsightLivePreviewSeriesResolver
	var generatedSeries []query.GeneratedTimeSeries

	// get a consistent time to use across all preview series
	previewTime := time.Now().UTC()
	clock := func() time.Time {
		return previewTime
	}
	interval := timeseries.TimeInterval{
		Unit:  types.IntervalUnit(args.Input.TimeScope.StepInterval.Unit),
		Value: int(args.Input.TimeScope.StepInterval.Value),
	}

	repos, err := getPreviewRepos(ctx, args.Input.RepositoryScope, r.logger)
	if err != nil {
		return nil, err
	}
	if len(repos) > maxPreviewRepos {
		return nil, errors.Newf("live preview is limited to %d repositories", maxPreviewRepos)
	}

	for _, seriesArgs := range args.Input.Series {

		var series []query.GeneratedTimeSeries
		var err error

		if seriesArgs.GeneratedFromCaptureGroups {
			if seriesArgs.GroupBy != nil {
				executor := query.NewComputeExecutor(r.postgresDB, clock)
				series, err = executor.Execute(ctx, seriesArgs.Query, *seriesArgs.GroupBy, repos)
				if err != nil {
					return nil, err
				}
			} else {
				executor := query.NewCaptureGroupExecutor(r.postgresDB, clock)
				series, err = executor.Execute(ctx, seriesArgs.Query, repos, interval)
				if err != nil {
					return nil, err
				}
			}
		} else {
			executor := query.NewStreamingExecutor(r.postgresDB, clock)
			series, err = executor.Execute(ctx, seriesArgs.Query, seriesArgs.Label, seriesArgs.Label, repos, interval)
			if err != nil {
				return nil, err
			}
		}
		generatedSeries = append(generatedSeries, series...)
	}

	foundData := false
	for i := range generatedSeries {
		foundData = foundData || len(generatedSeries[i].Points) > 0
		resolvers = append(resolvers, &searchInsightLivePreviewSeriesResolver{series: &generatedSeries[i]})
	}
	if !foundData {
		return nil, errors.Newf("Data for %s not found", pluralize("this repository", "these repositories", len(repos)))
	}

	return resolvers, nil
}

func pluralize(singular, plural string, n int) string {
	if n == 1 {
		return singular
	}
	return plural
}

type searchInsightLivePreviewSeriesResolver struct {
	series *query.GeneratedTimeSeries
}

func (s *searchInsightLivePreviewSeriesResolver) Points(ctx context.Context) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var resolvers []graphqlbackend.InsightsDataPointResolver
	for _, point := range s.series.Points {
		resolvers = append(resolvers, &insightsDataPointResolver{store.SeriesPoint{
			SeriesID: s.series.SeriesId,
			Time:     point.Time,
			Value:    float64(point.Count),
		}})
	}
	return resolvers, nil
}

func (s *searchInsightLivePreviewSeriesResolver) Label(ctx context.Context) (string, error) {
	return s.series.Label, nil
}

func getPreviewRepos(ctx context.Context, repoScope graphqlbackend.RepositoryScopeInput, logger log.Logger) ([]string, error) {
	var repos []string
	if repoScope.RepositoryCriteria != nil {
		repoQueryExecutor := query.NewStreamingRepoQueryExecutor(logger.Scoped("live_preview_resolver", ""))
		repoQuery, err := querybuilder.RepositoryScopeQuery(*repoScope.RepositoryCriteria)
		if err != nil {
			return nil, err
		}
		// Since preview is not allowed over "max_preview_repos" limit result set to avoid processing more results than neccessary
		limitedRepoQuery, err := repoQuery.WithCount(fmt.Sprintf("%d", maxPreviewRepos+1))
		if err != nil {
			return nil, err
		}
		repoList, err := repoQueryExecutor.ExecuteRepoList(ctx, string(limitedRepoQuery))
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(repoList); i++ {
			repos = append(repos, string(repoList[i].Name))
		}
	} else {
		repos = repoScope.Repositories
	}
	return repos, nil
}

func isValidPreviewArgs(args graphqlbackend.SearchInsightPreviewArgs) error {
	if args.Input.TimeScope.StepInterval == nil {
		return errors.New("live preview currently only supports a time interval time scope")
	}
	hasRepoCriteria := args.Input.RepositoryScope.RepositoryCriteria != nil
	// Error if both are provided
	if hasRepoCriteria && len(args.Input.RepositoryScope.Repositories) > 0 {
		return errors.New("can not specify both a repository list and a repository search")
	}

	if hasRepoCriteria {
		for i := 0; i < len(args.Input.Series); i++ {
			if args.Input.Series[i].GroupBy != nil {
				return errors.New("group by insights do not support selecting repositories using a search")
			}
		}
	}

	return nil
}
