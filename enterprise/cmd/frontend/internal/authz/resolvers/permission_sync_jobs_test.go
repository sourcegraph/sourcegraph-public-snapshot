package resolvers

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/stretchr/testify/require"
)

func TestPermissionSyncJobsResolver(t *testing.T) {
	ctx := context.Background()
	first := int32(1337)
	args := graphqlutil.ConnectionResolverArgs{First: &first}

	t.Run("No jobs found", func(t *testing.T) {
		db := database.NewMockDB()

		jobsStore := database.NewMockPermissionSyncJobStore()
		jobsStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{}, nil)

		db.PermissionSyncJobsFunc.SetDefaultReturn(jobsStore)

		resolver := NewPermissionSyncJobsResolver(db)
		jobsResolver, err := resolver.PermissionSyncJobs(ctx, graphqlbackend.ListPermissionSyncJobsArgs{ConnectionResolverArgs: args})
		require.NoError(t, err)
		jobs, err := jobsResolver.Nodes(ctx)
		require.NoError(t, err)
		require.Empty(t, jobs)
	})

	t.Run("One job found", func(t *testing.T) {
		db := database.NewMockDB()

		jobsStore := database.NewMockPermissionSyncJobStore()
		jobsStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{{ID: 1}}, nil)

		db.PermissionSyncJobsFunc.SetDefaultReturn(jobsStore)

		resolver := NewPermissionSyncJobsResolver(db)
		jobsResolver, err := resolver.PermissionSyncJobs(ctx, graphqlbackend.ListPermissionSyncJobsArgs{ConnectionResolverArgs: args})
		require.NoError(t, err)
		jobs, err := jobsResolver.Nodes(ctx)
		require.NoError(t, err)
		require.Len(t, jobs, 1)
		require.Equal(t, marshalPermissionSyncJobID(1), jobs[0].ID())
	})
}
