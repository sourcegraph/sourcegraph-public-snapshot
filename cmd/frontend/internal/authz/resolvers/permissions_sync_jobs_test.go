package resolvers

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestPermissionSyncJobsResolver(t *testing.T) {
	ctx := context.Background()
	first := int32(1337)
	args := graphqlutil.ConnectionResolverArgs{First: &first}

	t.Run("No jobs found", func(t *testing.T) {
		db := dbmocks.NewMockDB()

		jobsStore := dbmocks.NewMockPermissionSyncJobStore()
		jobsStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{}, nil)

		db.PermissionSyncJobsFunc.SetDefaultReturn(jobsStore)

		resolver, err := NewPermissionsSyncJobsResolver(db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
		require.NoError(t, err)
		jobs, err := resolver.Nodes(ctx)
		require.NoError(t, err)
		require.Empty(t, jobs)
	})

	t.Run("One job found", func(t *testing.T) {
		db := dbmocks.NewMockDB()

		jobsStore := dbmocks.NewMockPermissionSyncJobStore()
		jobsStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{{ID: 1, RepositoryID: 1}}, nil)
		repoStore := dbmocks.NewMockRepoStore()
		repoStore.GetFunc.SetDefaultReturn(&types.Repo{ID: 1}, nil)

		db.PermissionSyncJobsFunc.SetDefaultReturn(jobsStore)
		db.ReposFunc.SetDefaultReturn(repoStore)

		resolver, err := NewPermissionsSyncJobsResolver(db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
		require.NoError(t, err)
		jobs, err := resolver.Nodes(ctx)
		require.NoError(t, err)
		require.Len(t, jobs, 1)
		require.Equal(t, marshalPermissionsSyncJobID(1), jobs[0].ID())
	})
}
