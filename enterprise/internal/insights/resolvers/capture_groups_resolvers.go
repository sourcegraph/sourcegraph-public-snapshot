package resolvers

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/vcs/git"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
)

// type CaptureGroupResultsResolver interface {
// 	Groups(ctx context.Context) ([]CaptureGroupResolver, error)
// }
//
// type CaptureGroupResolver interface {
// 	Repo(ctx context.Context) string
// 	Commit(ctx context.Context) string
// 	Matches(ctx context.Context) []CaptureGroupMatchResolver
// }
// type CaptureGroupMatchResolver interface {
// 	Value(ctx context.Context) string
// 	Count(ctx context.Context) int
// }

// FirstOfMonthFrames builds a set of frames with a specific number of elements, such that all of the
// starting times of each frame < current will fall on the first of a month.
func FirstOfMonthFrames(numPoints int, current time.Time) []compression.Frame {
	if numPoints < 1 {
		return nil
	}
	times := make([]time.Time, 0, numPoints)
	year, month, _ := current.Date()
	firstOfCurrent := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	for i := 0 - numPoints + 1; i < 0; i++ {
		times = append(times, firstOfCurrent.AddDate(0, i, 0))
	}
	times = append(times, firstOfCurrent)
	times = append(times, current)

	frames := make([]compression.Frame, 0, len(times)-1)
	for i := 1; i < len(times); i++ {
		prev := times[i-1]
		frames = append(frames, compression.Frame{
			From: prev,
			To:   times[i],
		})
	}
	return frames
}

type CaptureGroupExecutor struct {
	repoStore database.RepoStore
	filter    compression.DataFrameFilter
	clock     func() time.Time
}

func NewCaptureGroupExecutor(ctx context.Context, postgres, insightsDb dbutil.DB, clock func() time.Time) *CaptureGroupExecutor {
	return &CaptureGroupExecutor{
		repoStore: database.Repos(postgres),
		// filter:    compression.NewHistoricalFilter(true, clock().Add(time.Hour*24*365*-1), insightsDb),
		filter: &compression.NoopFilter{},
		clock:  clock,
	}
}

func (c *CaptureGroupExecutor) Execute(ctx context.Context, query string, repositories []string) ([]livePreviewTimeSeries, error) {
	repoIds := make(map[string]api.RepoID)
	for _, repository := range repositories {
		repo, err := c.repoStore.GetByName(ctx, api.RepoName(repository))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch repository information for repository name: %s", repository)
		}
		repoIds[repository] = repo.ID
	}
	log15.Info("Generated repoIds", "repoids", repoIds)

	frames := FirstOfMonthFrames(7, c.clock())

	type timeCounts map[time.Time]int
	pivoted := make(map[string]timeCounts)

	for _, repository := range repositories {
		firstCommit, err := git.FirstEverCommit(ctx, api.RepoName(repository))
		if err != nil {
			return nil, errors.Wrapf(err, "FirstEverCommit")
		}
		// plan := c.filter.FilterFrames(ctx, frames, repoIds[repository])
		// uncompressed plan for now, because there is some complication between the way compressed plans are generated and needing to resolve revhashes
		plan := c.filter.FilterFrames(ctx, frames, repoIds[repository])

		generateTimes := func() map[time.Time]int {
			times := make(map[time.Time]int)
			for _, execution := range plan.Executions {
				times[execution.RecordingTime] = 0
				for _, recording := range execution.SharedRecordings {
					times[recording] = 0
				}
			}
			return times
		}
		for _, execution := range plan.Executions {
			log15.Info("Generated execution for live search query plan", "repository", repository, "execution", execution)
		}

		// we need to perform the pivot
		// var pivoted map[string]timeCounts
		for _, execution := range plan.Executions {
			if execution.RecordingTime.Before(firstCommit.Committer.Date) {
				// this logic is faulty, but works for now. If the plan was compressed (these executions had children) we would have to
				// iterate over the children to ensure they are also all before the first commit date. Otherwise, we would have to promote
				// that child to the new execution, and all of the remaining children (after the promoted one) become children of the new execution.
				// since we are using uncompressed plans (to avoid this problem and others) right now, each execution is standalone
				continue
			}

			commits, err := git.Commits(ctx, api.RepoName(repository), git.CommitsOptions{N: 1, Before: execution.RecordingTime.Format(time.RFC3339), DateOrder: true})
			if err != nil {
				return nil, errors.Wrap(err, "git.Commits")
			} else if len(commits) < 1 {
				// there is no commit so skip this execution. Once again faulty logic for the same reasons as above.
				continue
			}

			modifiedQuery := withCountUnlimited(query)
			modifiedQuery = fmt.Sprintf("%s repo:^%s$@%s", modifiedQuery, regexp.QuoteMeta(repository), commits[0].ID)

			log15.Info("executing query", "query", modifiedQuery)
			results, err := queryrunner.ComputeSearch(ctx, modifiedQuery)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to execute capture group search for repository:%s commit:%s", repository, execution.Revision)
			}

			grouped := queryrunner.GroupIt(results)
			sort.Slice(grouped, func(i, j int) bool {
				return grouped[i].Value < grouped[j].Value
			})
			log15.Info("grouped results", "grouped", grouped)

			for _, timeGroupElement := range grouped {
				value := timeGroupElement.Value
				if _, ok := pivoted[value]; !ok {
					pivoted[value] = generateTimes()
				}
				pivoted[value][execution.RecordingTime] = timeGroupElement.Count
				for _, children := range execution.SharedRecordings {
					pivoted[value][children] += timeGroupElement.Count
				}
			}
		}
	}

	var calculated []livePreviewTimeSeries
	seriesCount := 1
	for value, timeCounts := range pivoted {
		var timeseries []queryrunner.TimeDataPoint

		for key, val := range timeCounts {
			timeseries = append(timeseries, queryrunner.TimeDataPoint{
				Time:  key,
				Count: val,
			})
		}

		sort.Slice(timeseries, func(i, j int) bool {
			return timeseries[i].Time.Before(timeseries[j].Time)
		})

		// resolvers = append(resolvers, &captureGroupResultsResolver{results: timeseries, value: value})
		calculated = append(calculated, livePreviewTimeSeries{
			label:    value,
			points:   timeseries,
			seriesId: fmt.Sprintf("livepreview %d", seriesCount),
		})
		seriesCount++
	}
	return calculated, nil
}

