package codeintel

import (
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	AutoIndexEnqueuerConfig *enqueuer.Config
	HunkCacheSize           int
}

func (c *Config) Load() {
	enqueuerConfig := &enqueuer.Config{}
	enqueuerConfig.Load()
	c.AutoIndexEnqueuerConfig = enqueuerConfig

	c.HunkCacheSize = c.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "The capacity of the git diff hunk cache.")
}

func (c *Config) Validate() error {
	var errs *multierror.Error
	errs = multierror.Append(errs, c.BaseConfig.Validate())
	errs = multierror.Append(errs, c.AutoIndexEnqueuerConfig.Validate())
	return errs.ErrorOrNil()
}
