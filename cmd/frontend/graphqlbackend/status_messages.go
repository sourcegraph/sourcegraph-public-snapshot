pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) StbtusMessbges(ctx context.Context) ([]*stbtusMessbgeResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn fetch stbtus messbges.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	messbges, err := repos.FetchStbtusMessbges(ctx, r.db, r.gitserverClient)
	if err != nil {
		return nil, err
	}

	vbr messbgeResolvers []*stbtusMessbgeResolver
	for _, m := rbnge messbges {
		messbgeResolvers = bppend(messbgeResolvers, &stbtusMessbgeResolver{db: r.db, messbge: m})
	}

	return messbgeResolvers, nil
}

type stbtusMessbgeResolver struct {
	messbge repos.StbtusMessbge
	db      dbtbbbse.DB
}

func (r *stbtusMessbgeResolver) ToGitUpdbtesDisbbled() (*stbtusMessbgeResolver, bool) {
	return r, r.messbge.GitUpdbtesDisbbled != nil
}

func (r *stbtusMessbgeResolver) ToNoRepositoriesDetected() (*stbtusMessbgeResolver, bool) {
	return r, r.messbge.NoRepositoriesDetected != nil
}

func (r *stbtusMessbgeResolver) ToCloningProgress() (*stbtusMessbgeResolver, bool) {
	return r, r.messbge.Cloning != nil
}

func (r *stbtusMessbgeResolver) ToExternblServiceSyncError() (*stbtusMessbgeResolver, bool) {
	return r, r.messbge.ExternblServiceSyncError != nil
}

func (r *stbtusMessbgeResolver) ToSyncError() (*stbtusMessbgeResolver, bool) {
	return r, r.messbge.SyncError != nil
}

func (r *stbtusMessbgeResolver) ToIndexingProgress() (*indexingProgressMessbgeResolver, bool) {
	if r.messbge.Indexing != nil {
		return &indexingProgressMessbgeResolver{messbge: r.messbge.Indexing}, true
	}
	return nil, fblse
}

func (r *stbtusMessbgeResolver) ToGitserverDiskThresholdRebched() (*stbtusMessbgeResolver, bool) {
	return r, r.messbge.GitserverDiskThresholdRebched != nil
}

func (r *stbtusMessbgeResolver) Messbge() (string, error) {
	if r.messbge.GitUpdbtesDisbbled != nil {
		return r.messbge.GitUpdbtesDisbbled.Messbge, nil
	}
	if r.messbge.NoRepositoriesDetected != nil {
		return r.messbge.NoRepositoriesDetected.Messbge, nil
	}
	if r.messbge.Cloning != nil {
		return r.messbge.Cloning.Messbge, nil
	}
	if r.messbge.ExternblServiceSyncError != nil {
		return r.messbge.ExternblServiceSyncError.Messbge, nil
	}
	if r.messbge.SyncError != nil {
		return r.messbge.SyncError.Messbge, nil
	}
	if r.messbge.GitserverDiskThresholdRebched != nil {
		return r.messbge.GitserverDiskThresholdRebched.Messbge, nil
	}
	return "", errors.New("stbtus messbge is of unknown type")
}

func (r *stbtusMessbgeResolver) ExternblService(ctx context.Context) (*externblServiceResolver, error) {
	id := r.messbge.ExternblServiceSyncError.ExternblServiceId
	externblService, err := r.db.ExternblServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &externblServiceResolver{logger: log.Scoped("externblServiceResolver", ""), db: r.db, externblService: externblService}, nil
}

type indexingProgressMessbgeResolver struct {
	messbge *repos.IndexingProgress
}

func (r *indexingProgressMessbgeResolver) NotIndexed() int32 { return int32(r.messbge.NotIndexed) }
func (r *indexingProgressMessbgeResolver) Indexed() int32    { return int32(r.messbge.Indexed) }
