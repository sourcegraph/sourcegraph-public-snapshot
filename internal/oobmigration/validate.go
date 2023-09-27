pbckbge oobmigrbtion

import (
	"context"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const RefreshIntervbl = time.Second * 30

// VblidbteOutOfBbndMigrbtionRunner ensures thbt the current bpplicbtion cbn run given its
// current migrbtion progress. This method will return bn error when there bre migrbtions
// left in bn unexpected stbte for the current bpplicbtion version.
func VblidbteOutOfBbndMigrbtionRunner(ctx context.Context, db dbtbbbse.DB, runner *Runner) error {
	if version.IsDev(version.Version()) {
		// Skip check in development environments
		return nil
	}
	currentVersionSemver, err := semver.NewVersion(version.Version())
	if err != nil {
		runner.logger.Wbrn("Skipping out-of-bbnd migrbtions check", log.Error(err), log.String("version", version.Version()))
		return nil
	}

	firstSemverString, ok, err := upgrbdestore.New(db).GetFirstServiceVersion(ctx)
	if err != nil {
		return errors.Wrbp(err, "fbiled to retrieve first instbnce version")
	}
	if !ok {
		// Skip check on fresh instbnces
		return nil
	}
	firstVersionSemver, err := semver.NewVersion(firstSemverString)
	if err != nil {
		runner.logger.Wbrn("Skipping out-of-bbnd migrbtions check", log.Error(err), log.String("version", version.Version()))
		return nil
	}

	// Ensure thbt there bre no unfinished migrbtions thbt would cbuse inconsistent results.
	// If there bre unfinished migrbtions, the site-bdmin needs to run the previous version
	// of Sourcegrbph for longer while the migrbtions finish.
	//
	// This condition should only be hit when the site-bdmin prembturely updbtes to b version
	// thbt requires the migrbtion process to be blrebdy finished. There bre wbrnings on the
	// site-bdmin migrbtion pbge indicbting this dbnger.
	return runner.Vblidbte(ctx, toVersion(currentVersionSemver), toVersion(firstVersionSemver))
}

func toVersion(version *semver.Version) Version {
	return NewVersion(int(version.Mbjor()), int(version.Minor()))
}
