package window

import (
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Configuration represents the rollout windows configured on the site.
type Configuration struct {
	windows []Window
	mu      sync.RWMutex
}

// NewConfiguration constructs a Configuration based on the current site
// configuration, including watching the configuration for updates.
func NewConfiguration() *Configuration {
	cfg := &Configuration{}

	conf.Watch(func() {
		if err := cfg.update(conf.Get().BatchChangesRolloutWindows); err != nil {
			log15.Warn("ignoring erroneous batchChanges.rolloutWindows configuration", "err", err)
		}
	})

	return cfg
}

// ValidateConfiguration validates the given site configuration.
func ValidateConfiguration(raw *[]*schema.BatchChangeRolloutWindow) error {
	return (&Configuration{
		windows: []Window{},
	}).update(raw)
}

// TODO: does this need to be public?
func (cfg *Configuration) Current(now time.Time) *Window {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	var window *Window
	for i := range cfg.windows {
		if cfg.windows[i].IsOpen(now) {
			window = &cfg.windows[i]
		}
	}

	return window
}

// currentFor returns the current rollout window, if any, and the duration for
// which that window applies from now.
//
// As this is an internal function, no locking is applied here.
func (cfg *Configuration) currentFor(now time.Time) (*Window, *time.Duration) {
	// Find the last matching window that is currently active.
	var index int
	for i := range cfg.windows {
		if cfg.windows[i].IsOpen(now) {
			index = i
		}
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

// Estimate attempts to estimate when the given entry in a queue of changesets
// to be reconciled would be reconciled. nil indicates that there is no
// reasonable estimate, either because all windows are zero or the estimate is
// too far in the future to be reliable.
func (cfg *Configuration) Estimate(n int) *time.Time {
	// Roughly speaking, we iterate over schedules until we reach the one that
	// would include the given entry.
	rem := n
	at := time.Now()
	until := at.Add(7 * 24 * time.Hour)
	for at.Before(until) {
		schedule := cfg.scheduleAt(at, false)

		// An unlimited schedule means that the reconciliation will happen
		// immediately.
		if schedule.total() == -1 {
			return &at
		}

		rem -= schedule.total()
		if rem < 0 {
			// Try to figure out approximately where in the schedule this will
			// fall.
			perc := float64(schedule.total()) - (float64(+rem) / float64(schedule.total()))
			duration := time.Duration(float64(schedule.ValidUntil().Sub(at)) * perc)
			at.Add(duration)
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

// Schedule calculates a schedule for the near future (most likely about a
// minute), based on the history.
func (cfg *Configuration) Schedule() Schedule {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	// If there are no rollout windows, then we return an unlimited schedule and
	// have the scheduler check back in periodically in case the configuration
	// updated. Ten minutes is probably safe enough.
	if !cfg.HasRolloutWindows() {
		return newSchedule(time.Now(), 10*time.Minute, rate{n: -1})
	}

	return cfg.scheduleAt(time.Now(), true)
}

func (cfg *Configuration) scheduleAt(at time.Time, minimal bool) Schedule {
	window, validity := cfg.currentFor(at)

	// No window means a zero schedule should be returned until the next window
	// change.
	if window == nil {
		if validity != nil {
			if *validity >= 1*time.Minute && minimal {
				return newSchedule(at, 1*time.Minute, rate{n: 0})
			}
			return newSchedule(at, *validity, rate{n: 0})
		}
		// We should always have a validity in this case, but let's be defensive
		// if we don't for some reason. The scheduler can check back in a
		// minute.
		return newSchedule(at, 1*time.Minute, rate{n: 0})
	}

	// OK, so we have a rollout window. It may or may not have an expiry. Either
	// way, let's calculate how long we'd want to schedule in an ideal world.
	if validity == nil {
		if window.rate.unit == ratePerHour {
			return newSchedule(at, 1*time.Hour, window.rate)
		}
		// We never really want to have less than a minute's worth of schedule
		// in this case: we want to check occasionally for updated
		// configurations, but don't want to calculate every time the scheduler
		// needs a changeset.
		return newSchedule(at, 1*time.Minute, window.rate)
	}

	// TODO: minimise if needed.
	return newSchedule(at, *validity, window.rate)
}

func (cfg *Configuration) update(raw *[]*schema.BatchChangeRolloutWindow) error {
	// Ensure we always start with an empty window slice.
	windows := []Window{}

	// Update atomically once we're done generating the new slice.
	defer func() {
		cfg.mu.Lock()
		defer cfg.mu.Unlock()

		cfg.windows = windows
	}()

	// If there's no window configuration, there are no windows, and we can just
	// return here.
	if raw == nil {
		return nil
	}

	var errs *multierror.Error
	for i, rawWindow := range *raw {
		if window, err := parseWindow(rawWindow); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "window %d", i))
		} else {
			windows = append(windows, window)
		}
	}

	return errs.ErrorOrNil()
}
