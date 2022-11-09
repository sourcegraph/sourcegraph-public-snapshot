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
	// Else if the existing point is +/- half an interval close to the reference time, we add that to the list.
	currentPointIndex := 0
	for _, referenceTime := range referenceTimes {
		referenceTime := referenceTime
		if currentPointIndex < len(existingPoints) {
			existingTime := existingPoints[currentPointIndex]
			halfAnInterval := interval.toDuration() / 2
			if existingTime.Sub(referenceTime).Abs() <= halfAnInterval {
				calculatedRecordingTimes = append(calculatedRecordingTimes, existingTime)
				currentPointIndex++
			} else {
				calculatedRecordingTimes = append(calculatedRecordingTimes, referenceTime)
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
	switch t.unit {
	case year:
		return time.Hour * 24 * 365
	case month:
		return time.Hour * 24 * 30
	case week:
		return time.Hour * 24 * 7
	case day:
		return time.Hour * 24
	case hour:
		return time.Hour
	}
	return time.Hour
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
