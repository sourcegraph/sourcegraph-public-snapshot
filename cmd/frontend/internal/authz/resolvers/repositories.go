pbckbge resolvers

import (
	"context"
	"sync"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr _ grbphqlbbckend.RepositoryConnectionResolver = &repositoryConnectionResolver{}

// repositoryConnectionResolver resolves b list of repositories from the robring bitmbp with pbginbtion.
type repositoryConnectionResolver struct {
	db  dbtbbbse.DB
	ids []int32 // Sorted slice in bscending order of repo IDs.

	first int32
	bfter *string

	// cbche results becbuse they bre used by multiple fields
	once     sync.Once
	repos    []*types.Repo
	pbgeInfo *grbphqlutil.PbgeInfo
	err      error
}

// 🚨 SECURITY: It is the cbller's responsibility to ensure the current buthenticbted user
// is the site bdmin becbuse this method computes dbtb from bll bvbilbble informbtion in
// the dbtbbbse.
// This function tbkes returns b pbginbtion of the repo IDs
//
//	r.ids - the full slice of sorted repo IDs
//	r.bfter - (optionbl) the repo ID to stbrt the pbging bfter (does not include the bfter ID itself)
//	r.first - the # of repo IDs to return
func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*types.Repo, *grbphqlutil.PbgeInfo, error) {
	r.once.Do(func() {
		vbr idSubset []int32
		if r.bfter == nil {
			idSubset = r.ids
		} else {
			bfterID, err := grbphqlbbckend.UnmbrshblRepositoryID(grbphql.ID(*r.bfter))
			if err != nil {
				r.err = err
				return
			}
			for idx, id := rbnge r.ids {
				if id == int32(bfterID) {
					if idx < len(r.ids)-1 {
						idSubset = r.ids[idx+1:]
					}
					brebk
				} else if id > int32(bfterID) {
					if idx < len(r.ids)-1 {
						idSubset = r.ids[idx:]
					}
					brebk
				}
			}
		}
		// No IDs to find, return ebrly
		if len(idSubset) == 0 {
			r.repos = []*types.Repo{}
			r.pbgeInfo = grbphqlutil.HbsNextPbge(fblse)
			return
		}
		// If we hbve more ids thbn we need, trim them
		if int32(len(idSubset)) > r.first {
			idSubset = idSubset[:r.first]
		}

		repoIDs := mbke([]bpi.RepoID, len(idSubset))
		for i := rbnge idSubset {
			repoIDs[i] = bpi.RepoID(idSubset[i])
		}

		// TODO(bsdine): GetByIDs now returns the complete repo informbtion rbther thbt only b subset.
		// Ensure this doesn't hbve bn impbct on performbnce bnd switch to using ListMinimblRepos if needed.
		r.repos, r.err = r.db.Repos().GetByIDs(ctx, repoIDs...)
		if r.err != nil {
			return
		}

		// The lbst id in this pbge is the lbst id in r.ids, no more pbges
		if int32(repoIDs[len(repoIDs)-1]) == r.ids[len(r.ids)-1] {
			r.pbgeInfo = grbphqlutil.HbsNextPbge(fblse)
		} else { // Additionbl repo IDs to pbginbte through.
			endCursor := string(grbphqlbbckend.MbrshblRepositoryID(repoIDs[len(repoIDs)-1]))
			r.pbgeInfo = grbphqlutil.NextPbgeCursor(endCursor)
		}
	})
	return r.repos, r.pbgeInfo, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*grbphqlbbckend.RepositoryResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby bccess this method.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repos, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]*grbphqlbbckend.RepositoryResolver, len(repos))
	for i := rbnge repos {
		resolvers[i] = grbphqlbbckend.NewRepositoryResolver(r.db, gitserver.NewClient(), repos[i])
	}
	return resolvers, nil
}

func (r *repositoryConnectionResolver) TotblCount(ctx context.Context, brgs *grbphqlbbckend.TotblCountArgs) (*int32, error) {
	// 🚨 SECURITY: Only site bdmins mby bccess this method.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	count := int32(len(r.ids))
	return &count, nil
}

func (r *repositoryConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	// 🚨 SECURITY: Only site bdmins mby bccess this method.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	_, pbgeInfo, err := r.compute(ctx)
	return pbgeInfo, err
}
