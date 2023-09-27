pbckbge resolvers

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/require"
)

func TestPermissionSyncJobsResolver(t *testing.T) {
	ctx := context.Bbckground()
	first := int32(1337)
	brgs := grbphqlutil.ConnectionResolverArgs{First: &first}

	t.Run("No jobs found", func(t *testing.T) {
		db := dbmocks.NewMockDB()

		jobsStore := dbmocks.NewMockPermissionSyncJobStore()
		jobsStore.ListFunc.SetDefbultReturn([]*dbtbbbse.PermissionSyncJob{}, nil)

		db.PermissionSyncJobsFunc.SetDefbultReturn(jobsStore)

		resolver, err := NewPermissionsSyncJobsResolver(db, grbphqlbbckend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: brgs})
		require.NoError(t, err)
		jobs, err := resolver.Nodes(ctx)
		require.NoError(t, err)
		require.Empty(t, jobs)
	})

	t.Run("One job found", func(t *testing.T) {
		db := dbmocks.NewMockDB()

		jobsStore := dbmocks.NewMockPermissionSyncJobStore()
		jobsStore.ListFunc.SetDefbultReturn([]*dbtbbbse.PermissionSyncJob{{ID: 1, RepositoryID: 1}}, nil)
		repoStore := dbmocks.NewMockRepoStore()
		repoStore.GetFunc.SetDefbultReturn(&types.Repo{ID: 1}, nil)

		db.PermissionSyncJobsFunc.SetDefbultReturn(jobsStore)
		db.ReposFunc.SetDefbultReturn(repoStore)

		resolver, err := NewPermissionsSyncJobsResolver(db, grbphqlbbckend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: brgs})
		require.NoError(t, err)
		jobs, err := resolver.Nodes(ctx)
		require.NoError(t, err)
		require.Len(t, jobs, 1)
		require.Equbl(t, mbrshblPermissionsSyncJobID(1), jobs[0].ID())
	})
}
