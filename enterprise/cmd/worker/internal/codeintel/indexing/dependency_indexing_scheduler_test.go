package indexing

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestDependencyIndexingSchedulerHandler(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockExtSvcStore := NewMockExternalServiceStore()
	mockGitServer := NewMockGitserverClient()
	mockRepoUpdater := NewMockRepoUpdaterClient()
	mockScanner := NewMockPackageReferenceScanner()
	mockWorkerStore := NewMockWorkerStore()
	mockDBStore.WithFunc.SetDefaultReturn(mockDBStore)
	mockDBStore.GetUploadByIDFunc.SetDefaultReturn(dbstore.Upload{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockDBStore.ReferencesForUploadFunc.SetDefaultReturn(mockScanner, nil)

	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v2.2.0"}}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v3.2.0"}}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v3.2.2"}}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v2.2.1"}}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v4.2.3"}}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v1.2.0"}}, true, nil)
	mockScanner.NextFunc.SetDefaultReturn(shared.PackageReference{}, false, nil)

	mockGitServer.RepoInfoFunc.PushReturn(map[api.RepoName]*protocol.RepoInfo{
		"https://github.com/sample/text": {
			CloneInProgress: false,
			Cloned:          true,
		},
		"https://github.com/cheese/burger": {
			CloneInProgress: false,
			Cloned:          true,
		},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	handler := &dependencyIndexingSchedulerHandler{
		dbStore:       mockDBStore,
		indexEnqueuer: indexEnqueuer,
		extsvcStore:   mockExtSvcStore,
		workerStore:   mockWorkerStore,
		gitserver:     mockGitServer,
		repoUpdater:   mockRepoUpdater,
	}

	logger := logtest.Scoped(t)
	job := dbstore.DependencyIndexingJob{
		UploadID:            42,
		ExternalServiceKind: "",
		ExternalServiceSync: time.Time{},
	}
	if err := handler.Handle(context.Background(), logger, job); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockExtSvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to extsvcStore.List. want=%d have=%d", 0, len(mockExtSvcStore.ListFunc.History()))
	}

	if len(indexEnqueuer.QueueIndexesForPackageFunc.History()) != 6 {
		t.Errorf("unexpected number of calls to QueueIndexesForPackage. want=%d have=%d", 6, len(indexEnqueuer.QueueIndexesForPackageFunc.History()))
	} else {
		var packages []precise.Package
		for _, call := range indexEnqueuer.QueueIndexesForPackageFunc.History() {
			packages = append(packages, call.Arg1)
		}
		sort.Slice(packages, func(i, j int) bool {
			for _, pair := range [][2]string{
				{packages[i].Scheme, packages[j].Scheme},
				{packages[i].Name, packages[j].Name},
				{packages[i].Version, packages[j].Version},
			} {
				if pair[0] < pair[1] {
					return true
				}
				if pair[1] < pair[0] {
					break
				}
			}

			return false
		})

		expectedPackages := []precise.Package{
			{Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v2.2.1"},
			{Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v3.2.2"},
			{Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v4.2.3"},
			{Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v1.2.0"},
			{Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v2.2.0"},
			{Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v3.2.0"},
		}
		if diff := cmp.Diff(expectedPackages, packages); diff != "" {
			t.Errorf("unexpected packages (-want +got):\n%s", diff)
		}
	}
}

func TestDependencyIndexingSchedulerHandlerRequeueNotCloned(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockExtSvcStore := NewMockExternalServiceStore()
	mockGitServer := NewMockGitserverClient()
	mockRepoUpdater := NewMockRepoUpdaterClient()
	mockScanner := NewMockPackageReferenceScanner()
	mockWorkerStore := NewMockWorkerStore()
	mockDBStore.WithFunc.SetDefaultReturn(mockDBStore)
	mockDBStore.GetUploadByIDFunc.SetDefaultReturn(dbstore.Upload{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockDBStore.ReferencesForUploadFunc.SetDefaultReturn(mockScanner, nil)

	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v3.2.0"}}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v4.2.3"}}, true, nil)
	mockScanner.NextFunc.SetDefaultReturn(shared.PackageReference{}, false, nil)

	mockGitServer.RepoInfoFunc.PushReturn(map[api.RepoName]*protocol.RepoInfo{
		"github.com/sample/text": {
			CloneInProgress: false,
			Cloned:          true,
		},
		"github.com/cheese/burger": {
			CloneInProgress: true,
			Cloned:          false,
		},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	handler := &dependencyIndexingSchedulerHandler{
		dbStore:       mockDBStore,
		indexEnqueuer: indexEnqueuer,
		extsvcStore:   mockExtSvcStore,
		workerStore:   mockWorkerStore,
		gitserver:     mockGitServer,
		repoUpdater:   mockRepoUpdater,
	}

	job := dbstore.DependencyIndexingJob{
		UploadID:            42,
		ExternalServiceKind: "",
		ExternalServiceSync: time.Time{},
	}
	logger := logtest.Scoped(t)
	if err := handler.Handle(context.Background(), logger, job); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockWorkerStore.RequeueFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to Requeue. want=%d have=%d", 1, len(mockWorkerStore.RequeueFunc.History()))
	}

	if len(mockExtSvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to extsvcStore.List. want=%d have=%d", 0, len(mockExtSvcStore.ListFunc.History()))
	}

	if len(indexEnqueuer.QueueIndexesForPackageFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to QueueIndexesForPackage. want=%d have=%d", 0, len(indexEnqueuer.QueueIndexesForPackageFunc.History()))
	}
}

func TestDependencyIndexingSchedulerHandlerSkipNonExistant(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockExtSvcStore := NewMockExternalServiceStore()
	mockRepoUpdater := NewMockRepoUpdaterClient()
	mockGitServer := NewMockGitserverClient()
	mockScanner := NewMockPackageReferenceScanner()
	mockWorkerStore := NewMockWorkerStore()
	mockDBStore.WithFunc.SetDefaultReturn(mockDBStore)
	mockDBStore.GetUploadByIDFunc.SetDefaultReturn(dbstore.Upload{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockDBStore.ReferencesForUploadFunc.SetDefaultReturn(mockScanner, nil)

	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/sample/text", Version: "v3.2.0"}}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "https://github.com/cheese/burger", Version: "v4.2.3"}}, true, nil)
	mockScanner.NextFunc.SetDefaultReturn(shared.PackageReference{}, false, nil)

	mockGitServer.RepoInfoFunc.PushReturn(map[api.RepoName]*protocol.RepoInfo{
		"github.com/sample/text": {
			CloneInProgress: false,
			Cloned:          true,
		},
		"github.com/cheese/burger": {
			CloneInProgress: false,
			Cloned:          false,
		},
	}, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	handler := &dependencyIndexingSchedulerHandler{
		dbStore:       mockDBStore,
		indexEnqueuer: indexEnqueuer,
		extsvcStore:   mockExtSvcStore,
		workerStore:   mockWorkerStore,
		gitserver:     mockGitServer,
		repoUpdater:   mockRepoUpdater,
	}

	job := dbstore.DependencyIndexingJob{
		UploadID:            42,
		ExternalServiceKind: "",
		ExternalServiceSync: time.Time{},
	}
	logger := logtest.Scoped(t)
	if err := handler.Handle(context.Background(), logger, job); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockWorkerStore.RequeueFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to Requeue. want=%d have=%d", 0, len(mockWorkerStore.RequeueFunc.History()))
	}

	if len(mockExtSvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to extsvcStore.List. want=%d have=%d", 0, len(mockExtSvcStore.ListFunc.History()))
	}

	if len(indexEnqueuer.QueueIndexesForPackageFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to QueueIndexesForPackage. want=%d have=%d", 1, len(indexEnqueuer.QueueIndexesForPackageFunc.History()))
	}
}

func TestDependencyIndexingSchedulerHandlerShouldSkipRepository(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockExtSvcStore := NewMockExternalServiceStore()
	mockScanner := NewMockPackageReferenceScanner()
	mockGitServer := NewMockGitserverClient()
	mockDBStore.WithFunc.SetDefaultReturn(mockDBStore)
	mockDBStore.GetUploadByIDFunc.SetDefaultReturn(dbstore.Upload{ID: 42, RepositoryID: 51, Indexer: "lsif-tsc"}, true, nil)
	mockDBStore.ReferencesForUploadFunc.SetDefaultReturn(mockScanner, nil)

	indexEnqueuer := NewMockIndexEnqueuer()

	handler := &dependencyIndexingSchedulerHandler{
		dbStore:       mockDBStore,
		indexEnqueuer: indexEnqueuer,
		extsvcStore:   mockExtSvcStore,
		gitserver:     mockGitServer,
	}

	job := dbstore.DependencyIndexingJob{
		ExternalServiceKind: "",
		ExternalServiceSync: time.Time{},
		UploadID:            42,
	}
	logger := logtest.Scoped(t)
	if err := handler.Handle(context.Background(), logger, job); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(indexEnqueuer.QueueIndexesForPackageFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to QueueIndexesForPackage. want=%d have=%d", 0, len(indexEnqueuer.QueueIndexesForPackageFunc.History()))
	}
}
