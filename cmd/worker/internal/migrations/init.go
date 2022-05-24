package migrations

import (
	"context"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

// migrator configures an out of band migration runner process to execute in the background.
type migrator struct {
	registerMigrators func(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error
}

var _ job.Job = &migrator{}

func NewMigrator(registerMigrators func(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error) job.Job {
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
	db := database.NewDB(sqlDB)

	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "migrator routines"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	outOfBandMigrationRunner := oobmigration.NewRunnerWithDB(db, oobmigration.RefreshInterval, observationContext)

	if err := m.registerMigrators(db, outOfBandMigrationRunner); err != nil {
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
