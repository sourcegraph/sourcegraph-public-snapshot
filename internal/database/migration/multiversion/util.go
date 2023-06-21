package multiversion

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetServiceVersion(ctx context.Context, db database.DB) (oobmigration.Version, int, bool, error) {
	versionStr, ok, err := upgradestore.New(db).GetServiceVersion(ctx)
	if err != nil || !ok {
		return oobmigration.Version{}, 0, ok, err
	}

	version, patch, ok := oobmigration.NewVersionAndPatchFromString(versionStr)
	if !ok {
		return oobmigration.Version{}, 0, ok, errors.Newf("cannot parse version: %q - expected [v]X.Y[.Z]", versionStr)
	}

	return version, patch, true, nil
}
