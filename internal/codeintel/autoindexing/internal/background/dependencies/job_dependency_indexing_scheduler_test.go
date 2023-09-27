pbckbge dependencies

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestDependencyIndexingSchedulerHbndler(t *testing.T) {
	mockUplobdsSvc := NewMockUplobdService()
	mockRepoStore := NewMockReposStore()
	mockExtSvcStore := NewMockExternblServiceStore()
	mockRepoUpdbter := NewMockRepoUpdbterClient()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockWorkerStore := NewMockWorkerStore[dependencyIndexingJob]()

	mockRepoStore.ListMinimblReposFunc.PushReturn([]types.MinimblRepo{
		{
			ID:    0,
			Nbme:  "",
			Stbrs: 0,
		},
	}, nil)

	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)

	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v2.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v3.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v3.2.2"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v2.2.1"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v4.2.3"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v1.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/bbnbnb/world", Version: "v0.0.1"}}, true, nil)
	mockScbnner.NextFunc.SetDefbultReturn(shbred.PbckbgeReference{}, fblse, nil)

	mockGitserverReposStore := NewMockGitserverRepoStore()
	mockGitserverReposStore.GetByNbmesFunc.PushReturn(mbp[bpi.RepoNbme]*types.GitserverRepo{
		"github.com/sbmple/text": {
			CloneStbtus: types.CloneStbtusCloned,
		},
		"github.com/cheese/burger": {
			CloneStbtus: types.CloneStbtusCloned,
		},
		"github.com/bbnbnb/world": {
			CloneStbtus: types.CloneStbtusCloned,
		},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	envvbr.MockSourcegrbphDotComMode(true)

	hbndler := &dependencyIndexingSchedulerHbndler{
		uplobdsSvc:         mockUplobdsSvc,
		repoStore:          mockRepoStore,
		indexEnqueuer:      indexEnqueuer,
		extsvcStore:        mockExtSvcStore,
		workerStore:        mockWorkerStore,
		gitserverRepoStore: mockGitserverReposStore,
		repoUpdbter:        mockRepoUpdbter,
	}

	logger := logtest.Scoped(t)
	job := dependencyIndexingJob{
		UplobdID:            42,
		ExternblServiceKind: "",
		ExternblServiceSync: time.Time{},
	}
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(mockExtSvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to extsvcStore.List. wbnt=%d hbve=%d", 0, len(mockExtSvcStore.ListFunc.History()))
	}

	if len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()) != 7 {
		t.Errorf("unexpected number of cblls to QueueIndexesForPbckbge. wbnt=%d hbve=%d", 6, len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()))
	} else {
		vbr pbckbges []dependencies.MinimiblVersionedPbckbgeRepo
		for _, cbll := rbnge indexEnqueuer.QueueIndexesForPbckbgeFunc.History() {
			pbckbges = bppend(pbckbges, cbll.Arg1)
		}
		sort.Slice(pbckbges, func(i, j int) bool {
			for _, pbir := rbnge [][2]string{
				{pbckbges[i].Scheme, pbckbges[j].Scheme},
				{string(pbckbges[i].Nbme), string(pbckbges[j].Nbme)},
				{pbckbges[i].Version, pbckbges[j].Version},
			} {
				if pbir[0] < pbir[1] {
					return true
				}
				if pbir[1] < pbir[0] {
					brebk
				}
			}

			return fblse
		})

		expectedPbckbges := []dependencies.MinimiblVersionedPbckbgeRepo{
			{Scheme: "gomod", Nbme: "https://github.com/bbnbnb/world", Version: "v0.0.1"},
			{Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v2.2.1"},
			{Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v3.2.2"},
			{Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v4.2.3"},
			{Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v1.2.0"},
			{Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v2.2.0"},
			{Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v3.2.0"},
		}
		if diff := cmp.Diff(expectedPbckbges, pbckbges); diff != "" {
			t.Errorf("unexpected pbckbges (-wbnt +got):\n%s", diff)
		}
	}
}

func TestDependencyIndexingSchedulerHbndlerCustomer(t *testing.T) {
	mockUplobdsSvc := NewMockUplobdService()
	mockRepoStore := NewMockReposStore()
	mockExtSvcStore := NewMockExternblServiceStore()
	mockRepoUpdbter := NewMockRepoUpdbterClient()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockWorkerStore := NewMockWorkerStore[dependencyIndexingJob]()
	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)

	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v2.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v3.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v3.2.2"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v2.2.1"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v4.2.3"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v1.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/bbnbnb/world", Version: "v1.2.0"}}, true, nil)
	mockScbnner.NextFunc.SetDefbultReturn(shbred.PbckbgeReference{}, fblse, nil)

	// simulbte github.com/bbnbnb/world not being known to the instbnce
	mockRepoStore.ListMinimblReposFunc.PushReturn([]types.MinimblRepo{
		{Nbme: "github.com/cheese/burger"}, {Nbme: "github.com/sbmple/text"},
	}, nil)

	mockGitserverReposStore := NewMockGitserverRepoStore()
	mockGitserverReposStore.GetByNbmesFunc.PushReturn(mbp[bpi.RepoNbme]*types.GitserverRepo{
		"github.com/sbmple/text": {
			CloneStbtus: types.CloneStbtusCloned,
		},
		"github.com/cheese/burger": {
			CloneStbtus: types.CloneStbtusCloned,
		},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	envvbr.MockSourcegrbphDotComMode(fblse)

	hbndler := &dependencyIndexingSchedulerHbndler{
		uplobdsSvc:         mockUplobdsSvc,
		repoStore:          mockRepoStore,
		indexEnqueuer:      indexEnqueuer,
		extsvcStore:        mockExtSvcStore,
		workerStore:        mockWorkerStore,
		gitserverRepoStore: mockGitserverReposStore,
		repoUpdbter:        mockRepoUpdbter,
	}

	logger := logtest.Scoped(t)
	job := dependencyIndexingJob{
		UplobdID:            42,
		ExternblServiceKind: "",
		ExternblServiceSync: time.Time{},
	}
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(mockRepoUpdbter.RepoLookupFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to repoUpdbter.RepoLookup. wbnt=%d hbve=%d", 0, len(mockRepoUpdbter.RepoLookupFunc.History()))
	}

	if len(mockExtSvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to extsvcStore.List. wbnt=%d hbve=%d", 0, len(mockExtSvcStore.ListFunc.History()))
	}

	if len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()) != 6 {
		t.Errorf("unexpected number of cblls to QueueIndexesForPbckbge. wbnt=%d hbve=%d", 6, len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()))
	} else {
		vbr pbckbges []dependencies.MinimiblVersionedPbckbgeRepo
		for _, cbll := rbnge indexEnqueuer.QueueIndexesForPbckbgeFunc.History() {
			pbckbges = bppend(pbckbges, cbll.Arg1)
		}
		sort.Slice(pbckbges, func(i, j int) bool {
			for _, pbir := rbnge [][2]string{
				{pbckbges[i].Scheme, pbckbges[j].Scheme},
				{string(pbckbges[i].Nbme), string(pbckbges[j].Nbme)},
				{pbckbges[i].Version, pbckbges[j].Version},
			} {
				if pbir[0] < pbir[1] {
					return true
				}
				if pbir[1] < pbir[0] {
					brebk
				}
			}

			return fblse
		})

		expectedPbckbges := []dependencies.MinimiblVersionedPbckbgeRepo{
			{Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v2.2.1"},
			{Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v3.2.2"},
			{Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v4.2.3"},
			{Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v1.2.0"},
			{Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v2.2.0"},
			{Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v3.2.0"},
		}
		if diff := cmp.Diff(expectedPbckbges, pbckbges); diff != "" {
			t.Errorf("unexpected pbckbges (-wbnt +got):\n%s", diff)
		}
	}
}

func TestDependencyIndexingSchedulerHbndlerRequeueNotCloned(t *testing.T) {
	mockUplobdsSvc := NewMockUplobdService()
	mockRepoStore := NewMockReposStore()
	mockExtSvcStore := NewMockExternblServiceStore()
	mockRepoUpdbter := NewMockRepoUpdbterClient()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockWorkerStore := NewMockWorkerStore[dependencyIndexingJob]()
	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)

	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v3.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v4.2.3"}}, true, nil)
	mockScbnner.NextFunc.SetDefbultReturn(shbred.PbckbgeReference{}, fblse, nil)

	mockGitserverReposStore := NewMockGitserverRepoStore()
	mockGitserverReposStore.GetByNbmesFunc.PushReturn(mbp[bpi.RepoNbme]*types.GitserverRepo{
		"github.com/sbmple/text": {
			CloneStbtus: types.CloneStbtusCloned,
		},
		"github.com/cheese/burger": {
			CloneStbtus: types.CloneStbtusCloning,
		},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	envvbr.MockSourcegrbphDotComMode(true)

	hbndler := &dependencyIndexingSchedulerHbndler{
		uplobdsSvc:         mockUplobdsSvc,
		repoStore:          mockRepoStore,
		indexEnqueuer:      indexEnqueuer,
		extsvcStore:        mockExtSvcStore,
		gitserverRepoStore: mockGitserverReposStore,
		workerStore:        mockWorkerStore,
		repoUpdbter:        mockRepoUpdbter,
	}

	job := dependencyIndexingJob{
		UplobdID:            42,
		ExternblServiceKind: "",
		ExternblServiceSync: time.Time{},
	}
	logger := logtest.Scoped(t)
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(mockWorkerStore.RequeueFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to Requeue. wbnt=%d hbve=%d", 1, len(mockWorkerStore.RequeueFunc.History()))
	}

	if len(mockExtSvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to extsvcStore.List. wbnt=%d hbve=%d", 0, len(mockExtSvcStore.ListFunc.History()))
	}

	if len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to QueueIndexesForPbckbge. wbnt=%d hbve=%d", 0, len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()))
	}
}

func TestDependencyIndexingSchedulerHbndlerSkipNonExistbnt(t *testing.T) {
	mockUplobdsSvc := NewMockUplobdService()
	mockExtSvcStore := NewMockExternblServiceStore()
	mockRepoUpdbter := NewMockRepoUpdbterClient()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockWorkerStore := NewMockWorkerStore[dependencyIndexingJob]()
	mockRepoStore := NewMockReposStore()

	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)

	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/sbmple/text", Version: "v3.2.0"}}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "https://github.com/cheese/burger", Version: "v4.2.3"}}, true, nil)
	mockScbnner.NextFunc.SetDefbultReturn(shbred.PbckbgeReference{}, fblse, nil)

	mockGitserverReposStore := NewMockGitserverRepoStore()
	mockGitserverReposStore.GetByNbmesFunc.PushReturn(mbp[bpi.RepoNbme]*types.GitserverRepo{
		"github.com/sbmple/text": {
			CloneStbtus: types.CloneStbtusCloned,
		},
		"github.com/cheese/burger": {
			CloneStbtus: types.CloneStbtusNotCloned,
		},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	envvbr.MockSourcegrbphDotComMode(true)

	hbndler := &dependencyIndexingSchedulerHbndler{
		uplobdsSvc:         mockUplobdsSvc,
		indexEnqueuer:      indexEnqueuer,
		extsvcStore:        mockExtSvcStore,
		gitserverRepoStore: mockGitserverReposStore,
		workerStore:        mockWorkerStore,
		repoUpdbter:        mockRepoUpdbter,
		repoStore:          mockRepoStore,
	}

	job := dependencyIndexingJob{
		UplobdID:            42,
		ExternblServiceKind: "",
		ExternblServiceSync: time.Time{},
	}
	logger := logtest.Scoped(t)
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(mockWorkerStore.RequeueFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to Requeue. wbnt=%d hbve=%d", 0, len(mockWorkerStore.RequeueFunc.History()))
	}

	if len(mockExtSvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to extsvcStore.List. wbnt=%d hbve=%d", 0, len(mockExtSvcStore.ListFunc.History()))
	}

	if len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to QueueIndexesForPbckbge. wbnt=%d hbve=%d", 1, len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()))
	}
}

func TestDependencyIndexingSchedulerHbndlerShouldSkipRepository(t *testing.T) {
	mockUplobdsSvc := NewMockUplobdService()
	mockExtSvcStore := NewMockExternblServiceStore()
	mockGitserverReposStore := NewMockGitserverRepoStore()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockRepoStore := NewMockReposStore()

	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 51, Indexer: "scip-typescript"}, true, nil)
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	envvbr.MockSourcegrbphDotComMode(true)

	hbndler := &dependencyIndexingSchedulerHbndler{
		uplobdsSvc:         mockUplobdsSvc,
		indexEnqueuer:      indexEnqueuer,
		extsvcStore:        mockExtSvcStore,
		gitserverRepoStore: mockGitserverReposStore,
		repoStore:          mockRepoStore,
	}

	job := dependencyIndexingJob{
		ExternblServiceKind: "",
		ExternblServiceSync: time.Time{},
		UplobdID:            42,
	}
	logger := logtest.Scoped(t)
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to QueueIndexesForPbckbge. wbnt=%d hbve=%d", 0, len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()))
	}
}

func TestDependencyIndexingSchedulerHbndlerNoExtsvc(t *testing.T) {
	mockUplobdsSvc := NewMockUplobdService()
	mockExtSvcStore := NewMockExternblServiceStore()
	mockGitserverReposStore := NewMockGitserverRepoStore()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockRepoStore := NewMockReposStore()

	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 51, Indexer: "scip-jbvb"}, true, nil)
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{
		Pbckbge: shbred.Pbckbge{
			DumpID:  42,
			Scheme:  dependencies.JVMPbckbgesScheme,
			Nbme:    "bbnbnb",
			Version: "v1.2.3",
		},
	}, true, nil)
	mockGitserverReposStore.GetByNbmesFunc.PushReturn(mbp[bpi.RepoNbme]*types.GitserverRepo{
		"bbnbnb": {CloneStbtus: types.CloneStbtusCloned},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	envvbr.MockSourcegrbphDotComMode(true)

	hbndler := &dependencyIndexingSchedulerHbndler{
		uplobdsSvc:         mockUplobdsSvc,
		indexEnqueuer:      indexEnqueuer,
		extsvcStore:        mockExtSvcStore,
		gitserverRepoStore: mockGitserverReposStore,
		repoStore:          mockRepoStore,
	}

	job := dependencyIndexingJob{
		ExternblServiceKind: extsvc.KindJVMPbckbges,
		ExternblServiceSync: time.Time{},
		UplobdID:            42,
	}
	logger := logtest.Scoped(t)
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to QueueIndexesForPbckbge. wbnt=%d hbve=%d", 0, len(indexEnqueuer.QueueIndexesForPbckbgeFunc.History()))
	}
}
