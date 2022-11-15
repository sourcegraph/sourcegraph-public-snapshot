package codeintel

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
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

func (j *lsifuploadstoreExpirer) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	uploadStore, err := lsifuploadstore.New(context.Background(), lsifuploadstoreExpirerConfigInst.LSIFUploadStoreConfig, observation.ContextWithLogger(logger))
	if err != nil {
		logger.Fatal("Failed to create upload store", log.Error(err))
	}

	return []goroutine.BackgroundRoutine{
		uploadstore.NewExpirer(context.Background(), uploadStore, lsifuploadstoreExpirerConfigInst.prefix, lsifuploadstoreExpirerConfigInst.maxAge),
	}, nil
}

type lsifuploadstoreExpirerConfig struct {
	env.BaseConfig
	LSIFUploadStoreConfig *lsifuploadstore.Config

	prefix string
	maxAge time.Duration
}

var lsifuploadstoreExpirerConfigInst = &lsifuploadstoreExpirerConfig{}

func (c *lsifuploadstoreExpirerConfig) Load() {
	c.LSIFUploadStoreConfig = &lsifuploadstore.Config{}
	c.LSIFUploadStoreConfig.Load()

	c.prefix = c.GetOptional("CODEINTEL_UPLOADSTORE_EXPIRER_PREFIX", "The prefix of objects to expire in teh precise code intel upload bucket.")
	c.maxAge = c.GetInterval("CODEINTEL_UPLOADSTORE_EXPIRER_MAX_AGE", "168h", "The max age of objects in the precise code intel upload bucket.")
}
