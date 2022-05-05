package okay

import "time"

// Metric represents a particular metric attached to an event.
type Metric struct {
	// Type is either "count", "durationMs" or "number".
	Type string `json:"type"`
	// Value is the actual value reported for this metric in a given event.
	Value float64 `json:"value"`
}

func Count(value int) Metric {
	return Metric{Type: "count", Value: float64(value)}
}

func Duration(duration time.Duration) Metric {
	return Metric{Type: "durationMs", Value: float64(duration.Milliseconds())}
}

func Number(number int) Metric {
	return Metric{Type: "number", Value: float64(number)}
}
