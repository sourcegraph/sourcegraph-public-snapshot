package backfillv2

import (
	"time"
)

type intervalUnit string

const (
	month intervalUnit = "MONTH"
	day   intervalUnit = "DAY"
	week  intervalUnit = "WEEK"
	year  intervalUnit = "YEAR"
	hour  intervalUnit = "HOUR"
)

type timeInterval struct {
	Unit  intervalUnit
	Value int
}

var defaultInterval = timeInterval{
	Unit:  month,
	Value: 1,
}

func (t timeInterval) StepBackwards(start time.Time) time.Time {
	return t.step(start, backward)
}

func (t timeInterval) StepForwards(start time.Time) time.Time {
	return t.step(start, forward)
}

func (t timeInterval) IsValid() bool {
	validType := false
	switch t.Unit {
	case year:
		fallthrough
	case month:
		fallthrough
	case week:
		fallthrough
	case day:
		fallthrough
	case hour:
		validType = true
	}
	return validType && t.Value >= 0
}

type stepDirection int

const forward stepDirection = 1
const backward stepDirection = -1

func (t timeInterval) step(start time.Time, direction stepDirection) time.Time {
	switch t.Unit {
	case year:
		return start.AddDate(int(direction)*t.Value, 0, 0)
	case month:
		return start.AddDate(0, int(direction)*t.Value, 0)
	case week:
		return start.AddDate(0, 0, int(direction)*7*t.Value)
	case day:
		return start.AddDate(0, 0, int(direction)*t.Value)
	case hour:
		return start.Add(time.Hour * time.Duration(t.Value) * time.Duration(direction))
	default:
		// this doesn't really make sense, so return something?
		return start.AddDate(int(direction)*t.Value, 0, 0)
	}
}

func nextSnapshot(current time.Time) time.Time {
	year, month, day := current.In(time.UTC).Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
}
