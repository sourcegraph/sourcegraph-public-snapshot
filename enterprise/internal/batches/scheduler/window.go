package scheduler

import (
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

type windowRate struct {
	n    int
	unit windowRateUnit
}

type windowRateUnit int

const (
	windowRatePerSecond = iota
	windowRatePerMinute
	windowRatePerHour
)

func parseWindowRateUnit(raw string) (windowRateUnit, error) {
	if raw == "" {
		return windowRatePerSecond, errors.Errorf("malformed unit: %q", raw)
	}

	switch raw[0] {
	case 's':
		return windowRatePerSecond, nil
	case 'm':
		return windowRatePerMinute, nil
	case 'h':
		return windowRatePerHour, nil
	default:
		return windowRatePerSecond, errors.Errorf("malformed unit: %q", raw)
	}
}

func parseWindowRate(raw interface{}) (windowRate, error) {
	switch v := raw.(type) {
	case int:
		if v == 0 {
			return windowRate{n: 0}, nil
		}
		return windowRate{}, errors.Errorf("malformed rate (numeric values can only be 0): %d", v)

	case string:
		s := strings.ToLower(v)
		if s == "unlimited" {
			return windowRate{n: -1}, nil
		}

		wr := windowRate{}
		parts := strings.SplitN(s, "/", 2)
		if len(parts) != 2 {
			return windowRate{}, errors.Errorf("malformed rate: %q", raw)
		}

		var err error
		wr.n, err = strconv.Atoi(parts[0])
		if err != nil {
			return wr, errors.Errorf("malformed rate: %q", raw)
		}

		wr.unit, err = parseWindowRateUnit(parts[1])
		if err != nil {
			return wr, errors.Errorf("malformed rate: %q", raw)
		}

		return wr, nil

	default:
		return windowRate{}, errors.Errorf("malformed rate: unknown type %T", raw)
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

func parseTimePart(s string) (int8, error) {
	part, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return int8(part), nil
}

type RolloutWindow struct {
	days  []time.Weekday
	start *windowTime
	end   *windowTime
	rate  windowRate
}

func parseRolloutWindow(raw *schema.BatchChangeRolloutWindow) (RolloutWindow, error) {
	w := RolloutWindow{}
	var errs *multierror.Error

	if raw == nil {
		return w, errors.New("raw window cannot be nil")
	}

	w.days = make([]time.Weekday, len(raw.Days))
	for i := range raw.Days {
		if day, err := parseWeekday(raw.Days[i]); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			w.days[i] = day
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

	w.rate, err = parseWindowRate(raw.Rate)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	return w, errs.ErrorOrNil()
}

func parseWeekday(raw string) (time.Weekday, error) {
	// We're not going to replicate the full schema validation regex here; we'll
	// assume that the conf package did that satisfactorily and just parse what
	// we need to.
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

type RolloutWindowConfiguration struct {
	windows []RolloutWindow
}

func NewRolloutWindowConfiguration() *RolloutWindowConfiguration {
	rwc := &RolloutWindowConfiguration{
		windows: []RolloutWindow{},
	}

	conf.Watch(func() {
		if err := rwc.updateFromConfig(conf.Get().BatchChangesRolloutWindows); err != nil {
			log15.Warn("ignoring erroneous batchChanges.rolloutWindows configuration", "err", err)
		}
	})

	return rwc
}

func ValidateRolloutWindowConfiguration(raw *[]*schema.BatchChangeRolloutWindow) error {
	return (&RolloutWindowConfiguration{
		windows: []RolloutWindow{},
	}).updateFromConfig(raw)
}

func (rwc *RolloutWindowConfiguration) updateFromConfig(raw *[]*schema.BatchChangeRolloutWindow) error {
	if raw == nil {
		return nil
	}

	var errs *multierror.Error
	for i, rawWindow := range *raw {
		if window, err := parseRolloutWindow(rawWindow); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "window %d", i))
		} else {
			rwc.windows = append(rwc.windows, window)
		}
	}

	return errs.ErrorOrNil()
}
