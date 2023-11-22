package codenav

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
)

func defaultMockRepoStore() *dbmocks.MockRepoStore {
	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetReposSetByIDsFunc.SetDefaultHook(func(ctx context.Context, ids ...api.RepoID) (map[api.RepoID]*internaltypes.Repo, error) {
		m := map[api.RepoID]*internaltypes.Repo{}
		for _, id := range ids {
			m[id] = &internaltypes.Repo{
				ID:   id,
				Name: api.RepoName(fmt.Sprintf("r%d", id)),
			}
		}

		return m, nil
	})

	return repoStore
}
