package encryption

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type recordEncrypterJob struct {
	observationContext *observation.Context
}

func NewRecordEncrypterJob(observationContext *observation.Context) job.Job {
	return &recordEncrypterJob{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *recordEncrypterJob) Description() string {
	return "encrypter routines"
}

func (j *recordEncrypterJob) Config() []env.Config {
	return []env.Config{
		ConfigInst,
	}
}

func (j *recordEncrypterJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	metrics := newMetrics(j.observationContext)

	db, err := workerdb.InitDBWithLogger(observation.ContextWithLogger(logger, j.observationContext))
	if err != nil {
		return nil, err
	}
	store := database.NewRecordEncrypter(db)

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.EncryptionInterval, &recordEncrypter{
			store:   store,
			decrypt: ConfigInst.Decrypt,
			metrics: metrics,
			logger:  logger,
		}),
		goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.MetricsInterval, &recordCounter{
			store:   store,
			metrics: metrics,
			logger:  logger,
		}),
	}, nil
}
