package timeseries

import (
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/insights/types"
)

func BuildSampleTimes(numPoints int, interval TimeInterval, now time.Time) []time.Time {
	current := now
	times := make([]time.Time, 0, numPoints)
	times = append(times, now)

	for i := 0 - numPoints + 1; i < 0; i++ {
		current = interval.StepBackwards(current)
		times = append(times, current)
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})

	return times
}

func MakeRecordingsFromTimes(times []time.Time, snapshot bool) []types.RecordingTime {
	recordings := make([]types.RecordingTime, 0, len(times))
	for _, t := range times {
		recordings = append(recordings, types.RecordingTime{Snapshot: snapshot, Timestamp: t})
	}
	return recordings
}
