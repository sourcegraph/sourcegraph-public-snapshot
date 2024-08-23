package gcplogurl

import (
	"bytes"
	"time"

	"github.com/rickb777/date/period"
)

// TimeRange means when to display the logs.
type TimeRange interface {
	isTimeRange()
	marshalURL(vs values)
}

var _ TimeRange = (*RecentRange)(nil)
var _ TimeRange = (*SpecificTimeBetween)(nil)
var _ TimeRange = (*SpecificTimeWithRange)(nil)

// RecentRange provides "** seconds/minutes/hours/days ago".
type RecentRange struct {
	Last time.Duration
}

func (t *RecentRange) isTimeRange() {}

func (t *RecentRange) marshalURL(vs values) {
	p, _ := period.NewOf(t.Last)
	vs.Set("timeRange", p.String())
}

// SpecificTimeBetween pvovides custom range.
type SpecificTimeBetween struct {
	From time.Time
	To   time.Time
}

func (t *SpecificTimeBetween) isTimeRange() {}

func (t *SpecificTimeBetween) marshalURL(vs values) {
	var buf bytes.Buffer
	if v := t.From; !v.IsZero() {
		buf.WriteString(v.In(time.UTC).Format(time.RFC3339Nano))
	}
	buf.WriteString("/")
	if v := t.To; !v.IsZero() {
		buf.WriteString(v.In(time.UTC).Format(time.RFC3339Nano))
	}
	vs.Set("timeRange", buf.String())
}

// SpecificTimeWithRange provides jump tp time.
type SpecificTimeWithRange struct {
	At    time.Time
	Range time.Duration
}

func (t *SpecificTimeWithRange) isTimeRange() {}

func (t *SpecificTimeWithRange) marshalURL(vs values) {
	var buf bytes.Buffer
	buf.WriteString(t.At.In(time.UTC).Format(time.RFC3339Nano))
	buf.WriteString("/")
	buf.WriteString(t.At.In(time.UTC).Format(time.RFC3339Nano))
	buf.WriteString("--")
	p, _ := period.NewOf(t.Range)
	buf.WriteString(p.String())

	vs.Set("timeRange", buf.String())
}
