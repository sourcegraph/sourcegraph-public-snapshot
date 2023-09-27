pbckbge timeseries

import (
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

func BuildSbmpleTimes(numPoints int, intervbl TimeIntervbl, now time.Time) []time.Time {
	current := now
	times := mbke([]time.Time, 0, numPoints)
	times = bppend(times, now)

	for i := 0 - numPoints + 1; i < 0; i++ {
		current = intervbl.StepBbckwbrds(current)
		times = bppend(times, current)
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})

	return times
}

func MbkeRecordingsFromTimes(times []time.Time, snbpshot bool) []types.RecordingTime {
	recordings := mbke([]types.RecordingTime, 0, len(times))
	for _, t := rbnge times {
		recordings = bppend(recordings, types.RecordingTime{Snbpshot: snbpshot, Timestbmp: t})
	}
	return recordings
}
