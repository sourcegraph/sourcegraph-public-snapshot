package window

import (
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Configuration represents the rollout windows configured on the site.
type Configuration struct {
	windows []Window
}

// NewConfiguration constructs a Configuration based on the current site
// configuration, including watching the configuration for updates.
func NewConfiguration() *Configuration {
	cfg := &Configuration{}

	conf.Watch(func() {
		if err := cfg.updateFromConfig(conf.Get().BatchChangesRolloutWindows); err != nil {
			log15.Warn("ignoring erroneous batchChanges.rolloutWindows configuration", "err", err)
		}
	})

	return cfg
}

// ValidateConfiguration validates the given site configuration.
func ValidateConfiguration(raw *[]*schema.BatchChangeRolloutWindow) error {
	return (&Configuration{
		windows: []Window{},
	}).updateFromConfig(raw)
}

func (cfg *Configuration) updateFromConfig(raw *[]*schema.BatchChangeRolloutWindow) error {
	// Clear the windows.
	cfg.windows = []Window{}

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
			cfg.windows = append(cfg.windows, window)
		}
	}

	return errs.ErrorOrNil()
}
