package timeseries

import (
	"fmt"
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

// BuildFramesBetween builds frames that have a From time starting at `from` up until `to` at the given interval.
func BuildFramesBetween(from time.Time, to time.Time, interval TimeInterval) []types.Frame {
	times := []time.Time{from}

	current := from
	for current.Before(to) {
		current = interval.StepForwards(current)
		times = append(times, current)
	}

	frames := make([]types.Frame, 0, len(times)-1)
	for i := 1; i < len(times); i++ {
		prev := times[i-1]
		fmt.Println(times[i], prev)
		frames = append(frames, types.Frame{
			From: prev,
			To:   times[i],
		})
	}

	return frames
}

func GetStartTimesFromFrames(frames []types.Frame) []time.Time {
	startTimes := make([]time.Time, 0, len(frames))
	for _, frame := range frames {
		startTimes = append(startTimes, frame.From)
	}
	return startTimes
}
