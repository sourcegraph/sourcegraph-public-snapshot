package permssync

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/stretchr/testify/assert"
)

func TestSchedulePermsSync(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	permsSyncStore := database.NewMockPermissionSyncJobStore()
	permsSyncStore.CreateUserSyncJobFunc.SetDefaultReturn(nil)
	permsSyncStore.CreateRepoSyncJobFunc.SetDefaultReturn(nil)

	featureFlags := database.NewMockFeatureFlagStore()
	featureFlags.GetGlobalFeatureFlagsFunc.SetDefaultReturn(map[string]bool{featureFlagName: true}, nil)

	db := database.NewMockDB()
	db.PermissionSyncJobsFunc.SetDefaultReturn(permsSyncStore)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

	tests := map[string]struct {
		request  protocol.PermsSyncRequest
		userSync bool
	}{
		"User permissions sync scheduled": {
			request:  protocol.PermsSyncRequest{UserIDs: []int32{1}},
			userSync: true,
		},
		"Repo permissions sync scheduled": {
			request: protocol.PermsSyncRequest{RepoIDs: []api.RepoID{1}},
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			SchedulePermsSync(ctx, logger, db, testCase.request)
			if testCase.userSync {
				assert.Len(t, permsSyncStore.CreateUserSyncJobFunc.History(), 1)
				assert.Empty(t, permsSyncStore.CreateRepoSyncJobFunc.History())
				assert.Equal(t, int32(1), permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg1)
			} else {
				assert.Len(t, permsSyncStore.CreateRepoSyncJobFunc.History(), 1)
				assert.Empty(t, permsSyncStore.CreateUserSyncJobFunc.History())
				assert.Equal(t, api.RepoID(1), permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg1)
			}
		})
	}
}
