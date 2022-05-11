package okay

import (
	"fmt"
	"time"
)

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

func (m *Metric) ValueString() string {
	if m == nil {
		return "<nil>"
	}
	switch m.Type {
	case "count", "number":
		return fmt.Sprintf("%+v", m.Value)
	case "durationMs":
		dur := time.Duration(m.Value) * time.Millisecond
		return dur.String()
	}
	return fmt.Sprintf("invalid type %q", m.Type)
}
