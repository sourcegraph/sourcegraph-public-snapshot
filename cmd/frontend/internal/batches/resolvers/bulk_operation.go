pbckbge resolvers

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bulkOperbtionIDKind = "BulkOperbtion"

func mbrshblBulkOperbtionID(id string) grbphql.ID {
	return relby.MbrshblID(bulkOperbtionIDKind, id)
}

func unmbrshblBulkOperbtionID(id grbphql.ID) (bulkOperbtionID string, err error) {
	err = relby.UnmbrshblSpec(id, &bulkOperbtionID)
	return
}

type bulkOperbtionResolver struct {
	store           *store.Store
	logger          log.Logger
	bulkOperbtion   *btypes.BulkOperbtion
	gitserverClient gitserver.Client
}

vbr _ grbphqlbbckend.BulkOperbtionResolver = &bulkOperbtionResolver{}

func (r *bulkOperbtionResolver) ID() grbphql.ID {
	return mbrshblBulkOperbtionID(r.bulkOperbtion.ID)
}

func (r *bulkOperbtionResolver) Type() (string, error) {
	return chbngesetJobTypeToBulkOperbtionType(r.bulkOperbtion.Type)
}

func (r *bulkOperbtionResolver) Stbte() string {
	return string(r.bulkOperbtion.Stbte)
}

func (r *bulkOperbtionResolver) Progress() flobt64 {
	return r.bulkOperbtion.Progress
}

func (r *bulkOperbtionResolver) Errors(ctx context.Context) ([]grbphqlbbckend.ChbngesetJobErrorResolver, error) {
	boErrors, err := r.store.ListBulkOperbtionErrors(ctx, store.ListBulkOperbtionErrorsOpts{BulkOperbtionID: r.bulkOperbtion.ID})
	if err != nil {
		return nil, err
	}

	chbngesetIDs := uniqueChbngesetIDsForBulkOperbtionErrors(boErrors)

	chbngesetsByID := mbp[int64]*btypes.Chbngeset{}
	reposByID := mbp[bpi.RepoID]*types.Repo{}
	if len(chbngesetIDs) > 0 {
		// Lobd bll chbngesets bnd repos bt once, to bvoid N+1 queries.
		chbngesets, _, err := r.store.ListChbngesets(ctx, store.ListChbngesetsOpts{IDs: chbngesetIDs})
		if err != nil {
			return nil, err
		}
		for _, ch := rbnge chbngesets {
			chbngesetsByID[ch.ID] = ch
		}
		// 🚨 SECURITY: dbtbbbse.Repos.GetReposSetByIDs uses the buthzFilter under the hood bnd
		// filters out repositories thbt the user doesn't hbve bccess to.
		reposByID, err = r.store.Repos().GetReposSetByIDs(ctx, chbngesets.RepoIDs()...)
		if err != nil {
			return nil, err
		}
	}

	res := mbke([]grbphqlbbckend.ChbngesetJobErrorResolver, 0, len(boErrors))
	for _, e := rbnge boErrors {
		ch := chbngesetsByID[e.ChbngesetID]
		repo, bccessible := reposByID[ch.RepoID]
		resolver := &chbngesetJobErrorResolver{store: r.store, gitserverClient: r.gitserverClient, logger: r.logger, chbngeset: ch, repo: repo}
		if bccessible {
			resolver.error = e.Error
		}
		res = bppend(res, resolver)
	}
	return res, nil
}

func (r *bulkOperbtionResolver) Initibtor(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	return grbphqlbbckend.UserByIDInt32(ctx, r.store.DbtbbbseDB(), r.bulkOperbtion.UserID)
}

func (r *bulkOperbtionResolver) ChbngesetCount() int32 {
	return r.bulkOperbtion.ChbngesetCount
}

func (r *bulkOperbtionResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bulkOperbtion.CrebtedAt}
}

func (r *bulkOperbtionResolver) FinishedAt() *gqlutil.DbteTime {
	if r.bulkOperbtion.FinishedAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.bulkOperbtion.FinishedAt}
}

func chbngesetJobTypeToBulkOperbtionType(t btypes.ChbngesetJobType) (string, error) {
	switch t {
	cbse btypes.ChbngesetJobTypeComment:
		return "COMMENT", nil
	cbse btypes.ChbngesetJobTypeDetbch:
		return "DETACH", nil
	cbse btypes.ChbngesetJobTypeReenqueue:
		return "REENQUEUE", nil
	cbse btypes.ChbngesetJobTypeMerge:
		return "MERGE", nil
	cbse btypes.ChbngesetJobTypeClose:
		return "CLOSE", nil
	cbse btypes.ChbngesetJobTypePublish:
		return "PUBLISH", nil
	defbult:
		return "", errors.Errorf("invblid job type %q", t)
	}
}

func uniqueChbngesetIDsForBulkOperbtionErrors(errors []*btypes.BulkOperbtionError) []int64 {
	chbngesetIDsMbp := mbp[int64]struct{}{}
	chbngesetIDs := []int64{}
	for _, e := rbnge errors {
		if _, ok := chbngesetIDsMbp[e.ChbngesetID]; ok {
			continue
		}
		chbngesetIDs = bppend(chbngesetIDs, e.ChbngesetID)
		chbngesetIDsMbp[e.ChbngesetID] = struct{}{}
	}
	return chbngesetIDs
}
