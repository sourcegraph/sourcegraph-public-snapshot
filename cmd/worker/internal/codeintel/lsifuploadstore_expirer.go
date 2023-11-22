package codeintel

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type lsifuploadstoreExpirer struct{}

func NewPreciseCodeIntelUploadExpirer() job.Job {
	return &lsifuploadstoreExpirer{}
}

func (j *lsifuploadstoreExpirer) Description() string {
	return ""
}

func (j *lsifuploadstoreExpirer) Config() []env.Config {
	return []env.Config{
		lsifuploadstoreExpirerConfigInst,
	}
}

func (j *lsifuploadstoreExpirer) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	ctx := context.Background()

	uploadStore, err := lsifuploadstore.New(ctx, observationCtx, lsifuploadstoreExpirerConfigInst.LSIFUploadStoreConfig)
	if err != nil {
		observationCtx.Logger.Fatal("Failed to create upload store", log.Error(err))
	}

	return []goroutine.BackgroundRoutine{
		uploadstore.NewExpirer(ctx, uploadStore, lsifuploadstoreExpirerConfigInst.prefix, lsifuploadstoreExpirerConfigInst.maxAge, lsifuploadstoreExpirerConfigInst.interval),
	}, nil
}

type lsifuploadstoreExpirerConfig struct {
	env.BaseConfig

	prefix                string
	maxAge                time.Duration
	interval              time.Duration
	LSIFUploadStoreConfig *lsifuploadstore.Config
}

var lsifuploadstoreExpirerConfigInst = &lsifuploadstoreExpirerConfig{}

func (c *lsifuploadstoreExpirerConfig) Load() {
	c.LSIFUploadStoreConfig = &lsifuploadstore.Config{}
	c.LSIFUploadStoreConfig.Load()

	c.prefix = c.GetOptional("CODEINTEL_UPLOADSTORE_EXPIRER_PREFIX", "The prefix of objects to expire in the precise code intel upload bucket.")
	c.maxAge = c.GetInterval("CODEINTEL_UPLOADSTORE_EXPIRER_MAX_AGE", "168h", "The max age of objects in the precise code intel upload bucket.")
	c.interval = c.GetInterval("CODEINTEL_UPLOADSTORE_EXPIRER_INTERVAL", "1h", "The frequency at which to expire precise code intel upload bucket objects.")
}

func (c *lsifuploadstoreExpirerConfig) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.LSIFUploadStoreConfig.Validate())
	return errs
}
