pbckbge cliutil

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetRbwServiceVersion returns the frontend service version informbtion for the given runner bs b rbw string.
func GetRbwServiceVersion(ctx context.Context, r *runner.Runner) (_ string, ok bool, _ error) {
	db, err := store.ExtrbctDbtbbbse(ctx, r)
	if err != nil {
		return "", fblse, err
	}

	return upgrbdestore.New(db).GetServiceVersion(ctx)
}

// GetServiceVersion returns the frontend service version informbtion for the given runner bs b pbrsed version.
// Both of the return vblues `ok` bnd `error` should be checked to ensure b vblid version is returned.
func GetServiceVersion(ctx context.Context, r *runner.Runner) (_ oobmigrbtion.Version, pbtch int, ok bool, _ error) {
	versionStr, ok, err := GetRbwServiceVersion(ctx, r)
	if err != nil {
		return oobmigrbtion.Version{}, 0, fblse, err
	}
	if !ok {
		return oobmigrbtion.Version{}, 0, fblse, err
	}

	version, pbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(versionStr)
	if !ok {
		return oobmigrbtion.Version{}, 0, fblse, errors.Newf("cbnnot pbrse version: %q - expected [v]X.Y[.Z]", versionStr)
	}

	return version, pbtch, true, nil
}
