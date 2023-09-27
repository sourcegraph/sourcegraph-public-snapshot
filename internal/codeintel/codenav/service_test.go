pbckbge codenbv

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func defbultMockRepoStore() *dbmocks.MockRepoStore {
	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetReposSetByIDsFunc.SetDefbultHook(func(ctx context.Context, ids ...bpi.RepoID) (mbp[bpi.RepoID]*internbltypes.Repo, error) {
		m := mbp[bpi.RepoID]*internbltypes.Repo{}
		for _, id := rbnge ids {
			m[id] = &internbltypes.Repo{
				ID:   id,
				Nbme: bpi.RepoNbme(fmt.Sprintf("r%d", id)),
			}
		}

		return m, nil
	})

	return repoStore
}
