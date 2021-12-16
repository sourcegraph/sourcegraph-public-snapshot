package migrations

import (
	"context"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
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

func (m *migrator) Config() []env.Config {
	return nil
}

func (m *migrator) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	sqlDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(sqlDB)

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	outOfBandMigrationRunner := oobmigration.NewRunnerWithDB(db, time.Second*30, observationContext)

	if err := m.registerMigrators(db, outOfBandMigrationRunner); err != nil {
		return nil, err
	}

	if os.Getenv("SRC_DISABLE_OOBMIGRATION_VALIDATION") != "" {
		log15.Warn("Skipping out-of-band migrations check")
	} else {
		// See #29074
		if err := validateOutOfBandMigrationRunner(ctx, db, outOfBandMigrationRunner); err != nil {
			return nil, err
		}
	}

	return []goroutine.BackgroundRoutine{
		outOfBandMigrationRunner,
	}, nil
}

// validateOutOfBandMigrationRunner ensures that the current application can run given its
// current migration progress. This method will return an error when there are migrations
// left in an unexpected state for the current application version.
func validateOutOfBandMigrationRunner(ctx context.Context, db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	if version.IsDev(version.Version()) {
		log15.Warn("Skipping out-of-band migrations check (dev mode)", "version", version.Version())
		return nil
	}
	currentSemver, err := semver.NewVersion(version.Version())
	if err != nil {
		log15.Warn("Skipping out-of-band migrations check", "version", version.Version(), "error", err)
		return nil
	}

	firstVersion, err := backend.GetFirstServiceVersion(ctx, db, "frontend")
	if err != nil {
		return errors.Wrap(err, "failed to retrieve first instance version")
	}
	firstVersionSemver, err := semver.NewVersion(firstVersion)
	if err != nil {
		log15.Warn("Skipping out-of-band migrations check", "version", version.Version(), "error", err)
		return nil
	}

	// Ensure that there are no unfinished migrations that would cause inconsistent results.
	// If there are unfinished migrations, the site-admin needs to run the previous version
	// of Sourcegraph for longer while the migrations finish.
	//
	// This condition should only be hit when the site-admin prematurely updates to a version
	// that requires the migration process to be already finished. There are warnings on the
	// site-admin migration page indicating this danger.
	return outOfBandMigrationRunner.Validate(
		ctx,
		oobmigration.NewVersion(int(currentSemver.Major()), int(currentSemver.Minor())),
		oobmigration.NewVersion(int(firstVersionSemver.Major()), int(firstVersionSemver.Minor())),
	)
}
