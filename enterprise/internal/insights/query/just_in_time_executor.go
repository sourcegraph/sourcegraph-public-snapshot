package query

import (
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type GeneratedTimeSeries struct {
	Label    string
	Points   []TimeDataPoint
	SeriesId string
}

type timeCounts map[time.Time]int
type justInTimeExecutor struct {
	db        database.DB
	repoStore database.RepoStore
	filter    compression.DataFrameFilter
	clock     func() time.Time
}

func generateTimes(plan compression.BackfillPlan) map[time.Time]int {
	times := make(map[time.Time]int)
	for _, execution := range plan.Executions {
		times[execution.RecordingTime] = 0
		for _, recording := range execution.SharedRecordings {
			times[recording] = 0
		}
	}
	return times
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
