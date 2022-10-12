package migrations

import (
	"context"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// migrator configures an out of band migration runner process to execute in the background.
type migrator struct {
	registerMigrators  oobmigration.RegisterMigratorsFunc
	observationContext *observation.Context
}

var _ job.Job = &migrator{}

func NewMigrator(registerMigrators oobmigration.RegisterMigratorsFunc) job.Job {
	return &migrator{
		registerMigrators:  registerMigrators,
		observationContext: &observation.TestContext,
	}
}

func (m *migrator) Description() string {
	return ""
}

func (m *migrator) Config() []env.Config {
	return nil
}

func (m *migrator) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDBWithLogger(logger, m.observationContext)
	if err != nil {
		return nil, err
	}

	outOfBandMigrationRunner := oobmigration.NewRunnerWithDB(db, oobmigration.RefreshInterval, m.observationContext)

	if outOfBandMigrationRunner.SynchronizeMetadata(startupCtx); err != nil {
		return nil, errors.Wrap(err, "failed to synchronized out of band migration metadata")
	}

	if err := m.registerMigrators(startupCtx, db, outOfBandMigrationRunner); err != nil {
		return nil, err
	}

	if os.Getenv("SRC_DISABLE_OOBMIGRATION_VALIDATION") != "" {
		logger.Warn("Skipping out-of-band migrations check")
	} else {
		if err := oobmigration.ValidateOutOfBandMigrationRunner(startupCtx, db, outOfBandMigrationRunner); err != nil {
			return nil, err
		}
	}

	return []goroutine.BackgroundRoutine{
		outOfBandMigrationRunner,
	}, nil
}
