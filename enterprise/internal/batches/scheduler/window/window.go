package window

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/schema"
)

// Window represents a single rollout window configured on a site.
type Window struct {
	days  *weekdaySet
	start *windowTime
	end   *windowTime
	rate  rate
}

func (w *Window) isTimeWithin(when timeOfDay) bool {
	if w.start == nil || w.end == nil {
		return true
	}

	return !(when.before(w.start.timeOfDay()) || when.after(w.end.timeOfDay()))
}

// IsOpen checks if this window is currently open.
func (w *Window) IsOpen(at time.Time) bool {
	return w.days.includes(at.Weekday()) && w.isTimeWithin(timeOfDayFromTime(at))
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
		t = w.start.timeOfDay()
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

type rate struct {
	n    int
	unit rateUnit
}

func (r rate) IsUnlimited() bool {
	return r.n == -1
}

type rateUnit int

const (
	ratePerSecond = iota
	ratePerMinute
	ratePerHour
)

func (ru rateUnit) AsDuration() time.Duration {
	switch ru {
	case ratePerSecond:
		return time.Second
	case ratePerMinute:
		return time.Minute
	case ratePerHour:
		return time.Hour
	default:
		panic(fmt.Sprintf("invalid rateUnit value: %v", ru))
	}
}

func parseRateUnit(raw string) (rateUnit, error) {
	// We're not going to replicate the full schema validation regex here; we'll
	// assume that the conf package did that satisfactorily and just parse what
	// we need to, ensuring we can't panic.
	if raw == "" {
		return ratePerSecond, errors.Errorf("malformed unit: %q", raw)
	}

	switch raw[0] {
	case 's', 'S':
		return ratePerSecond, nil
	case 'm', 'M':
		return ratePerMinute, nil
	case 'h', 'H':
		return ratePerHour, nil
	default:
		return ratePerSecond, errors.Errorf("malformed unit: %q", raw)
	}
}

func parseRate(raw interface{}) (rate, error) {
	switch v := raw.(type) {
	case int:
		if v == 0 {
			return rate{n: 0}, nil
		}
		return rate{}, errors.Errorf("malformed rate (numeric values can only be 0): %d", v)

	case string:
		s := strings.ToLower(v)
		if s == "unlimited" {
			return rate{n: -1}, nil
		}

		wr := rate{}
		parts := strings.SplitN(s, "/", 2)
		if len(parts) != 2 {
			return rate{}, errors.Errorf("malformed rate: %q", raw)
		}

		var err error
		wr.n, err = strconv.Atoi(parts[0])
		if err != nil || wr.n < 0 {
			return wr, errors.Errorf("malformed rate: %q", raw)
		}

		wr.unit, err = parseRateUnit(parts[1])
		if err != nil {
			return wr, errors.Errorf("malformed rate: %q", raw)
		}

		return wr, nil

	default:
		return rate{}, errors.Errorf("malformed rate: unknown type %T", raw)
	}
}

type windowTime struct {
	hour   int8
	minute int8
}

func newWindowTime(raw string) (*windowTime, error) {
	// An empty time is valid.
	if raw == "" {
		return nil, nil
	}

	wt := &windowTime{}
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return nil, errors.Errorf("malformed time: %q", raw)
	}

	var err error
	wt.hour, err = parseTimePart(parts[0])
	if err != nil || wt.hour < 0 || wt.hour > 23 {
		return nil, errors.Errorf("malformed time: %q", raw)
	}

	wt.minute, err = parseTimePart(parts[1])
	if err != nil || wt.minute < 0 || wt.minute > 59 {
		return nil, errors.Errorf("malformed time: %q", raw)
	}

	return wt, nil
}

func (wt *windowTime) timeOfDay() timeOfDay {
	return timeOfDayFromParts(wt.hour, wt.minute)
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
	var errs *multierror.Error

	if raw == nil {
		return w, errors.New("raw window cannot be nil")
	}

	w.days = newWeekdaySet()
	for i := range raw.Days {
		if day, err := parseWeekday(raw.Days[i]); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			w.days.add(day)
		}
	}

	var err error
	w.start, err = newWindowTime(raw.Start)
	if err != nil {
		errs = multierror.Append(errs, errors.Wrap(err, "start time"))
	}
	w.end, err = newWindowTime(raw.End)
	if err != nil {
		errs = multierror.Append(errs, errors.Wrap(err, "end time"))
	}
	if (w.start != nil && w.end == nil) || (w.start == nil && w.end != nil) {
		errs = multierror.Append(errs, errors.New("both start and end times must be provided"))
	} else if w.start != nil && w.end != nil && !w.start.timeOfDay().before(w.end.timeOfDay()) {
		errs = multierror.Append(errs, errors.New("end time must be after the start time"))
	}

	w.rate, err = parseRate(raw.Rate)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	return w, errs.ErrorOrNil()
}
