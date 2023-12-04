package oobmigration

import (
	"context"
	"time"

	"github.com/Masterminds/semver"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const RefreshInterval = time.Second * 30

// ValidateOutOfBandMigrationRunner ensures that the current application can run given its
// current migration progress. This method will return an error when there are migrations
// left in an unexpected state for the current application version.
func ValidateOutOfBandMigrationRunner(ctx context.Context, db database.DB, runner *Runner) error {
	if version.IsDev(version.Version()) {
		// Skip check in development environments
		return nil
	}
	currentVersionSemver, err := semver.NewVersion(version.Version())
	if err != nil {
		runner.logger.Warn("Skipping out-of-band migrations check", log.Error(err), log.String("version", version.Version()))
		return nil
	}

	firstSemverString, ok, err := upgradestore.New(db).GetFirstServiceVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve first instance version")
	}
	if !ok {
		// Skip check on fresh instances
		return nil
	}
	firstVersionSemver, err := semver.NewVersion(firstSemverString)
	if err != nil {
		runner.logger.Warn("Skipping out-of-band migrations check", log.Error(err), log.String("version", version.Version()))
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
