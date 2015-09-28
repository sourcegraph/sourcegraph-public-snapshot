package traceapp

import (
	"fmt"

	"sourcegraph.com/sourcegraph/appdash"
)

// aggItem represents a set of traces with the name (label) and their cumulative
// time.
type aggItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type aggMode int

const (
	traceOnly aggMode = iota
	spanOnly
	traceAndSpan
)

// parseAggMode parses an aggregation mode:
//
//  "trace-only" -> traceOnly
//  "span-only" -> spanOnly
//  "trace-and-span" -> traceAndSpan
//
func parseAggMode(s string) aggMode {
	switch s {
	case "trace-only":
		return traceOnly
	case "span-only":
		return spanOnly
	case "trace-and-span":
		return traceAndSpan
	default:
		return traceOnly
	}
}

// aggregate aggregates and encodes the given traces as JSON to the given writer.
func (a *App) aggregate(traces []*appdash.Trace, mode aggMode) ([]*aggItem, error) {
	aggregated := make(map[string]*aggItem)

	// up updates the aggregated map with the given label and value.
	up := func(label string, value int64) {
		// Grab the aggregation item for the named trace, or create a new one if it
		// does not already exist.
		i, ok := aggregated[label]
		if !ok {
			i = &aggItem{Label: label}
			aggregated[label] = i
		}

		// Perform aggregation.
		i.Value += value
		if i.Value == 0 {
			i.Value = 1 // Must be positive values or else d3pie won't render.
		}
	}

	for _, trace := range traces {
		// Calculate the cumulative time -- which we can already get through the
		// profile view's calculation method.
		profiles, childProf, err := a.calcProfile(nil, trace)
		if err != nil {
			return nil, err
		}

		if mode == traceOnly {
			up(childProf.Name, childProf.TimeCum)
		} else if mode == spanOnly {
			for _, spanProf := range profiles[1:] {
				up(spanProf.Name, spanProf.Time)
			}
		} else if mode == traceAndSpan {
			for _, spanProf := range profiles[1:] {
				up(fmt.Sprintf("%s: %s", childProf.Name, spanProf.Name), spanProf.Time)
			}
		}
	}

	// Form an array (d3pie needs a JSON array, not a map).
	list := make([]*aggItem, 0, len(aggregated))
	for _, item := range aggregated {
		list = append(list, item)
	}
	return list, nil
}
