package codeintel

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	LSIFUploadStoreConfig          *lsifuploadstore.Config
	AutoIndexEnqueuerConfig        *enqueuer.Config
	HunkCacheSize                  int
	MaximumIndexesPerMonikerSearch int
}

func (c *Config) Load() {
	enqueuerConfig := &enqueuer.Config{}
	enqueuerConfig.Load()
	c.AutoIndexEnqueuerConfig = enqueuerConfig

	c.LSIFUploadStoreConfig = &lsifuploadstore.Config{}
	c.LSIFUploadStoreConfig.Load()

	c.HunkCacheSize = c.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "The capacity of the git diff hunk cache.")
	c.MaximumIndexesPerMonikerSearch = c.GetInt("PRECISE_CODE_INTEL_MAXIMUM_INDEXES_PER_MONIKER_SEARCH", "50", "The maximum number of indexes to search at once when doing cross-index code navigation.")
}

func (c *Config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.LSIFUploadStoreConfig.Validate())
	errs = errors.Append(errs, c.AutoIndexEnqueuerConfig.Validate())
	return errs
}
