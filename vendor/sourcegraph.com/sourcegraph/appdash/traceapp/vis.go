package traceapp

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"sourcegraph.com/sourcegraph/appdash"

	// Unmarshaling of events depends on the fact that they are registered with
	// Appdash.
	_ "sourcegraph.com/sourcegraph/appdash/httptrace"
	_ "sourcegraph.com/sourcegraph/appdash/sqltrace"
)

// errTimelineItemValidation is returned by timelineItem.Valid when either
// timelineItem.Label or timelime.ItemFullLabel are empty.
var errTimelineItemValidation = errors.New("timeline item validation error")

type timelineItem struct {
	Label        string                  `json:"label"`
	FullLabel    string                  `json:"fullLabel"`
	Times        []*timelineItemTimespan `json:"times"`
	Data         map[string]string       `json:"rawData"`
	SpanID       string                  `json:"spanID"`
	ParentSpanID string                  `json:"parentSpanID"`
	URL          string                  `json:"url"`
	Visible      bool                    `json:"visible"`
}

func (tl *timelineItem) Valid() bool {
	// d3 timeline chart depends on
	// item.Label & item.FullLabel
	if tl.Label == "" || tl.FullLabel == "" {
		return false
	}
	return true
}

type timelineItemTimespan struct {
	Label    string `json:"label"`
	Start    int64  `json:"starting_time"` // msec since epoch
	End      int64  `json:"ending_time"`   // msec since epoch
	Duration int64  `json:"duration"`
}

func (a *App) d3timeline(t *appdash.Trace) ([]timelineItem, error) {
	return a.d3timelineInner(t, 0)
}

func (a *App) d3timelineInner(t *appdash.Trace, depth int) ([]timelineItem, error) {
	var items []timelineItem

	var events []appdash.Event
	if err := appdash.UnmarshalEvents(t.Span.Annotations, &events); err != nil {
		return nil, err
	}

	var u *url.URL
	if t.ID.Parent == 0 {
		var err error
		u, err = a.URLToTrace(t.ID.Trace)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		u, err = a.URLToTraceSpan(t.ID.Trace, t.ID.Span)
		if err != nil {
			return nil, err
		}
	}

	item := timelineItem{
		Label:     t.Span.Name(),
		FullLabel: t.Span.Name(),
		Data:      t.Annotations.StringMap(),
		SpanID:    t.Span.ID.Span.String(),
		URL:       u.String(),
	}

	if !item.Valid() {
		return nil, errTimelineItemValidation
	}

	if t.Span.ID.Parent != 0 {
		item.ParentSpanID = t.Span.ID.Parent.String()
	}
	if depth <= 1 {
		item.Visible = true
	}
	for _, e := range events {
		if e, ok := e.(appdash.TimespanEvent); ok {
			// Continue to next iteration
			// if e.Start() or e.End() are empty time values.
			if e.Start() == (time.Time{}) || e.End() == (time.Time{}) {
				if a.Log != nil {
					a.Log.Printf("Found a TimespanEvent: %+v with invalid/zero times.", e)
				}
				// Continuing so frontend does not break due to current event missing start/end time values.
				continue
			}
			start := e.Start().UnixNano() / int64(time.Millisecond)
			end := e.End().UnixNano() / int64(time.Millisecond)
			ts := timelineItemTimespan{
				Start: start,
				End:   end,
			}
			if t.Span.ID.Parent == 0 {
				ts.Label = e.Schema()
				item.Times = append(item.Times, &ts)
			} else {
				if item.Times == nil {
					item.Times = append(item.Times, &ts)
				} else {
					if item.Times[0].Start > start {
						item.Times[0].Start = start
					}
					if item.Times[0].End < end {
						item.Times[0].End = end
					}
				}
			}
		}
	}
	for _, ts := range item.Times {
		msec := time.Duration(item.Times[0].End-item.Times[0].Start) * time.Millisecond
		if msec > 0 {
			ts.Label = fmt.Sprintf("%s (%s)", item.Label, msec)
			ts.Duration = int64(msec)
		}
	}
	if len(item.Times) == 0 {
		// Items with a null times array will crash d3-timeline.js as it tries
		// to iterate over it. This means the trace doesn't have a single
		// TimespanEvent and is thus invalid.
		return nil, nil
	}
	items = append(items, item)

	for _, child := range t.Sub {
		subItems, err := a.d3timelineInner(child, depth+1)
		if err != nil {
			return nil, err
		}
		items = append(items, subItems...)
	}

	return items, nil
}
