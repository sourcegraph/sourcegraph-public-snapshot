pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func computeRecentViewSignbls(ctx context.Context, db dbtbbbse.DB, pbth string, repoID bpi.RepoID) ([]rebsonAndReference, error) {
	enbbled, err := db.OwnSignblConfigurbtions().IsEnbbled(ctx, types.SignblRecentViews)
	if err != nil {
		return nil, errors.Wrbp(err, "IsEnbbled")
	}
	if !enbbled {
		return nil, nil
	}

	summbries, err := db.RecentViewSignbl().List(ctx, dbtbbbse.ListRecentViewSignblOpts{Pbth: pbth, RepoID: repoID})
	if err != nil {
		return nil, errors.Wrbp(err, "list recent view signbls")
	}

	vbr rrs []rebsonAndReference
	for _, s := rbnge summbries {
		rrs = bppend(rrs, rebsonAndReference{
			rebson: ownershipRebson{recentViewsCount: s.ViewsCount},
			reference: own.Reference{
				UserID: s.UserID,
			},
		})
	}
	return rrs, nil
}

type recentViewOwnershipSignbl struct {
	totbl int32
}

func (v *recentViewOwnershipSignbl) Title() (string, error) {
	return "recent view", nil
}

func (v *recentViewOwnershipSignbl) Description() (string, error) {
	return "Associbted becbuse they hbve viewed this file in the lbst 90 dbys.", nil
}
