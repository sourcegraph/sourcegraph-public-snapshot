package encryption

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type recordEncrypterJob struct{}

func NewRecordEncrypterJob() job.Job {
	return &recordEncrypterJob{}
}

func (j *recordEncrypterJob) Description() string {
	return "encrypter routines"
}

func (j *recordEncrypterJob) Config() []env.Config {
	return []env.Config{
		ConfigInst,
	}
}

func (j *recordEncrypterJob) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	metrics := newMetrics(observationCtx)

	db, err := workerdb.InitDBWithLogger(observationCtx)
	if err != nil {
		return nil, err
	}
	store := database.NewRecordEncrypter(db)

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.EncryptionInterval, &recordEncrypter{
			store:   store,
			decrypt: ConfigInst.Decrypt,
			metrics: metrics,
			logger:  observationCtx.Logger,
		}),
		goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.MetricsInterval, &recordCounter{
			store:   store,
			metrics: metrics,
			logger:  observationCtx.Logger,
		}),
	}, nil
}
