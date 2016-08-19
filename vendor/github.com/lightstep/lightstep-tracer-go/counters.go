package lightstep

import "github.com/lightstep/lightstep-tracer-go/lightstep_thrift"

const (
	// "Recorded" = recorded at the client end of the instrumentation,
	// prior to any sampling or buffering
	CounterNameRecordedSpans = "recorded_spans"

	// "Malformed" = rejected due to the record being incomplete/inconsistent
	CounterNameMalformedSpans = "malformed_spans"

	// "Dropped" = discarded due to buffer size limitations
	CounterNameDroppedSpans = "dropped_spans"
)

// A set of counter values for a given time window
type counterSet struct {
	droppedSpans int64
}

func (c *counterSet) toThrift() []*lightstep_thrift.NamedCounter {
	// Map the compile-time fields to string-value pairs to keep the
	// communication protocol generic.
	table := []struct {
		name  string
		value int64
	}{
		{CounterNameDroppedSpans, c.droppedSpans},
	}
	counters := make([]*lightstep_thrift.NamedCounter, 0, len(table))
	for _, pair := range table {
		if pair.value == 0 {
			continue
		}
		counters = append(counters, &lightstep_thrift.NamedCounter{
			pair.name, int64(pair.value),
		})
	}
	return counters
}
