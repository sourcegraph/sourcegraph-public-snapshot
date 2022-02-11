package oobmigration

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/semver"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const RefreshInterval = time.Second * 30

// ValidateOutOfBandMigrationRunner ensures that the current application can run given its
// current migration progress. This method will return an error when there are migrations
// left in an unexpected state for the current application version.
func ValidateOutOfBandMigrationRunner(ctx context.Context, db database.DB, runner *Runner) error {
	if version.IsDev(version.Version()) {
		log15.Warn("Skipping out-of-band migrations check (dev mode)", "version", version.Version())
		return nil
	}
	currentVersionSemver, err := semver.NewVersion(version.Version())
	if err != nil {
		log15.Warn("Skipping out-of-band migrations check", "version", version.Version(), "error", err)
		return nil
	}

	// TODO - re-implement without importing cmd/frontend
	firstSemverString, err := backend.GetFirstServiceVersion(ctx, db, "frontend")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log15.Warn("Skipping out-of-band migrations check (fresh instance)", "version", version.Version())
			return nil
		}

		return errors.Wrap(err, "failed to retrieve first instance version")
	}
	firstVersionSemver, err := semver.NewVersion(firstSemverString)
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
	return runner.Validate(ctx, toVersion(currentVersionSemver), toVersion(firstVersionSemver))
}

func toVersion(version *semver.Version) Version {
	return NewVersion(int(version.Major()), int(version.Minor()))
}
