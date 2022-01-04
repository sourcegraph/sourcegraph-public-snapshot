package codeintel

import (
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	UploadStoreConfig       *uploadstore.Config
	AutoIndexEnqueuerConfig *enqueuer.Config
	HunkCacheSize           int
}

var config = &Config{}

func init() {
	uploadStoreConfig := &uploadstore.Config{}
	uploadStoreConfig.Load()
	config.UploadStoreConfig = uploadStoreConfig

	enqueuerConfig := &enqueuer.Config{}
	enqueuerConfig.Load()
	config.AutoIndexEnqueuerConfig = enqueuerConfig

	config.HunkCacheSize = config.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "The capacity of the git diff hunk cache.")
}

func (c *Config) Validate() error {
	var errs *multierror.Error
	errs = multierror.Append(errs, c.BaseConfig.Validate())
	errs = multierror.Append(errs, c.UploadStoreConfig.Validate())
	errs = multierror.Append(errs, c.AutoIndexEnqueuerConfig.Validate())
	return errs.ErrorOrNil()
}
