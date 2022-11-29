package insights

import (
	"sort"
	"time"
)

func calculateRecordingTimes(createdAt time.Time, lastRecordedAt time.Time, interval timeInterval, existingPoints []time.Time) []time.Time {
	referenceTimes := buildRecordingTimes(12, interval, createdAt.Truncate(time.Hour*24))
	if !lastRecordedAt.IsZero() {
		// If we've had recordings since we need to step through them.
		referenceTimes = append(referenceTimes, buildRecordingTimesBetween(createdAt.Truncate(time.Hour*24), lastRecordedAt, interval)[1:]...)
	}

	if len(existingPoints) == 0 {
		return referenceTimes
	}

	var calculatedRecordingTimes []time.Time
	// For each reference time, we compare it to existing recording times.
	// If the reference time is before the next existing point (i.e. no time exists there), we add it to the list.
	// Else if the existing point is + half an interval close to the reference time, we add that to the list.
	// We use + half an interval because once a recording is queued there might be a delay in stamping. It could not
	// be stamped in the past.
	currentPointIndex := 0
	for _, referenceTime := range referenceTimes {
		referenceTime := referenceTime
		if currentPointIndex < len(existingPoints) {
			existingTime := existingPoints[currentPointIndex]
			intervalDuration := interval.toDuration() // precise to rough estimate of an interval's length (e.g. 1 year = 365 * 24 hours)
			halfAnInterval := intervalDuration / 2
			if interval.unit == hour {
				halfAnInterval = intervalDuration / 4
			}
			differenceInExpectedTime := existingTime.Sub(referenceTime)
			if differenceInExpectedTime >= 0 && differenceInExpectedTime <= halfAnInterval {
				calculatedRecordingTimes = append(calculatedRecordingTimes, existingTime)
				currentPointIndex++
			} else {
				calculatedRecordingTimes = append(calculatedRecordingTimes, referenceTime)
				if referenceTime.After(existingTime) {
					currentPointIndex++
				}
			}
		} else {
			calculatedRecordingTimes = append(calculatedRecordingTimes, referenceTime)
		}
	}

	return calculatedRecordingTimes
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
