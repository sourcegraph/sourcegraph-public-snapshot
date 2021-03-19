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
	rwc := &Configuration{
		windows: []Window{},
	}

	conf.Watch(func() {
		if err := rwc.updateFromConfig(conf.Get().BatchChangesRolloutWindows); err != nil {
			log15.Warn("ignoring erroneous batchChanges.rolloutWindows configuration", "err", err)
		}
	})

	return rwc
}

// ValidateConfiguration validates the given site configuration.
func ValidateConfiguration(raw *[]*schema.BatchChangeRolloutWindow) error {
	return (&Configuration{
		windows: []Window{},
	}).updateFromConfig(raw)
}

func (rwc *Configuration) updateFromConfig(raw *[]*schema.BatchChangeRolloutWindow) error {
	if raw == nil {
		return nil
	}

	var errs *multierror.Error
	for i, rawWindow := range *raw {
		if window, err := parseWindow(rawWindow); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "window %d", i))
		} else {
			rwc.windows = append(rwc.windows, window)
		}
	}

	return errs.ErrorOrNil()
}
