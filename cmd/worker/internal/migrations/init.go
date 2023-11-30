package migrations

import (
	"context"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// migrator configures an out of band migration runner process to execute in the background.
type migrator struct {
	registerMigrators oobmigration.RegisterMigratorsFunc
}

var _ job.Job = &migrator{}

func NewMigrator(registerMigrators oobmigration.RegisterMigratorsFunc) job.Job {
	return &migrator{registerMigrators}
}

func (m *migrator) Description() string {
	return ""
}

func (m *migrator) Config() []env.Config {
	return nil
}

func (m *migrator) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	outOfBandMigrationRunner := oobmigration.NewRunnerWithDB(observationCtx, db, oobmigration.RefreshInterval)

	if err := outOfBandMigrationRunner.SynchronizeMetadata(startupCtx); err != nil {
		return nil, errors.Wrap(err, "failed to synchronize out of band migration metadata")
	}

	if err := m.registerMigrators(startupCtx, db, outOfBandMigrationRunner); err != nil {
		return nil, err
	}

	if os.Getenv("SRC_DISABLE_OOBMIGRATION_VALIDATION") != "" {
		if !deploy.IsSingleBinary() {
			observationCtx.Logger.Warn("Skipping out-of-band migrations check")
		}
	} else {
		if err := oobmigration.ValidateOutOfBandMigrationRunner(startupCtx, db, outOfBandMigrationRunner); err != nil {
			return nil, err
		}
	}

	version, err := currentVersion(observationCtx.Logger)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		&outOfBandMigrationRunnerWrapper{Runner: outOfBandMigrationRunner, version: version},
	}, nil
}

type outOfBandMigrationRunnerWrapper struct {
	*oobmigration.Runner
	version oobmigration.Version
}

func (w *outOfBandMigrationRunnerWrapper) Start() {
	w.Runner.Start(w.version)
}
