package resolvers

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

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

func (r *Resolver) CaptureGroup(ctx context.Context, args graphqlbackend.CaptureGroupArgs) ([]graphqlbackend.CaptureGroupResultsResolver, error) {

	// var counts [][]queryrunner.GroupedResults
	//
	// var byTime map[string][]queryrunner.TimeDataPoint

	repoStore := database.Repos(r.postgresDB)
	repo, err := repoStore.GetByName(ctx, api.RepoName(args.Repo))
	if err != nil {
		return nil, err
	}

	frames := FirstOfMonthFrames(7, time.Now())
	filter := compression.NewHistoricalFilter(true, time.Now().Add(time.Hour*24*365*-1), r.postgresDB)
	plan := filter.FilterFrames(ctx, frames, repo.ID)

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

	type timeCounts map[time.Time]int
	// we need to perform the pivot
	// var pivoted map[string]timeCounts
	pivoted := make(map[string]timeCounts)
	for _, execution := range plan.Executions {
		query := withCountUnlimited(args.Query)
		query = fmt.Sprintf("%s repo:^%s$@%s", query, regexp.QuoteMeta(args.Repo), execution.Revision)

		results, err := queryrunner.ComputeSearch(ctx, args.Query)
		if err != nil {
			return nil, err
		}

		grouped := queryrunner.GroupIt(results)
		sort.Slice(grouped, func(i, j int) bool {
			return grouped[i].Value < grouped[j].Value
		})

		for _, timeGroupElement := range grouped {
			value := timeGroupElement.Value
			if _, ok := pivoted[value]; !ok {
				pivoted[value] = generateTimes()
			}
			pivoted[value][execution.RecordingTime] = timeGroupElement.Count
			for _, children := range execution.SharedRecordings {
				pivoted[value][children] = timeGroupElement.Count
			}
		}
	}

	var resolvers []graphqlbackend.CaptureGroupResultsResolver

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

		resolvers = append(resolvers, &captureGroupResultsResolver{results: timeseries, value: value})
	}

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
