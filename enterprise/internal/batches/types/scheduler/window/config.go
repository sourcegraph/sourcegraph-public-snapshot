package window

import (
	"math"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Configuration represents the rollout windows configured on the site.
type Configuration struct {
	windows []Window
}

// NewConfiguration constructs a Configuration based on the given site
// configuration.
func NewConfiguration(raw *[]*schema.BatchChangeRolloutWindow) (*Configuration, error) {
	windows, err := parseConfiguration(raw)
	if err != nil {
		return nil, err
	}

	return &Configuration{windows: windows}, nil
}

// Estimate attempts to estimate when the given entry in a queue of changesets
// to be reconciled would be reconciled. nil indicates that there is no
// reasonable estimate, either because all windows are zero or the estimate is
// too far in the future to be reliable.
func (cfg *Configuration) Estimate(now time.Time, n int) *time.Time {
	if !cfg.HasRolloutWindows() {
		return &now
	}

	// Roughly speaking, we iterate over schedules until we reach the one that
	// would include the given entry. If we hit a week in the future, we'll
	// bail, because a lot can happen in a week.
	rem := n
	at := now
	until := at.Add(7 * 24 * time.Hour)
	for at.Before(until) {
		schedule := cfg.scheduleAt(at)

		// An unlimited schedule means that the reconciliation will happen
		// immediately at that point the window opens.
		if schedule.total() == -1 {
			return &at
		}

		total := schedule.total()
		if total == 0 {
			at = schedule.ValidUntil()
			continue
		}

		rem -= total
		if rem < 0 {
			// We know how many extra reconciliations will occur within this
			// schedule, so we can use that calculate what percentage of the way
			// into the window our target will be reconciled, then we can
			// multiple the schedule duration by that to get the approximate
			// time.
			perc := 1.0 - math.Abs(float64(rem))/float64(total)
			duration := time.Duration(float64(schedule.ValidUntil().Sub(at)) * perc)
			at = at.Add(duration)
			return &at
		} else if rem == 0 {
			// Special case: this will be the very last entry to be reconciled.
			at = schedule.ValidUntil()
			return &at
		}

		at = schedule.ValidUntil()
	}

	return nil
}

// HasRolloutWindows returns true if one or more windows have been defined.
func (cfg *Configuration) HasRolloutWindows() bool {
	return len(cfg.windows) != 0
}

// Schedule returns the currently active schedule.
func (cfg *Configuration) Schedule() *Schedule {
	// If there are no rollout windows, then we return an unlimited schedule and
	// have the scheduler check back in periodically in case the configuration
	// updated. Ten minutes is probably safe enough.
	if !cfg.HasRolloutWindows() {
		return newSchedule(time.Now(), 10*time.Minute, rate{n: -1})
	}

	return cfg.scheduleAt(time.Now())
}

// windowFor returns the rollout window for the given time, if any, and the
// duration for which that window applies. The duration will be nil if the
// current window applies indefinitely.
func (cfg *Configuration) windowFor(now time.Time) (*Window, *time.Duration) {
	// If there are no rollout windows, there's no current window. This should
	// be checked before entry, but let's at least not panic here.
	if len(cfg.windows) == 0 {
		return nil, nil
	}

	// Find the last matching window that is currently active.
	index := -1
	for i := range cfg.windows {
		if cfg.windows[i].IsOpen(now) {
			index = i
		}
	}
	if index == -1 {
		// No matching window, so let's figure out when the next window would
		// open and return a nil window.
		var next *time.Time
		for i := range cfg.windows {
			at := cfg.windows[i].NextOpenAfter(now)
			if next == nil || at.Before(*next) {
				next = &at
			}
		}

		// If we never saw a time, that's weird, since this scenario shouldn't
		// occur if there are windows defined, but let's just say nothing can
		// happen forever for now.
		if next == nil {
			return nil, nil
		}

		duration := next.Sub(now)
		return nil, &duration
	}
	window := &cfg.windows[index]

	// Calculate when this window closes. This may not be the end time on the
	// window: if there's a later window that starts before the end time of this
	// window, that will end up taking precedence.
	var end *time.Time

	if window.end == nil {
		// There may still be a weekday restriction, so we should figure that out.
		if !window.days.all() {
			start := now.Truncate(24 * time.Hour)
			for {
				start = start.Add(24 * time.Hour)
				if !window.days.includes(start.Weekday()) {
					start.Add(-1 * time.Second)
					end = &start
					break
				} else if start.After(now.Add(7 * 24 * time.Hour)) {
					panic("could not find end of a day-limited window in the next week")
				}
			}
		}
	} else {
		// We have a concrete end time for this window, so we can set end to
		// that.
		windowEnd := time.Date(now.Year(), now.Month(), now.Day(), int(window.end.hour), int(window.end.minute), 0, 0, time.UTC)
		end = &windowEnd
	}

	// Now we iterate over the subsequent windows in the configuration and see
	// if any of them would start before the existing end time, which would make
	// them active (since they're subsequent). Note that we're using a C style
	// for loop here instead of slicing: we'd have to check the bounds of the
	// cfg.windows slice before being able to subslice, and this feels more
	// readable.
	for i := index + 1; i < len(cfg.windows); i++ {
		nextActive := cfg.windows[i].NextOpenAfter(now)
		if end != nil {
			if nextActive.Before(*end) {
				end = &nextActive
			}
		} else {
			end = &nextActive
		}
	}

	// If we still don't have an end time, then this window remains open forever
	// and cannot be overridden. Cool.
	if end == nil {
		return window, nil
	}

	// Otherwise, let's calculate how long we have until this window closes, and
	// return that.
	d := end.Sub(now)
	return window, &d
}

// scheduleAt constructs a schedule that is valid at the given time. Note that
// scheduleAt does _not_ check if there are rollout windows configured at all:
// the caller must do this.
func (cfg *Configuration) scheduleAt(at time.Time) *Schedule {
	// Get the window in effect at this time, along with how long it's valid
	// for.
	window, validity := cfg.windowFor(at)

	// No window means a zero schedule should be returned until the next window
	// change.
	if window == nil {
		if validity != nil {
			return newSchedule(at, *validity, rate{n: 0})
		}
		// We should always have a validity in this case, but let's be defensive
		// if we don't for some reason. The scheduler can check back in a
		// minute.
		return newSchedule(at, 1*time.Minute, rate{n: 0})
	}

	// OK, so we have a rollout window. It may or may not have an expiry. If it
	// doesn't, then let's hand back a day of schedule.
	if validity == nil {
		return newSchedule(at, 24*time.Hour, window.rate)
	}

	// Otherwise, we can provide a schedule that goes right up to the end of the
	// window, at which point the scheduler can check back in and get the new
	// schedule.
	return newSchedule(at, *validity, window.rate)
}

func parseConfiguration(raw *[]*schema.BatchChangeRolloutWindow) ([]Window, error) {
	// Ensure we always start with an empty window slice.
	windows := []Window{}

	// If there's no window configuration, there are no windows, and we can just
	// return here.
	if raw == nil {
		return windows, nil
	}

	var errs error
	for i, rawWindow := range *raw {
		if window, err := parseWindow(rawWindow); err != nil {
			errs = errors.Append(errs, errors.Wrapf(err, "window %d", i))
		} else {
			windows = append(windows, window)
		}
	}

	return windows, errs
}
