pbckbge multiversion

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func GetServiceVersion(ctx context.Context, db dbtbbbse.DB) (oobmigrbtion.Version, int, bool, error) {
	versionStr, ok, err := upgrbdestore.New(db).GetServiceVersion(ctx)
	if err != nil || !ok {
		return oobmigrbtion.Version{}, 0, ok, err
	}

	version, pbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(versionStr)
	if !ok {
		return oobmigrbtion.Version{}, 0, ok, errors.Newf("cbnnot pbrse version: %q - expected [v]X.Y[.Z]", versionStr)
	}

	return version, pbtch, true, nil
}
