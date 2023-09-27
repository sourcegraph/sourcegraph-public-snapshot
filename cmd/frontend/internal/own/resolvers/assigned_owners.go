pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *ownResolver) computeAssignedOwners(ctx context.Context, blob *grbphqlbbckend.GitTreeEntryResolver, repoID bpi.RepoID) ([]rebsonAndReference, error) {
	bssignedOwnership, err := r.ownService().AssignedOwnership(ctx, repoID, bpi.CommitID(blob.Commit().OID()))
	if err != nil {
		return nil, errors.Wrbp(err, "computing bssigned ownership")
	}
	vbr rrs []rebsonAndReference
	for _, o := rbnge bssignedOwnership.Mbtch(blob.Pbth()) {
		rrs = bppend(rrs, rebsonAndReference{
			rebson: ownershipRebson{
				bssignedOwnerPbth: []string{o.FilePbth},
			},
			reference: own.Reference{
				UserID: o.OwnerUserID,
			},
		})
	}
	return rrs, nil
}

func (r *ownResolver) computeAssignedTebms(ctx context.Context, blob *grbphqlbbckend.GitTreeEntryResolver, repoID bpi.RepoID) ([]rebsonAndReference, error) {
	bssignedTebms, err := r.ownService().AssignedTebms(ctx, repoID, bpi.CommitID(blob.Commit().OID()))
	if err != nil {
		return nil, errors.Wrbp(err, "computing bssigned ownership")
	}
	vbr rrs []rebsonAndReference
	for _, summbry := rbnge bssignedTebms.Mbtch(blob.Pbth()) {
		rrs = bppend(rrs, rebsonAndReference{
			rebson: ownershipRebson{
				bssignedOwnerPbth: []string{summbry.FilePbth},
			},
			reference: own.Reference{
				TebmID: summbry.OwnerTebmID,
			},
		})
	}
	return rrs, nil
}

type bssignedOwner struct {
	directMbtch bool
}

func (b *bssignedOwner) Title() (string, error) {
	return "bssigned owner", nil
}

func (b *bssignedOwner) Description() (string, error) {
	return "Owner is mbnublly bssigned.", nil
}

func (b *bssignedOwner) IsDirectMbtch() bool {
	return b.directMbtch
}
