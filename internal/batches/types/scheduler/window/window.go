package window

import (
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Window represents a single rollout window configured on a site.
type Window struct {
	days  weekdaySet
	start *timeOfDay
	end   *timeOfDay
	rate  rate
}

func (w *Window) covers(when timeOfDay) bool {
	if w.start == nil || w.end == nil {
		return true
	}

	return !(when.before(*w.start) || when.after(*w.end) || when.equal(*w.end))
}

// IsOpen checks if this window is currently open.
func (w *Window) IsOpen(at time.Time) bool {
	return w.days.includes(at.Weekday()) && w.covers(timeOfDayFromTime(at))
}

// NextOpenAfter returns the time that this window will next be open.
func (w *Window) NextOpenAfter(after time.Time) time.Time {
	// If the window is currently open, then the next time it will be open is...
	// well, now.
	if w.IsOpen(after) {
		return after
	}

	// From here, the simplest way to find the next active time is to take the
	// start time for this window (which is 00:00 if w.start is nil), then walk
	// forward until we find a weekday where this window is open.
	var t timeOfDay
	if w.start != nil {
		t = *w.start
	}

	when := time.Date(after.Year(), after.Month(), after.Day(), int(t.hour), int(t.minute), 0, 0, time.UTC)
	for {
		if w.days.includes(when.Weekday()) && when.After(after) {
			return when
		} else if when.Sub(after) > 7*24*time.Hour {
			// This should never happen!
			panic("cannot find the next time this window is active after searching the next week")
		}
		when = when.Add(24 * time.Hour)
	}
}

func parseWindowTime(raw string) (*timeOfDay, error) {
	// An empty time is valid.
	if raw == "" {
		return nil, nil
	}

	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return nil, errors.Errorf("malformed time: %q", raw)
	}

	hour, err := parseTimePart(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return nil, errors.Errorf("malformed time: %q", raw)
	}

	minute, err := parseTimePart(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return nil, errors.Errorf("malformed time: %q", raw)
	}

	wt := timeOfDayFromParts(hour, minute)
	return &wt, nil
}

func parseTimePart(s string) (int8, error) {
	part, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return int8(part), nil
}

func parseWeekday(raw string) (time.Weekday, error) {
	// We're not going to replicate the full schema validation regex here; we'll
	// assume that the conf package did that satisfactorily and just parse what
	// we need to, ensuring we can't panic.
	if len(raw) < 3 {
		return time.Sunday, errors.Errorf("unknown weekday: %q", raw)
	}

	switch strings.ToLower(raw[0:3]) {
	case "sun":
		return time.Sunday, nil
	case "mon":
		return time.Monday, nil
	case "tue":
		return time.Tuesday, nil
	case "wed":
		return time.Wednesday, nil
	case "thu":
		return time.Thursday, nil
	case "fri":
		return time.Friday, nil
	case "sat":
		return time.Saturday, nil
	default:
		return time.Sunday, errors.Errorf("unknown weekday: %q", raw)
	}
}

func parseWindow(raw *schema.BatchChangeRolloutWindow) (Window, error) {
	w := Window{}
	var errs error

	if raw == nil {
		return w, errors.New("raw window cannot be nil")
	}

	w.days = newWeekdaySet()
	for i := range raw.Days {
		if day, err := parseWeekday(raw.Days[i]); err != nil {
			errs = errors.Append(errs, err)
		} else {
			w.days.add(day)
		}
	}

	var err error
	w.start, err = parseWindowTime(raw.Start)
	if err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "start time"))
	}
	w.end, err = parseWindowTime(raw.End)
	if err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "end time"))
	}
	if (w.start != nil && w.end == nil) || (w.start == nil && w.end != nil) {
		errs = errors.Append(errs, errors.New("both start and end times must be provided"))
	} else if w.start != nil && w.end != nil && !w.start.before(*w.end) {
		errs = errors.Append(errs, errors.New("end time must be after the start time"))
	}

	w.rate, err = parseRate(raw.Rate)
	if err != nil {
		errs = errors.Append(errs, err)
	}

	return w, errs
}
