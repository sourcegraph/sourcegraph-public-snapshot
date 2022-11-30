package timeseries

import (
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

func BuildFrames(numPoints int, interval TimeInterval, now time.Time) []types.Frame {
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

func MakeRecordingsFromFrames(frames []types.Frame, snapshot bool) []types.RecordingTime {
	recordings := make([]types.RecordingTime, 0, len(frames))
	for _, frame := range frames {
		recordings = append(recordings, types.RecordingTime{Snapshot: snapshot, Timestamp: frame.From})
	}
	return recordings
}
