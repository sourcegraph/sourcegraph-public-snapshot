package cliutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetRawServiceVersion returns the frontend service version information for the given runner as a raw string.
func GetRawServiceVersion(ctx context.Context, r *runner.Runner) (_ string, ok bool, _ error) {
	db, err := store.ExtractDatabase(ctx, r)
	if err != nil {
		return "", false, err
	}

	return upgradestore.New(db).GetServiceVersion(ctx)
}

// GetServiceVersion returns the frontend service version information for the given runner as a parsed version.
// Both of the return values `ok` and `error` should be checked to ensure a valid version is returned.
func GetServiceVersion(ctx context.Context, r *runner.Runner) (_ oobmigration.Version, patch int, ok bool, _ error) {
	versionStr, ok, err := GetRawServiceVersion(ctx, r)
	if err != nil {
		return oobmigration.Version{}, 0, false, err
	}
	if !ok {
		return oobmigration.Version{}, 0, false, err
	}

	version, patch, ok := oobmigration.NewVersionAndPatchFromString(versionStr)
	if !ok {
		return oobmigration.Version{}, 0, false, errors.Newf("cannot parse version: %q - expected [v]X.Y[.Z]", versionStr)
	}

	return version, patch, true, nil
}
