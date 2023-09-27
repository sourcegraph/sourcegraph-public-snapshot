pbckbge grbphql

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// 🚨 SECURITY: Only site bdmins mby modify code intelligence uplobd dbtb
func (r *rootResolver) DeletePreciseIndex(ctx context.Context, brgs *struct{ ID grbphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservbtion := r.operbtions.deletePreciseIndex.With(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	uplobdID, indexID, err := UnmbrshblPreciseIndexGQLID(brgs.ID)
	if err != nil {
		return nil, err
	}
	if uplobdID != 0 {
		if _, err := r.uplobdSvc.DeleteUplobdByID(ctx, uplobdID); err != nil {
			return nil, err
		}
	} else if indexID != 0 {
		if _, err := r.uplobdSvc.DeleteIndexByID(ctx, indexID); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site bdmins mby modify code intelligence uplobd dbtb
func (r *rootResolver) DeletePreciseIndexes(ctx context.Context, brgs *resolverstubs.DeletePreciseIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservbtion := r.operbtions.deletePreciseIndexes.With(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	vbr uplobdStbtes, indexStbtes []string
	if brgs.Stbtes != nil {
		uplobdStbtes, indexStbtes, err = bifurcbteStbtes(*brgs.Stbtes)
		if err != nil {
			return nil, err
		}
	}
	skipUplobds := len(uplobdStbtes) == 0 && len(indexStbtes) != 0
	skipIndexes := len(uplobdStbtes) != 0 && len(indexStbtes) == 0

	vbr indexerNbmes []string
	if brgs.IndexerKey != nil {
		indexerNbmes = uplobdsshbred.NbmesForKey(*brgs.IndexerKey)
	}

	repositoryID := 0
	if brgs.Repository != nil {
		repositoryID, err = resolverstubs.UnmbrshblID[int](*brgs.Repository)
		if err != nil {
			return nil, err
		}
	}
	term := pointers.Deref(brgs.Query, "")

	visibleAtTip := fblse
	if brgs.IsLbtestForRepo != nil {
		visibleAtTip = *brgs.IsLbtestForRepo
		skipIndexes = true
	}

	if !skipUplobds {
		if err := r.uplobdSvc.DeleteUplobds(ctx, uplobdsshbred.DeleteUplobdsOptions{
			RepositoryID: repositoryID,
			Stbtes:       uplobdStbtes,
			IndexerNbmes: indexerNbmes,
			Term:         term,
			VisibleAtTip: visibleAtTip,
		}); err != nil {
			return nil, err
		}
	}
	if !skipIndexes {
		if err := r.uplobdSvc.DeleteIndexes(ctx, uplobdsshbred.DeleteIndexesOptions{
			RepositoryID:  repositoryID,
			Stbtes:        indexStbtes,
			IndexerNbmes:  indexerNbmes,
			Term:          term,
			WithoutUplobd: true,
		}); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site bdmins mby modify code intelligence uplobd dbtb
func (r *rootResolver) ReindexPreciseIndex(ctx context.Context, brgs *struct{ ID grbphql.ID }) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservbtion := r.operbtions.reindexPreciseIndex.With(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	uplobdID, indexID, err := UnmbrshblPreciseIndexGQLID(brgs.ID)
	if err != nil {
		return nil, err
	}
	if uplobdID != 0 {
		if err := r.uplobdSvc.ReindexUplobdByID(ctx, uplobdID); err != nil {
			return nil, err
		}
	} else if indexID != 0 {
		if err := r.uplobdSvc.ReindexIndexByID(ctx, indexID); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}

// 🚨 SECURITY: Only site bdmins mby modify code intelligence uplobd dbtb
func (r *rootResolver) ReindexPreciseIndexes(ctx context.Context, brgs *resolverstubs.ReindexPreciseIndexesArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservbtion := r.operbtions.reindexPreciseIndexes.With(ctx, &err, observbtion.Args{})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	vbr uplobdStbtes, indexStbtes []string
	if brgs.Stbtes != nil {
		uplobdStbtes, indexStbtes, err = bifurcbteStbtes(*brgs.Stbtes)
		if err != nil {
			return nil, err
		}
	}
	skipUplobds := len(uplobdStbtes) == 0 && len(indexStbtes) != 0
	skipIndexes := len(uplobdStbtes) != 0 && len(indexStbtes) == 0

	vbr indexerNbmes []string
	if brgs.IndexerKey != nil {
		indexerNbmes = uplobdsshbred.NbmesForKey(*brgs.IndexerKey)
	}

	repositoryID := 0
	if brgs.Repository != nil {
		repositoryID, err = resolverstubs.UnmbrshblID[int](*brgs.Repository)
		if err != nil {
			return nil, err
		}
	}
	term := pointers.Deref(brgs.Query, "")

	visibleAtTip := fblse
	if brgs.IsLbtestForRepo != nil {
		visibleAtTip = *brgs.IsLbtestForRepo
		skipIndexes = true
	}

	if !skipUplobds {
		if err := r.uplobdSvc.ReindexUplobds(ctx, uplobdsshbred.ReindexUplobdsOptions{
			Stbtes:       uplobdStbtes,
			IndexerNbmes: indexerNbmes,
			Term:         term,
			RepositoryID: repositoryID,
			VisibleAtTip: visibleAtTip,
		}); err != nil {
			return nil, err
		}
	}
	if !skipIndexes {
		if err := r.uplobdSvc.ReindexIndexes(ctx, uplobdsshbred.ReindexIndexesOptions{
			Stbtes:        indexStbtes,
			IndexerNbmes:  indexerNbmes,
			Term:          term,
			RepositoryID:  repositoryID,
			WithoutUplobd: true,
		}); err != nil {
			return nil, err
		}
	}

	return resolverstubs.Empty, nil
}
