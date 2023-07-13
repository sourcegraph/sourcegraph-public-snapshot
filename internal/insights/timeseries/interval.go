package timeseries

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/insights/types"
)

type TimeInterval struct {
	Unit  types.IntervalUnit
	Value int
}

var DefaultInterval = TimeInterval{
	Unit:  types.Month,
	Value: 1,
}

func (t TimeInterval) StepBackwards(start time.Time) time.Time {
	return t.step(start, backward)
}

func (t TimeInterval) StepForwards(start time.Time) time.Time {
	return t.step(start, forward)
}

func (t TimeInterval) IsValid() bool {
	validType := false
	switch t.Unit {
	case types.Year:
		fallthrough
	case types.Month:
		fallthrough
	case types.Week:
		fallthrough
	case types.Day:
		fallthrough
	case types.Hour:
		validType = true
	}
	return validType && t.Value >= 0
}

type stepDirection int

const forward stepDirection = 1
const backward stepDirection = -1

func (t TimeInterval) step(start time.Time, direction stepDirection) time.Time {
	switch t.Unit {
	case types.Year:
		return start.AddDate(int(direction)*t.Value, 0, 0)
	case types.Month:
		return start.AddDate(0, int(direction)*t.Value, 0)
	case types.Week:
		return start.AddDate(0, 0, int(direction)*7*t.Value)
	case types.Day:
		return start.AddDate(0, 0, int(direction)*t.Value)
	case types.Hour:
		return start.Add(time.Hour * time.Duration(t.Value) * time.Duration(direction))
	default:
		// this doesn't really make sense, so return something?
		return start.AddDate(int(direction)*t.Value, 0, 0)
	}
}
