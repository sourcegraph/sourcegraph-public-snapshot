package codeintel

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type config struct {
	env.BaseConfig

	LSIFUploadStoreConfig          *lsifuploadstore.Config
	MaximumIndexesPerMonikerSearch int
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.LSIFUploadStoreConfig = &lsifuploadstore.Config{}
	c.LSIFUploadStoreConfig.Load()

	c.MaximumIndexesPerMonikerSearch = c.GetInt("PRECISE_CODE_INTEL_MAXIMUM_INDEXES_PER_MONIKER_SEARCH", "500", "The maximum number of indexes to search at once when doing cross-index code navigation.")
}

func (c *config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.LSIFUploadStoreConfig.Validate())
	return errs
}
