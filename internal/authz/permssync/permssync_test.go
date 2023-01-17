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

func TestSchedulePermsSync_UserPermsTest(t *testing.T) {
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

	request := protocol.PermsSyncRequest{UserIDs: []int32{1}, Reason: ReasonManualUserSync, TriggeredByUserID: int32(123)}
	SchedulePermsSync(ctx, logger, db, request)
	assert.Len(t, permsSyncStore.CreateUserSyncJobFunc.History(), 1)
	assert.Empty(t, permsSyncStore.CreateRepoSyncJobFunc.History())
	assert.Equal(t, int32(1), permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg1)
	assert.NotNil(t, permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg2)
	assert.Equal(t, ReasonManualUserSync, permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg2.Reason)
	assert.Equal(t, int32(123), permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg2.TriggeredByUserID)
}

func TestSchedulePermsSync_RepoPermsTest(t *testing.T) {
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

	request := protocol.PermsSyncRequest{RepoIDs: []api.RepoID{1}, Reason: ReasonManualRepoSync}
	SchedulePermsSync(ctx, logger, db, request)
	assert.Len(t, permsSyncStore.CreateRepoSyncJobFunc.History(), 1)
	assert.Empty(t, permsSyncStore.CreateUserSyncJobFunc.History())
	assert.Equal(t, api.RepoID(1), permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg1)
	assert.NotNil(t, permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg1)
	assert.Equal(t, ReasonManualRepoSync, permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg2.Reason)
	assert.Equal(t, int32(0), permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg2.TriggeredByUserID)
}
