package recording_times

import (
	"sort"
	"time"
)

func calculateRecordingTimes(createdAt time.Time, lastRecordedAt time.Time, interval timeInterval, existingPoints []time.Time) []time.Time {
	referenceTimes := buildRecordingTimes(12, interval, createdAt.Truncate(time.Minute))
	if !lastRecordedAt.IsZero() {
		// If we've had recordings since we need to step through them.
		referenceTimes = append(referenceTimes, buildRecordingTimesBetween(createdAt.Truncate(time.Minute), lastRecordedAt, interval)[1:]...)
	}

	if len(existingPoints) == 0 {
		return referenceTimes
	}

	// The set of recording times will be augmented with zeros for missing points in the expected leading set and the
	// expected trailing set.
	// i.e. for reference times [R1, R2, R3, R4, R5, R6, R7, R8] and existing times [E4, E6, E7],
	// the calculated times will be [R1, R2, R3, E4, E6, E7, R8]

	var calculatedRecordingTimes []time.Time

	// If the first existing point is newer than the oldest expected point then leading points are added.
	oldestReferencePoint := referenceTimes[0]
	if !withinHalfAnInterval(existingPoints[0], oldestReferencePoint, interval) {
		for i := 0; i < len(referenceTimes) && !withinHalfAnInterval(referenceTimes[i], existingPoints[0], interval); i++ {
			calculatedRecordingTimes = append(calculatedRecordingTimes, referenceTimes[i])
		}
	}
	// Any existing middle points are added.
	calculatedRecordingTimes = append(calculatedRecordingTimes, existingPoints...)

	// If the last existing point is older than the newest expected point then trailing points are added.
	newestReferencePoint := referenceTimes[len(referenceTimes)-1]
	if !withinHalfAnInterval(newestReferencePoint, existingPoints[len(existingPoints)-1], interval) {
		var backwardTrailingPoints []time.Time
		// We have to walk backwards through the trailing reference times as we do not know how many extra reference
		// times need to be added.
		for i := len(referenceTimes) - 1; i >= 0 && !withinHalfAnInterval(referenceTimes[i], existingPoints[len(existingPoints)-1], interval); i-- {
			backwardTrailingPoints = append(backwardTrailingPoints, referenceTimes[i])
		}
		// Once we've walked back through the reference times we append them in the correct order.
		for i := len(backwardTrailingPoints) - 1; i >= 0; i-- {
			calculatedRecordingTimes = append(calculatedRecordingTimes, backwardTrailingPoints[i])
		}
	}

	return calculatedRecordingTimes
}

func withinHalfAnInterval(a, b time.Time, interval timeInterval) bool {
	intervalDuration := interval.toDuration() // precise to rough estimate of an interval's length (e.g. 1 year = 365 * 24 hours)
	halfAnInterval := intervalDuration / 2
	if interval.unit == hour {
		halfAnInterval = intervalDuration / 4
	}
	differenceInExpectedTime := b.Sub(a)
	return differenceInExpectedTime >= 0 && differenceInExpectedTime <= halfAnInterval
}

type intervalUnit string

const (
	month intervalUnit = "MONTH"
	day   intervalUnit = "DAY"
	week  intervalUnit = "WEEK"
	year  intervalUnit = "YEAR"
	hour  intervalUnit = "HOUR"
)

type timeInterval struct {
	unit  intervalUnit
	value int
}

func (t timeInterval) toDuration() time.Duration {
	var singleUnitDuration time.Duration
	switch t.unit {
	case year:
		singleUnitDuration = time.Hour * 24 * 365
	case month:
		singleUnitDuration = time.Hour * 24 * 30
	case week:
		singleUnitDuration = time.Hour * 24 * 7
	case day:
		singleUnitDuration = time.Hour * 24
	case hour:
		singleUnitDuration = time.Hour
	}
	return singleUnitDuration * time.Duration(t.value)
}

func buildRecordingTimes(numPoints int, interval timeInterval, now time.Time) []time.Time {
	current := now
	times := make([]time.Time, 0, numPoints)
	times = append(times, now)

	for i := 0 - numPoints + 1; i < 0; i++ {
		current = interval.stepBackwards(current)
		times = append(times, current)
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	return times
}

// buildRecordingTimesBetween builds times starting at `start` up until `end` at the given interval.
func buildRecordingTimesBetween(start time.Time, end time.Time, interval timeInterval) []time.Time {
	times := []time.Time{}

	current := start
	for current.Before(end) {
		times = append(times, current)
		current = interval.stepForwards(current)
	}

	return times
}

func (t timeInterval) stepBackwards(start time.Time) time.Time {
	return t.step(start, backward)
}

func (t timeInterval) stepForwards(start time.Time) time.Time {
	return t.step(start, forward)
}

type stepDirection int

const (
	forward  stepDirection = 1
	backward stepDirection = -1
)

func (t timeInterval) step(start time.Time, direction stepDirection) time.Time {
	switch t.unit {
	case year:
		return start.AddDate(int(direction)*t.value, 0, 0)
	case month:
		return start.AddDate(0, int(direction)*t.value, 0)
	case week:
		return start.AddDate(0, 0, int(direction)*7*t.value)
	case day:
		return start.AddDate(0, 0, int(direction)*t.value)
	case hour:
		return start.Add(time.Hour * time.Duration(t.value) * time.Duration(direction))
	default:
		// this doesn't really make sense, so return something?
		return start.AddDate(int(direction)*t.value, 0, 0)
	}
}
