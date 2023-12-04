package permssync

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
)

func TestSchedulePermsSync_UserPermsTest(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	permsSyncStore := dbmocks.NewMockPermissionSyncJobStore()
	permsSyncStore.CreateUserSyncJobFunc.SetDefaultReturn(nil)
	permsSyncStore.CreateRepoSyncJobFunc.SetDefaultReturn(nil)

	featureFlags := dbmocks.NewMockFeatureFlagStore()
	featureFlags.GetGlobalFeatureFlagsFunc.SetDefaultReturn(map[string]bool{}, nil)

	db := dbmocks.NewMockDB()
	db.PermissionSyncJobsFunc.SetDefaultReturn(permsSyncStore)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

	syncTime := time.Now().Add(13 * time.Second)
	request := ScheduleSyncOpts{UserIDs: []int32{1}, Reason: database.ReasonManualUserSync, TriggeredByUserID: int32(123), ProcessAfter: syncTime}
	SchedulePermsSync(ctx, logger, db, request)
	assert.Len(t, permsSyncStore.CreateUserSyncJobFunc.History(), 1)
	assert.Empty(t, permsSyncStore.CreateRepoSyncJobFunc.History())
	assert.Equal(t, int32(1), permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg1)
	assert.NotNil(t, permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg2)
	assert.Equal(t, database.ReasonManualUserSync, permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg2.Reason)
	assert.Equal(t, int32(123), permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg2.TriggeredByUserID)
	assert.Equal(t, syncTime, permsSyncStore.CreateUserSyncJobFunc.History()[0].Arg2.ProcessAfter)
}

func TestSchedulePermsSync_RepoPermsTest(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	permsSyncStore := dbmocks.NewMockPermissionSyncJobStore()
	permsSyncStore.CreateUserSyncJobFunc.SetDefaultReturn(nil)
	permsSyncStore.CreateRepoSyncJobFunc.SetDefaultReturn(nil)

	db := dbmocks.NewMockDB()
	db.PermissionSyncJobsFunc.SetDefaultReturn(permsSyncStore)

	syncTime := time.Now().Add(37 * time.Second)
	request := ScheduleSyncOpts{RepoIDs: []api.RepoID{1}, Reason: database.ReasonManualRepoSync, ProcessAfter: syncTime}
	SchedulePermsSync(ctx, logger, db, request)
	assert.Len(t, permsSyncStore.CreateRepoSyncJobFunc.History(), 1)
	assert.Empty(t, permsSyncStore.CreateUserSyncJobFunc.History())
	assert.Equal(t, api.RepoID(1), permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg1)
	assert.NotNil(t, permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg1)
	assert.Equal(t, database.ReasonManualRepoSync, permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg2.Reason)
	assert.Equal(t, int32(0), permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg2.TriggeredByUserID)
	assert.Equal(t, syncTime, permsSyncStore.CreateRepoSyncJobFunc.History()[0].Arg2.ProcessAfter)
}
