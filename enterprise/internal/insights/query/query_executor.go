package query

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CaptureGroupExecutor struct {
	repoStore database.RepoStore
	filter    compression.DataFrameFilter
	clock     func() time.Time
}

func NewCaptureGroupExecutor(postgres, insightsDb dbutil.DB, clock func() time.Time) *CaptureGroupExecutor {
	return &CaptureGroupExecutor{
		repoStore: database.Repos(postgres),
		// filter:    compression.NewHistoricalFilter(true, clock().Add(time.Hour*24*365*-1), insightsDb),
		filter: &compression.NoopFilter{},
		clock:  clock,
	}
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

	type timeCounts map[time.Time]int
	pivoted := make(map[string]timeCounts)

	for _, repository := range repositories {
		firstCommit, err := git.FirstEverCommit(ctx, api.RepoName(repository), authz.DefaultSubRepoPermsChecker)
		if err != nil {
			return nil, errors.Wrapf(err, "FirstEverCommit")
		}
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

		// we need to perform the pivot from time -> {label, count} to label -> {time, count}
		for _, execution := range plan.Executions {
			if execution.RecordingTime.Before(firstCommit.Committer.Date) {
				// this logic is faulty, but works for now. If the plan was compressed (these executions had children) we would have to
				// iterate over the children to ensure they are also all before the first commit date. Otherwise, we would have to promote
				// that child to the new execution, and all of the remaining children (after the promoted one) become children of the new execution.
				// since we are using uncompressed plans (to avoid this problem and others) right now, each execution is standalone
				continue
			}

			commits, err := git.Commits(ctx, api.RepoName(repository), git.CommitsOptions{N: 1, Before: execution.RecordingTime.Format(time.RFC3339), DateOrder: true}, authz.DefaultSubRepoPermsChecker)
			if err != nil {
				return nil, errors.Wrap(err, "git.Commits")
			} else if len(commits) < 1 {
				// there is no commit so skip this execution. Once again faulty logic for the same reasons as above.
				continue
			}

			modifiedQuery := withCountUnlimited(query)
			modifiedQuery = fmt.Sprintf("%s repo:^%s$@%s", modifiedQuery, regexp.QuoteMeta(repository), commits[0].ID)

			log15.Debug("executing query", "query", modifiedQuery)
			results, err := ComputeSearch(ctx, modifiedQuery)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to execute capture group search for repository:%s commit:%s", repository, execution.Revision)
			}

			grouped := GroupByCaptureMatch(results)
			sort.Slice(grouped, func(i, j int) bool {
				return grouped[i].Value < grouped[j].Value
			})
			log15.Debug("grouped results", "grouped", grouped)

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

	var calculated []GeneratedTimeSeries
	seriesCount := 1
	for value, timeCounts := range pivoted {
		var timeseries []TimeDataPoint

		for key, val := range timeCounts {
			timeseries = append(timeseries, TimeDataPoint{
				Time:  key,
				Count: val,
			})
		}

		sort.Slice(timeseries, func(i, j int) bool {
			return timeseries[i].Time.Before(timeseries[j].Time)
		})

		calculated = append(calculated, GeneratedTimeSeries{
			Label:    value,
			Points:   timeseries,
			SeriesId: fmt.Sprintf("dynamic-series-%d", seriesCount),
		})
		seriesCount++
	}
	return calculated, nil
}

func withCountUnlimited(s string) string {
	if strings.Contains(s, "count:") {
		return s
	}
	return s + " count:all"
}

func BuildFrames(numPoints int, interval timeseries.TimeInterval, now time.Time) []types.Frame {
	current := now
	times := make([]time.Time, 0, numPoints)
	times = append(times, now)
	times = append(times, now) // looks weird but is so we can get a frame that is the current point

	for i := 0 - numPoints + 1; i < 0; i++ {
		current = interval.StepBackwards(current)
		times = append(times, current)
	}

	frames := make([]types.Frame, 0, len(times)-1)
	for i := 1; i < len(times); i++ {
		prev := times[i-1]
		frames = append(frames, types.Frame{
			From: times[i],
			To:   prev,
		})
	}

	sort.Slice(frames, func(i, j int) bool {
		return frames[i].From.Before(frames[j].From)
	})
	return frames
}

type GeneratedTimeSeries struct {
	Label    string
	Points   []TimeDataPoint
	SeriesId string
}
