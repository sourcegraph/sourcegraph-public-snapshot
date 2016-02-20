package traceapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

type profile struct {
	Name                        string
	URL                         string
	Time, TimeChildren, TimeCum int64
}

// calcProfile calculates a profile for the given trace and appends it to the
// given buffer (buf), which is then returned (prof). If an error is returned,
// all other returned values are nil. The childProf is literally the *profile
// associated with the given trace (t).
func (a *App) calcProfile(buf []*profile, t *appdash.Trace) (prof []*profile, childProf *profile, err error) {
	// Unmarshal the trace's span events.
	var events []appdash.Event
	if err := appdash.UnmarshalEvents(t.Span.Annotations, &events); err != nil {
		return nil, nil, err
	}

	// Get the proper URL to the trace view.
	var u *url.URL
	if t.ID.Parent == 0 {
		u, err = a.URLToTrace(t.ID.Trace)
		if err != nil {
			return nil, nil, err
		}
	} else {
		u, err = a.URLToTraceSpan(t.ID.Trace, t.ID.Span)
		if err != nil {
			return nil, nil, err
		}
	}

	// Initialize the span's profile structure. We use either the span's given
	// name, or it's ID as a string if it has no given name.
	p := &profile{
		Name: t.Span.Name(),
		URL:  u.String(),
	}
	if len(p.Name) == 0 {
		p.Name = t.Span.ID.Span.String()
	}
	buf = append(buf, p)

	// Store the time for the largest timespan event the span has.
	for _, ev := range events {
		ts, ok := ev.(appdash.TimespanEvent)
		if !ok {
			continue
		}
		// To match the timeline properly we use floats and round up.
		msf := float64(ts.End().Sub(ts.Start())) / float64(time.Millisecond)
		ms := int64(msf + 0.5)
		if ms > p.Time {
			p.Time = ms
		}
	}

	// TimeChildren is our time + the children's time.
	p.TimeChildren = p.Time

	// The cumulative time is our time + all children's time.
	p.TimeCum = p.Time

	// Descend recursively into each sub-trace and calculate the profile for
	// each child span.
	for _, child := range t.Sub {
		buf, childProf, err = a.calcProfile(buf, child)
		if err != nil {
			return nil, nil, err
		}

		// Aggregate our direct children's time.
		p.TimeChildren += childProf.Time

		// As our child's profile has the cumulative time (which is initially,
		// it's self time) -- we can simply aggregate it here and we have our
		// trace's cumulative time (i.e. it is effectively recursive).
		p.TimeCum += childProf.TimeCum
	}
	return buf, p, nil
}

// profile generates and encodes the given trace as JSON to the given writer.
func (a *App) profile(t *appdash.Trace, out io.Writer) error {
	// Generate the profile.
	prof, _, err := a.calcProfile(nil, t)
	if err != nil {
		return err
	}

	// Encode to JSON.
	j, err := json.Marshal(prof)
	if err != nil {
		return err
	}

	// Write out.
	_, err = io.Copy(out, bytes.NewReader(j))
	return err
}
