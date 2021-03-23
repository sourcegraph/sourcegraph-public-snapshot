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

func (cfg *Configuration) Current() *Window {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	var window *Window
	now := time.Now()
	for i := range cfg.windows {
		if cfg.windows[i].IsActive(now) {
			window = &cfg.windows[i]
		}
	}

	return window
}

// Estimate attempts to estimate when the given entry in a queue of changesets to be reconciled would be
func (cfg *Configuration) Estimate(n int) time.Time {
	// TODO: replace with real logic.
	return time.Now()
}

// HasRolloutWindows returns true if one or more windows have been defined.
func (cfg *Configuration) HasRolloutWindows() bool {
	return len(cfg.windows) != 0
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
