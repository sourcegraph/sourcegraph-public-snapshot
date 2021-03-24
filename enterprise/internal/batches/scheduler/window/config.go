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
// to be reconciled would be reconciled.
func (cfg *Configuration) Estimate(history History, n int) time.Time {
	// TODO: replace with real logic.
	return time.Now()
}

// HasRolloutWindows returns true if one or more windows have been defined.
func (cfg *Configuration) HasRolloutWindows() bool {
	return len(cfg.windows) != 0
}

// Schedule calculates a schedule for the near future (most likely about a
// minute), based on the history. If nil is returned, then an unlimited number
// of changesets may be reconciled.
func (cfg *Configuration) Schedule(history History) Schedule {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	// Special cases. So many special cases.

	// Firstly, if there are no rollout windows, then we just return nil and the
	// scheduler can do what it wants.
	if !cfg.HasRolloutWindows() {
		return nil
	}

	now := time.Now()
	window, validity := cfg.currentFor(now)

	// Next up, no window means a zero schedule should be returned until the
	// next window change.
	if window == nil {
		if validity != nil {
			return &zeroSchedule{baseSchedule{base: now, duration: *validity}}
		} else {
			// We should always have a validity in this case, but let's be
			// defensive if we don't for some reason. The scheduler can check
			// back in a minute.
			return &zeroSchedule{baseSchedule{base: now, duration: 1 * time.Minute}}
		}
	}

	// OK, so we have a rollout window. It may or may not have an expiry. Either
	// way, let's calculate how long we'd want to schedule in an ideal world.
	// For per-second or per-minute rates, this is super easy: we can return a
	// schedule with an increment of one minute, in which case we can ignore the
	// expiry: even if it's within the next minute, the next check will kick in
	// the new window and its rate without any complication, since windows only
	// have per-minute resolution.
	//
	// TODO: an obvious improvement here is to set the duration to be the full
	// validity, if present.
	if window.rate.unit == ratePerSecond {
		return &linearSchedule{
			baseSchedule: baseSchedule{base: now, duration: 1 * time.Minute},
			n:            window.rate.n * 60,
		}
	} else if window.rate.unit == ratePerMinute {
		return &linearSchedule{
			baseSchedule: baseSchedule{base: now, duration: 1 * time.Minute},
			n:            window.rate.n,
		}
	}

	// The rate here is per-hour, which means that we have to provide a longer
	// schedule increment, which in turn means that we have to factor the expiry
	// in. Let's figure out how long we want to have this schedule run in an
	// ideal world, and then we can work backwards from there.
	//
	// Let's flip it upside down and reverse it: how long should we wait between
	// reconciliations?
	perN := (1 * time.Hour) / time.Duration(window.rate.n)

	// If it's more than a minute, then we're only ever going to provide a
	// schedule that allows for zero or one reconcile.
	if perN > 1*time.Minute {
		// One ping only, Vasily.
		if validity == nil || *validity >= perN {
			return &linearSchedule{
				baseSchedule: baseSchedule{base: now, duration: perN},
				n:            1,
			}
		}
		// The rollout window is going to change. We could do something
		// complicated here like look ahead, or try to round and hope for the
		// best, but instead, let's be conservative and not allow something to
		// be scheduled and have the scheduler come back when the rollout window
		// changes.
		return &zeroSchedule{baseSchedule{base: now, duration: *validity}}
	}

	// If it's less than a minute, then we're getting perilously close to having
	// to use some floating point maths. Let's set up a schedule for as far as
	// we can if there's a defined validity, otherwise we'll set up the next
	// hour.
	if validity == nil || *validity >= 1*time.Hour {
		return &linearSchedule{
			baseSchedule: baseSchedule{base: now, duration: 1 * time.Hour},
			n:            window.rate.n,
		}
	}

	return &linearSchedule{
		baseSchedule: baseSchedule{base: now, duration: *validity},
		n:            int(float64(window.rate.n) * (float64(*validity) / float64(1*time.Hour))),
	}
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
