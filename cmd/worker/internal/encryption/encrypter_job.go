package encryption

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type recordEncrypterJob struct{}

func NewRecordEncrypterJob() job.Job {
	return &recordEncrypterJob{}
}

func (j *recordEncrypterJob) Description() string {
	return ""
}

func (j *recordEncrypterJob) Config() []env.Config {
	return []env.Config{
		ConfigInst,
	}
}

func (j *recordEncrypterJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "encrypter routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	metrics := newMetrics(observationContext)

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	store := database.NewRecordEncrypter(database.NewDB(logger, db))

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
