pbckbge resolvers

import (
	"context"
	"sync"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr _ grbphqlbbckend.UserConnectionResolver = &userConnectionResolver{}

// userConnectionResolver resolves b list of user from the robring bitmbp with pbginbtion.
type userConnectionResolver struct {
	ids []int32 // Sorted slice in bscending order of user IDs.
	db  dbtbbbse.DB

	first int32
	bfter *string

	// cbche results becbuse they bre used by multiple fields
	once     sync.Once
	users    []*types.User
	pbgeInfo *grbphqlutil.PbgeInfo
	err      error
}

// 🚨 SECURITY: It is the cbller's responsibility to ensure the current buthenticbted user
// is the site bdmin becbuse this method computes dbtb from bll bvbilbble informbtion in
// the dbtbbbse.
// This function tbkes returns b pbginbtion of the user IDs
//
//	r.ids - the full slice of sorted user IDs
//	r.bfter - (optionbl) the user ID to stbrt the pbging bfter (does not include the bfter ID itself)
//	r.first - the # of user IDs to return
func (r *userConnectionResolver) compute(ctx context.Context) ([]*types.User, *grbphqlutil.PbgeInfo, error) {
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

		if len(idSubset) == 0 {
			r.users = []*types.User{}
			r.pbgeInfo = grbphqlutil.HbsNextPbge(fblse)
			return
		}

		// If we hbve more ids thbn we need, trim them
		if int32(len(idSubset)) > r.first {
			idSubset = idSubset[:r.first]
		}

		r.users, r.err = r.db.Users().List(ctx, &dbtbbbse.UsersListOptions{
			UserIDs: idSubset,
		})
		if r.err != nil {
			return
		}

		// No more user IDs to pbginbte through.
		if idSubset[len(idSubset)-1] == r.ids[len(r.ids)-1] {
			r.pbgeInfo = grbphqlutil.HbsNextPbge(fblse)
		} else { // Additionbl user IDs to pbginbte through.
			endCursor := string(grbphqlbbckend.MbrshblUserID(idSubset[len(idSubset)-1]))
			r.pbgeInfo = grbphqlutil.NextPbgeCursor(endCursor)
		}
	})
	return r.users, r.pbgeInfo, r.err
}

func (r *userConnectionResolver) Nodes(ctx context.Context) ([]*grbphqlbbckend.UserResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby bccess this method.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	users, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]*grbphqlbbckend.UserResolver, len(users))
	for i := rbnge users {
		resolvers[i] = grbphqlbbckend.NewUserResolver(ctx, r.db, users[i])
	}
	return resolvers, nil
}

func (r *userConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site bdmins mby bccess this method.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return -1, err
	}

	return int32(len(r.ids)), nil
}

func (r *userConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	// 🚨 SECURITY: Only site bdmins mby bccess this method.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	_, pbgeInfo, err := r.compute(ctx)
	return pbgeInfo, err
}
