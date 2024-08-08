package resolvers

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestPermissionSyncJobsResolver(t *testing.T) {
	ctx := context.Background()
	first := int32(1337)
	args := gqlutil.ConnectionResolverArgs{First: &first}

	t.Run("No jobs found", func(t *testing.T) {
		db := dbmocks.NewMockDB()

		jobsStore := dbmocks.NewMockPermissionSyncJobStore()
		jobsStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{}, nil)

		db.PermissionSyncJobsFunc.SetDefaultReturn(jobsStore)
		logger := logtest.NoOp(t)

		resolver, err := NewPermissionsSyncJobsResolver(logger, db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
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
		logger := logtest.NoOp(t)

		resolver, err := NewPermissionsSyncJobsResolver(logger, db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
		require.NoError(t, err)
		jobs, err := resolver.Nodes(ctx)
		require.NoError(t, err)
		require.Len(t, jobs, 1)
		require.Equal(t, marshalPermissionsSyncJobID(1), jobs[0].ID())
	})

	t.Run("logs non-not-found errors", func(t *testing.T) {
		t.Run("repository error", func(t *testing.T) {
			// Create a mock database
			db := dbmocks.NewMockDB()

			// Create a mock permissions sync job store
			jobStore := dbmocks.NewMockPermissionSyncJobStore()

			// Set up the expectations for the job store
			job := &database.PermissionSyncJob{ID: 1, RepositoryID: 1}
			jobStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{job}, nil)

			db.PermissionSyncJobsFunc.SetDefaultReturn(jobStore)

			unexpectedError := errors.New("this is a test error")

			reposStore := dbmocks.NewMockRepoStore()
			reposStore.GetFunc.SetDefaultReturn(nil, unexpectedError)
			db.ReposFunc.SetDefaultReturn(reposStore)

			logger, dumpLogs := logtest.Captured(t)
			// Create a permissions sync job connection store
			resolver, err := NewPermissionsSyncJobsResolver(logger, db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
			require.NoError(t, err)

			nodes, err := resolver.Nodes(ctx)
			require.NoError(t, err) // Assert that no error is returned
			require.Equal(t, 0, len(nodes))

			// Assert that the error is logged
			logEntries := dumpLogs()
			var foundLogEntry bool
			for _, entry := range logEntries {
				errText, ok := entry.Fields["error"].(string)

				if ok && strings.Contains(errText, unexpectedError.Error()) {
					foundLogEntry = true
					break
				}
			}

			require.True(t, foundLogEntry, "Expected log entry not found")
		})

		t.Run("user error", func(t *testing.T) {
			// Create a mock database
			db := dbmocks.NewMockDB()

			// Create a mock permissions sync job store
			jobStore := dbmocks.NewMockPermissionSyncJobStore()

			// Set up the expectations for the job store
			job := &database.PermissionSyncJob{ID: 1, UserID: 1}
			jobStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{job}, nil)

			db.PermissionSyncJobsFunc.SetDefaultReturn(jobStore)

			unexpectedError := errors.New("this is a test error")

			userStore := dbmocks.NewMockUserStore()
			userStore.GetByIDFunc.SetDefaultReturn(nil, unexpectedError)
			db.UsersFunc.SetDefaultReturn(userStore)

			logger, dumpLogs := logtest.Captured(t)
			// Create a permissions sync job connection store
			resolver, err := NewPermissionsSyncJobsResolver(logger, db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
			require.NoError(t, err)

			nodes, err := resolver.Nodes(ctx)
			require.NoError(t, err)
			require.Equal(t, 0, len(nodes))

			// Assert that the error is logged
			logEntries := dumpLogs()
			var foundLogEntry bool
			for _, entry := range logEntries {
				errText, ok := entry.Fields["error"].(string)

				if ok && strings.Contains(errText, unexpectedError.Error()) {
					foundLogEntry = true
					break
				}
			}

			require.True(t, foundLogEntry, "Expected log entry not found")
		})
	})

	t.Run("skips over not found errors", func(t *testing.T) {
		t.Run("repository error", func(t *testing.T) {

			// Create a mock
			db := dbmocks.NewMockDB()

			// Create a mock permissions sync job store
			jobStore := dbmocks.NewMockPermissionSyncJobStore()

			// Set up the expectations for the job store
			job := &database.PermissionSyncJob{ID: 1, RepositoryID: 1}
			jobStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{job}, nil)

			db.PermissionSyncJobsFunc.SetDefaultReturn(jobStore)

			notFoundError := testNotFoundError{}

			repoStore := dbmocks.NewMockRepoStore()
			repoStore.GetFunc.SetDefaultReturn(nil, notFoundError)
			db.ReposFunc.SetDefaultReturn(repoStore)

			logger, dumpLogs := logtest.Captured(t)
			// Create a permissions sync job connection store
			resolver, err := NewPermissionsSyncJobsResolver(logger, db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
			require.NoError(t, err)

			nodes, err := resolver.Nodes(ctx)
			require.NoError(t, err) // Assert that no error is returned
			require.Equal(t, 0, len(nodes))

			foundError := false
			for _, entry := range dumpLogs() {
				if strings.Contains(entry.Message, notFoundError.Error()) {
					foundError = true
					break
				}

				errText, ok := entry.Fields["error"].(string)
				if ok && strings.Contains(errText, notFoundError.Error()) {
					foundError = true
					break
				}
			}

			require.False(t, foundError, "Unexpected error found in logs")
		})

		t.Run("user error", func(t *testing.T) {

			// Create a mock
			db := dbmocks.NewMockDB()

			// Create a mock permissions sync job store
			jobStore := dbmocks.NewMockPermissionSyncJobStore()

			// Set up the expectations for the job store
			job := &database.PermissionSyncJob{ID: 1, RepositoryID: 1}
			jobStore.ListFunc.SetDefaultReturn([]*database.PermissionSyncJob{job}, nil)

			db.PermissionSyncJobsFunc.SetDefaultReturn(jobStore)

			notFoundError := testNotFoundError{}

			userStore := dbmocks.NewMockRepoStore()
			userStore.GetFunc.SetDefaultReturn(nil, notFoundError)
			db.ReposFunc.SetDefaultReturn(userStore)

			logger, dumpLogs := logtest.Captured(t)
			// Create a permissions sync job connection store
			resolver, err := NewPermissionsSyncJobsResolver(logger, db, graphqlbackend.ListPermissionsSyncJobsArgs{ConnectionResolverArgs: args})
			require.NoError(t, err)

			nodes, err := resolver.Nodes(ctx)
			require.NoError(t, err) // Assert that no error is returned
			require.Equal(t, 0, len(nodes))

			foundError := false
			for _, entry := range dumpLogs() {
				if strings.Contains(entry.Message, notFoundError.Error()) {
					foundError = true
					break
				}

				errText, ok := entry.Fields["error"].(string)
				if ok && strings.Contains(errText, notFoundError.Error()) {
					foundError = true
					break
				}
			}

			require.False(t, foundError, "Unexpected error found in logs")
		})
	})
}

type testNotFoundError struct{}

func (testNotFoundError) NotFound() bool {
	return true
}

func (testNotFoundError) Error() string {
	return "this is a test not found error"
}

var _ error = testNotFoundError{}