func (r *Resolver) CaptureGroup(ctx context.Context, args graphqlbackend.CaptureGroupArgs) ([]graphqlbackend.CaptureGroupResultsResolver, error) {

	var resolvers []graphqlbackend.CaptureGroupResultsResolver

	return resolvers, nil
}

func withCountUnlimited(s string) string {
	if strings.Contains(s, "count:") {
		return s
	}
	return s + " count:all"
}

// to generate a time series:
// execute one search per commit
// for each results, save time with search results
// group by matched values
// sort each series by time
// result is array of timeseries

func (r *disabledResolver) CaptureGroup(ctx context.Context, args graphqlbackend.CaptureGroupArgs) ([]graphqlbackend.CaptureGroupResultsResolver, error) {
	return nil, errors.New(r.reason)
}

type captureGroupResultsResolver struct {
	results []queryrunner.TimeDataPoint
	value   string
}

func (c *captureGroupResultsResolver) Value(ctx context.Context) string {
	return c.value
}

// func (c *captureGroupResultsResolver) Groups(ctx context.Context) ([]graphqlbackend.CaptureGroupResolver, error) {
// 	var resolvers []graphqlbackend.CaptureGroupResolver
// 	for _, result := range c.results {
// 		resolvers = append(resolvers, &captureGroupResolver{result: result})
// 	}
// 	return resolvers, nil
// }

func (c *captureGroupResultsResolver) Groups(ctx context.Context) ([]graphqlbackend.CaptureGroupMatchResolver, error) {
	var resolvers []graphqlbackend.CaptureGroupMatchResolver
	for _, result := range c.results {
		resolvers = append(resolvers, &captureGroupMatchResolver{time: result.Time, count: int32(result.Count)})
	}
	return resolvers, nil
}

// type captureGroupResolver struct {
// 	result queryrunner.GroupedResults
// }
//
// func (c *captureGroupResolver) Repo(ctx context.Context) string {
// 	return ""
// }
//
// func (c *captureGroupResolver) Matches(ctx context.Context) []graphqlbackend.CaptureGroupMatchResolver {
// 	var resolvers []graphqlbackend.CaptureGroupMatchResolver
// 	return []captureGroupMatchResolver{{value: c.result.Value, count: int32(c.result.Count)}}
// 	return resolvers
// }

type captureGroupMatchResolver struct {
	time  time.Time
	count int32
}

func (c *captureGroupMatchResolver) Time(ctx context.Context) graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: c.time}
}

func (c *captureGroupMatchResolver) Count(ctx context.Context) int32 {
	return c.count
}

func (r *Resolver) SearchInsightLivePreview(ctx context.Context, args graphqlbackend.SearchInsightLivePreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	if !args.Input.GeneratedFromCaptureGroups {
		return nil, errors.New("live preview is currently only supported for generated series from capture groups")
	}

	executor := NewCaptureGroupExecutor(ctx, r.postgresDB, r.insightsDB, time.Now)
	generatedSeries, err := executor.Execute(ctx, args.Input.Query, args.Input.RepositoryScope.Repositories)
	if err != nil {
		return nil, err
	}

	var resolvers []graphqlbackend.SearchInsightLivePreviewSeriesResolver
	for i := range generatedSeries {
		resolvers = append(resolvers, &searchInsightLivePreviewSeriesResolver{series: &generatedSeries[i]})
	}

	return resolvers, nil
}

func (r *disabledResolver) SearchInsightLivePreview(ctx context.Context, args graphqlbackend.SearchInsightLivePreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	return nil, errors.New(r.reason)
}

type searchInsightLivePreviewSeriesResolver struct {
	series *livePreviewTimeSeries
}

func (s *searchInsightLivePreviewSeriesResolver) Points(ctx context.Context) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var resolvers []graphqlbackend.InsightsDataPointResolver
	for _, point := range s.series.points {
		resolvers = append(resolvers, &insightsDataPointResolver{store.SeriesPoint{
			SeriesID: s.series.seriesId,
			Time:     point.Time,
			Value:    float64(point.Count),
		}})
	}
	return resolvers, nil
}

func (s *searchInsightLivePreviewSeriesResolver) Label(ctx context.Context) (string, error) {
	return s.series.label, nil
}

type livePreviewTimeSeries struct {
	label    string
	points   []queryrunner.TimeDataPoint
	seriesId string
}
