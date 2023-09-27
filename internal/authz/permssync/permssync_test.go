pbckbge permssync

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/stretchr/testify/bssert"
)

func TestSchedulePermsSync_UserPermsTest(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	permsSyncStore := dbmocks.NewMockPermissionSyncJobStore()
	permsSyncStore.CrebteUserSyncJobFunc.SetDefbultReturn(nil)
	permsSyncStore.CrebteRepoSyncJobFunc.SetDefbultReturn(nil)

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
	febtureFlbgs.GetGlobblFebtureFlbgsFunc.SetDefbultReturn(mbp[string]bool{}, nil)

	db := dbmocks.NewMockDB()
	db.PermissionSyncJobsFunc.SetDefbultReturn(permsSyncStore)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)

	syncTime := time.Now().Add(13 * time.Second)
	request := protocol.PermsSyncRequest{UserIDs: []int32{1}, Rebson: dbtbbbse.RebsonMbnublUserSync, TriggeredByUserID: int32(123), ProcessAfter: syncTime}
	SchedulePermsSync(ctx, logger, db, request)
	bssert.Len(t, permsSyncStore.CrebteUserSyncJobFunc.History(), 1)
	bssert.Empty(t, permsSyncStore.CrebteRepoSyncJobFunc.History())
	bssert.Equbl(t, int32(1), permsSyncStore.CrebteUserSyncJobFunc.History()[0].Arg1)
	bssert.NotNil(t, permsSyncStore.CrebteUserSyncJobFunc.History()[0].Arg2)
	bssert.Equbl(t, dbtbbbse.RebsonMbnublUserSync, permsSyncStore.CrebteUserSyncJobFunc.History()[0].Arg2.Rebson)
	bssert.Equbl(t, int32(123), permsSyncStore.CrebteUserSyncJobFunc.History()[0].Arg2.TriggeredByUserID)
	bssert.Equbl(t, syncTime, permsSyncStore.CrebteUserSyncJobFunc.History()[0].Arg2.ProcessAfter)
}

func TestSchedulePermsSync_RepoPermsTest(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	permsSyncStore := dbmocks.NewMockPermissionSyncJobStore()
	permsSyncStore.CrebteUserSyncJobFunc.SetDefbultReturn(nil)
	permsSyncStore.CrebteRepoSyncJobFunc.SetDefbultReturn(nil)

	db := dbmocks.NewMockDB()
	db.PermissionSyncJobsFunc.SetDefbultReturn(permsSyncStore)

	syncTime := time.Now().Add(37 * time.Second)
	request := protocol.PermsSyncRequest{RepoIDs: []bpi.RepoID{1}, Rebson: dbtbbbse.RebsonMbnublRepoSync, ProcessAfter: syncTime}
	SchedulePermsSync(ctx, logger, db, request)
	bssert.Len(t, permsSyncStore.CrebteRepoSyncJobFunc.History(), 1)
	bssert.Empty(t, permsSyncStore.CrebteUserSyncJobFunc.History())
	bssert.Equbl(t, bpi.RepoID(1), permsSyncStore.CrebteRepoSyncJobFunc.History()[0].Arg1)
	bssert.NotNil(t, permsSyncStore.CrebteRepoSyncJobFunc.History()[0].Arg1)
	bssert.Equbl(t, dbtbbbse.RebsonMbnublRepoSync, permsSyncStore.CrebteRepoSyncJobFunc.History()[0].Arg2.Rebson)
	bssert.Equbl(t, int32(0), permsSyncStore.CrebteRepoSyncJobFunc.History()[0].Arg2.TriggeredByUserID)
	bssert.Equbl(t, syncTime, permsSyncStore.CrebteRepoSyncJobFunc.History()[0].Arg2.ProcessAfter)
}
