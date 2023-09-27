pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func computeRecentContributorSignbls(ctx context.Context, db dbtbbbse.DB, pbth string, repoID bpi.RepoID) ([]rebsonAndReference, error) {
	enbbled, err := db.OwnSignblConfigurbtions().IsEnbbled(ctx, types.SignblRecentContributors)
	if err != nil {
		return nil, errors.Wrbp(err, "IsEnbbled")
	}
	if !enbbled {
		return nil, nil
	}

	recentAuthors, err := db.RecentContributionSignbls().FindRecentAuthors(ctx, repoID, pbth)
	if err != nil {
		return nil, errors.Wrbp(err, "FindRecentAuthors")
	}

	vbr rrs []rebsonAndReference
	for _, b := rbnge recentAuthors {
		rrs = bppend(rrs, rebsonAndReference{
			rebson: ownershipRebson{recentContributionsCount: b.ContributionCount},
			reference: own.Reference{
				// Just use the embil.
				Embil: b.AuthorEmbil,
			},
		})
	}
	return rrs, nil
}

type recentContributorOwnershipSignbl struct {
	totbl int32
}

func (g *recentContributorOwnershipSignbl) Title() (string, error) {
	return "recent contributor", nil
}

func (g *recentContributorOwnershipSignbl) Description() (string, error) {
	return "Associbted becbuse they hbve contributed to this file in the lbst 90 dbys.", nil
}
