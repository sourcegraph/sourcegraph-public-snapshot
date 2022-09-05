package migrations

import (
	"context"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// migrator configures an out of band migration runner process to execute in the background.
type migrator struct {
	registerMigrators oobmigration.RegisterMigratorsFunc
}

var _ job.Job = &migrator{}

func NewMigrator(registerMigrators oobmigration.RegisterMigratorsFunc) job.Job {
	return &migrator{
		registerMigrators: registerMigrators,
	}
}

func (m *migrator) Description() string {
	return ""
}

func (m *migrator) Config() []env.Config {
	return nil
}

func (m *migrator) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	sqlDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(logger, sqlDB)

	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "migrator routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	outOfBandMigrationRunner := oobmigration.NewRunnerWithDB(db, oobmigration.RefreshInterval, observationContext)

	if outOfBandMigrationRunner.SynchronizeMetadata(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to synchronized out of band migration metadata")
	}

	if err := m.registerMigrators(ctx, db, outOfBandMigrationRunner); err != nil {
		return nil, err
	}

	if os.Getenv("SRC_DISABLE_OOBMIGRATION_VALIDATION") != "" {
		logger.Warn("Skipping out-of-band migrations check")
	} else {
		if err := oobmigration.ValidateOutOfBandMigrationRunner(ctx, db, outOfBandMigrationRunner); err != nil {
			return nil, err
		}
	}

	return []goroutine.BackgroundRoutine{
		outOfBandMigrationRunner,
	}, nil
}
