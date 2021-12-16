package timeseries

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

type TimeInterval struct {
	Unit  types.IntervalUnit
	Value int
}

func (t TimeInterval) StepBackwards(start time.Time) time.Time {
	return t.step(start, backward)
}

func (t TimeInterval) StepForwards(start time.Time) time.Time {
	return t.step(start, forward)
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
